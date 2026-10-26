package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type retiredUser struct {
	userID   sql.NullInt64
	username string
	email    string
}

func retirementStatus(accountStatus, latestSubscriptionStatus string, trialExpiresAt sql.NullTime, now time.Time) string {
	switch {
	case strings.EqualFold(latestSubscriptionStatus, "CANCELLED"):
		return "DISCONTINUED"
	case strings.EqualFold(latestSubscriptionStatus, "EXPIRED"),
		strings.EqualFold(accountStatus, "EXPIRED"),
		trialExpiresAt.Valid && !trialExpiresAt.Time.After(now):
		return "TRIAL_EXPIRED"
	case strings.EqualFold(accountStatus, "TERMINATED"):
		return "ACCOUNT_DELETED"
	default:
		return "ACCOUNT_DELETED"
	}
}

// archiveTenantUsers writes a durable identity/status record before the tenant
// database and its control-plane routing records are removed. Password hashes
// and other credentials are intentionally never copied to the retirement ledger.
func archiveTenantUsers(controlDB *sql.DB, tenantID, databaseName, status string, tenantDatabaseExists bool) error {
	usersByEmail := make(map[string]retiredUser)

	addUser := func(user retiredUser) {
		user.email = strings.TrimSpace(user.email)
		user.username = strings.TrimSpace(user.username)
		if user.email == "" {
			return
		}
		if user.username == "" {
			user.username = user.email
		}
		key := strings.ToLower(user.email)
		existing, found := usersByEmail[key]
		if !found {
			usersByEmail[key] = user
			return
		}
		// Merge fields independently so an exact username from a pending
		// registration cannot erase an ID already known by the control plane.
		if (existing.username == "" || existing.username == existing.email) && user.username != user.email {
			existing.username = user.username
		}
		if !existing.userID.Valid && user.userID.Valid {
			existing.userID = user.userID
		}
		if existing.email == "" {
			existing.email = user.email
		}
		usersByEmail[key] = existing
	}

	accountRows, err := controlDB.Query(
		`SELECT user_id, email FROM platform_user_accounts WHERE tenant_id=$1`, tenantID)
	if err != nil {
		return fmt.Errorf("read platform accounts for retirement: %w", err)
	}
	for accountRows.Next() {
		var userID int64
		var email string
		if err := accountRows.Scan(&userID, &email); err != nil {
			accountRows.Close()
			return fmt.Errorf("scan platform account for retirement: %w", err)
		}
		addUser(retiredUser{userID: sql.NullInt64{Int64: userID, Valid: true}, username: email, email: email})
	}
	if err := accountRows.Err(); err != nil {
		accountRows.Close()
		return fmt.Errorf("read platform account rows for retirement: %w", err)
	}
	accountRows.Close()

	pendingRows, err := controlDB.Query(
		`SELECT pr.username, pr.email
		 FROM platform_pending_registrations pr
		 JOIN platform_registration_registry rr ON lower(rr.email_id)=lower(pr.email)
		 WHERE rr.last_tenant_id=$1`, tenantID)
	if err != nil {
		return fmt.Errorf("read pending registrations for retirement: %w", err)
	}
	for pendingRows.Next() {
		var username, email string
		if err := pendingRows.Scan(&username, &email); err != nil {
			pendingRows.Close()
			return fmt.Errorf("scan pending registration for retirement: %w", err)
		}
		addUser(retiredUser{username: username, email: email})
	}
	if err := pendingRows.Err(); err != nil {
		pendingRows.Close()
		return fmt.Errorf("read pending registration rows for retirement: %w", err)
	}
	pendingRows.Close()

	registryRows, err := controlDB.Query(
		`SELECT email_id FROM platform_registration_registry WHERE last_tenant_id=$1`, tenantID)
	if err != nil {
		return fmt.Errorf("read registration identity for retirement: %w", err)
	}
	for registryRows.Next() {
		var email string
		if err := registryRows.Scan(&email); err != nil {
			registryRows.Close()
			return fmt.Errorf("scan registration identity for retirement: %w", err)
		}
		addUser(retiredUser{username: email, email: email})
	}
	if err := registryRows.Err(); err != nil {
		registryRows.Close()
		return fmt.Errorf("read registration identity rows for retirement: %w", err)
	}
	registryRows.Close()

	if tenantDatabaseExists {
		tenantDB, err := OpenDatabase(databaseName)
		if err != nil {
			return fmt.Errorf("open tenant database before retirement: %w", err)
		}
		tenantRows, queryErr := tenantDB.Query(`SELECT id, username, email FROM users`)
		if queryErr != nil {
			tenantDB.Close()
			return fmt.Errorf("read tenant users before retirement: %w", queryErr)
		}
		for tenantRows.Next() {
			var id int64
			var username, email string
			if err := tenantRows.Scan(&id, &username, &email); err != nil {
				tenantRows.Close()
				tenantDB.Close()
				return fmt.Errorf("scan tenant user before retirement: %w", err)
			}
			addUser(retiredUser{userID: sql.NullInt64{Int64: id, Valid: true}, username: username, email: email})
		}
		if err := tenantRows.Err(); err != nil {
			tenantRows.Close()
			tenantDB.Close()
			return fmt.Errorf("read tenant user rows before retirement: %w", err)
		}
		tenantRows.Close()
		if err := tenantDB.Close(); err != nil {
			return fmt.Errorf("close tenant database after retirement scan: %w", err)
		}
	}

	if len(usersByEmail) == 0 {
		return nil
	}

	tx, err := controlDB.Begin()
	if err != nil {
		return fmt.Errorf("begin retired-user ledger transaction: %w", err)
	}
	defer tx.Rollback()

	for _, user := range usersByEmail {
		var userID interface{}
		if user.userID.Valid {
			userID = user.userID.Int64
		}
		_, err := tx.Exec(
			`INSERT INTO platform_retired_users
			 (tenant_id, user_id, username, email, status, retired_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			 ON CONFLICT (tenant_id, email) DO UPDATE
			 SET user_id=COALESCE(EXCLUDED.user_id, platform_retired_users.user_id),
			     username=CASE
			       WHEN EXCLUDED.username=EXCLUDED.email THEN platform_retired_users.username
			       ELSE EXCLUDED.username
			     END,
			     status=EXCLUDED.status,
			     retired_at=NOW(),
			     updated_at=NOW()`,
			tenantID, userID, user.username, user.email, status)
		if err != nil {
			return fmt.Errorf("record retired user %q: %w", user.email, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit retired-user ledger: %w", err)
	}
	return nil
}
