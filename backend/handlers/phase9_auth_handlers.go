package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/RK-Consulting/skill-sifter/auth"
	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/platformaccess"
	"github.com/RK-Consulting/skill-sifter/models"
	"golang.org/x/crypto/bcrypt"
)

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	var user models.User
	var hashedPassword string
	var tenantUserID int
	var tenantID string
	var platformRole string
	err := db.DB.QueryRow(`
		SELECT user_id, tenant_id, role
		FROM platform_user_accounts
		WHERE LOWER(email)=LOWER($1)
	`, creds.Email).Scan(&tenantUserID, &tenantID, &platformRole)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	access, err := platformaccess.ResolveLoginAccess(db.DB, tenantUserID, tenantID)
	if err != nil {
		access, err = platformaccess.ResolveProvisioningAccess(db.DB, tenantUserID, tenantID)
		if err != nil {
			respondWithError(w, http.StatusForbidden, "Tenant subscription or access is not active")
			return
		}
	}
	tenantDB, err := db.TenantDB(db.DB, tenantID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Tenant database is not ready")
		return
	}
	err = tenantDB.QueryRow(`
		SELECT id, username, email, password, role, tenant_id, created_at
		FROM users WHERE id=$1 AND LOWER(email)=LOWER($2)
	`, tenantUserID, creds.Email).Scan(
		&user.ID, &user.Username, &user.Email, &hashedPassword,
		&user.Role, &user.TenantID, &user.CreatedAt)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	user.Password = ""
	user.CompanyName = ""
	user.Role = platformRole

	tokenString, err := auth.GenerateToken(user, user.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Login successful",
		Data: models.TokenResponse{
			Token:              tokenString,
			User:               user,
			SubscriptionStatus: access.SubscriptionStatus,
			PlanCode:           access.PlanCode,
		},
	})
}

func GetCurrentAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "Authentication context missing")
		return
	}
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	access, err := platformaccess.ResolveLoginAccess(db.DB, userID, tenantID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Tenant subscription or access is not active")
		return
	}

	companyName, _ := r.Context().Value("companyName").(string)

	var userCount, userLimit int
	tenantDB, err := db.TenantDB(db.DB, tenantID)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Tenant database is not ready")
		return
	}
	if err := tenantDB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&userCount); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read tenant users")
		return
	}
	if err := db.DB.QueryRow(`SELECT user_limit FROM platform_subscriptions WHERE tenant_id=$1 AND status IN ('TRIAL','ACTIVE') AND (ends_at IS NULL OR ends_at >= NOW()) ORDER BY starts_at DESC LIMIT 1`, tenantID).Scan(&userLimit); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read subscription user limit")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Account access retrieved successfully",
		Data: map[string]interface{}{
			"userId":             userID,
			"tenantId":           access.TenantID,
			"companyName":        companyName,
			"role":               access.Role,
			"accountStatus":      access.AccountStatus,
			"provisioningStatus": access.ProvisioningStatus,
			"subscriptionStatus": access.SubscriptionStatus,
			"planCode":           access.PlanCode,
			"userCount":          userCount,
			"userLimit":          userLimit,
		},
	})
}

func ProvisionCurrentTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}
	companyName, _ := r.Context().Value("companyName").(string)
	if companyName == "" {
		respondWithError(w, http.StatusBadRequest, "Tenant company name missing")
		return
	}

	recovery, _ := r.Context().Value("provisioningRecovery").(bool)
	var initialUser *db.TenantUser
	var registrationID int64
	if recovery {
		email, _ := r.Context().Value("email").(string)
		if email == "" {
			respondWithError(w, http.StatusUnauthorized, "Provisioning recovery identity missing")
			return
		}
		var username, passwordHash string
		if err := db.DB.QueryRow(`
			SELECT pr.id, pr.username, pr.password_hash, pt.company_name
			FROM platform_registration_registry rr
			JOIN platform_tenants pt ON pt.tenant_id=rr.last_tenant_id
			JOIN platform_pending_registrations pr
			  ON lower(pr.email)=lower(rr.email_id)
			 AND pr.email_verified_at IS NOT NULL
			WHERE lower(rr.email_id)=lower($1)
			  AND rr.last_tenant_id=$2
			ORDER BY pr.id DESC
			LIMIT 1`, email, tenantID).Scan(&registrationID, &username, &passwordHash, &companyName); err != nil {
			respondWithError(w, http.StatusForbidden, "Provisioning recovery registration is no longer available")
			return
		}
		initialUser = &db.TenantUser{
			Username: username,
			Email:    email,
			Password: passwordHash,
			Role:     "admin",
		}
	}

	databaseName, tenantUserID, err := db.ProvisionTenantDatabase(db.DB, tenantID, initialUser)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Tenant database provisioning failed")
		return
	}

	if recovery {
		if _, err = db.DB.Exec(`INSERT INTO platform_user_accounts(tenant_id,user_id,email,role)
			VALUES($1,$2,$3,'admin')
			ON CONFLICT (tenant_id,user_id) DO UPDATE
			SET email=EXCLUDED.email, role=EXCLUDED.role, updated_at=NOW()`,
			tenantID, tenantUserID, initialUser.Email); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not finalize platform account")
			return
		}
		_, _ = db.DB.Exec("DELETE FROM platform_verification_codes WHERE registration_id=$1", registrationID)
		_, _ = db.DB.Exec("DELETE FROM platform_pending_registrations WHERE id=$1", registrationID)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Tenant database is ready",
		Data: map[string]interface{}{
			"tenantId":  tenantID,
			"database":  databaseName,
			"status":    "READY",
			"recovered": recovery,
		},
	})
}
