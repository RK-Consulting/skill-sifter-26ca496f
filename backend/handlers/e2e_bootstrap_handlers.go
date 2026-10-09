package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/RK-Consulting/skill-sifter/db"
	"golang.org/x/crypto/bcrypt"
)

const (
	e2eSmokeTenantID = "e2e_smoke_tenant"
	e2eSmokeCompany  = "SkillSifter E2E Smoke"
	e2eSmokeEmail    = "e2e-admin@skillsifter.in"
	e2eSmokePlan     = "starter_monthly"
)

// BootstrapE2ESmokeAccount creates or repairs only the dedicated production
// smoke tenant. It is deliberately outside the authoritative schema baseline.
// The password is supplied by the CI secret and is stored only as a bcrypt hash
// in the tenant database. Calling this endpoint repeatedly is safe.
func BootstrapE2ESmokeAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Email != e2eSmokeEmail {
		respondWithError(w, http.StatusBadRequest, "Only the dedicated E2E smoke account can be bootstrapped")
		return
	}
	if len(input.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "E2E smoke password must be at least 6 characters")
		return
	}

	planLimit, err := e2eSmokePlanLimit()
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "E2E smoke plan is not available")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not secure E2E smoke password")
		return
	}

	if err := ensureE2EControlState(planLimit); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not bootstrap E2E smoke account")
		return
	}

	_, userID, err := db.ProvisionTenantDatabase(db.DB, e2eSmokeTenantID, &db.TenantUser{
		Username: "E2E Administrator",
		Email:    e2eSmokeEmail,
		Password: string(passwordHash),
		Role:     "admin",
	})
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Could not provision E2E smoke tenant")
		return
	}

	if _, err = db.DB.Exec(`
		INSERT INTO platform_user_accounts(tenant_id,user_id,email,role)
		VALUES($1,$2,$3,'admin')
		ON CONFLICT (tenant_id,user_id) DO UPDATE
		SET email=EXCLUDED.email, role=EXCLUDED.role, updated_at=NOW()
	`, e2eSmokeTenantID, userID, e2eSmokeEmail); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not finalize E2E platform account")
		return
	}

	_, _ = db.DB.Exec("DELETE FROM platform_verification_codes WHERE platform_account_id IN (SELECT id FROM platform_user_accounts WHERE tenant_id=$1)", e2eSmokeTenantID)
	_, _ = db.DB.Exec("DELETE FROM platform_pending_registrations WHERE lower(email)=lower($1)", e2eSmokeEmail)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "E2E smoke account is ready",
		"data": map[string]interface{}{
			"tenantId": e2eSmokeTenantID,
			"email":    e2eSmokeEmail,
			"status":   "READY",
		},
	})
}

// ResetE2ESmokeTenantData removes all tenant business data from the dedicated
// production smoke tenant while preserving its roles and administrator account.
// It is authenticated with the same CI-only E2E administrator credentials used
// by the production smoke workflow.
func ResetE2ESmokeTenantData(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Email != e2eSmokeEmail || len(input.Password) < 6 {
		respondWithError(w, http.StatusUnauthorized, "E2E smoke credentials are invalid")
		return
	}

	tenantDB, err := db.OpenDatabase(db.TenantDatabaseName(e2eSmokeTenantID))
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Could not open E2E smoke tenant database")
		return
	}
	defer tenantDB.Close()

	var passwordHash string
	if err := tenantDB.QueryRow(`
		SELECT password
		FROM users
		WHERE lower(email)=lower($1) AND tenant_id=$2 AND role='admin'
		LIMIT 1
	`, input.Email, e2eSmokeTenantID).Scan(&passwordHash); err != nil {
		respondWithError(w, http.StatusUnauthorized, "E2E smoke credentials are invalid")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "E2E smoke credentials are invalid")
		return
	}

	if _, err := tenantDB.Exec(`
		TRUNCATE TABLE
			clients,
			requirements,
			candidates,
			candidate_language_expertise,
			candidate_expertise,
			resumes,
			resume_search_logs,
			candidate_professional_profiles,
			candidate_employment_history,
			candidate_education,
			candidate_certifications,
			candidate_projects,
			interviews,
			recruitment_screenings,
			recruitment_submissions,
			recruitment_submission_feedback,
			recruitment_selections,
			recruitment_offers,
			recruitment_joinings,
			recruitment_billings,
			audit_events
		RESTART IDENTITY CASCADE
	`); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not reset E2E smoke tenant data")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "E2E smoke tenant data reset",
		"data": map[string]interface{}{
			"tenantId": e2eSmokeTenantID,
			"status":   "RESET",
		},
	})
}

func e2eSmokePlanLimit() (int, error) {
	var limit int
	err := db.DB.QueryRow(
		"SELECT user_limit FROM platform_plans WHERE code=$1 AND active=TRUE",
		e2eSmokePlan,
	).Scan(&limit)
	return limit, err
}

func ensureE2EControlState(planLimit int) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var company string
	err = tx.QueryRow(
		"SELECT company_name FROM platform_tenants WHERE tenant_id=$1",
		e2eSmokeTenantID,
	).Scan(&company)
	if err == sql.ErrNoRows {
		var conflictingTenant string
		err = tx.QueryRow(
			"SELECT tenant_id FROM platform_tenants WHERE company_name=$1",
			e2eSmokeCompany,
		).Scan(&conflictingTenant)
		if err != sql.ErrNoRows {
			if err == nil {
				return &e2eBootstrapConflictError{tenantID: conflictingTenant}
			}
			return err
		}

		if _, err = tx.Exec(`
			INSERT INTO platform_tenants(
				tenant_id, company_name, account_status, provisioning_status,
				trial_started_at, trial_expires_at, data_deletion_at
			) VALUES($1,$2,'ACTIVE','PENDING',NOW(),NULL,NULL)
		`, e2eSmokeTenantID, e2eSmokeCompany); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if company != e2eSmokeCompany {
		return &e2eBootstrapConflictError{tenantID: e2eSmokeTenantID}
	}

	if _, err = tx.Exec(`
		UPDATE platform_tenants
		SET company_name=$2, account_status='ACTIVE', updated_at=NOW()
		WHERE tenant_id=$1
	`, e2eSmokeTenantID, e2eSmokeCompany); err != nil {
		return err
	}

	var subscriptionID int64
	err = tx.QueryRow(
		"SELECT id FROM platform_subscriptions WHERE tenant_id=$1 ORDER BY starts_at DESC, id DESC LIMIT 1",
		e2eSmokeTenantID,
	).Scan(&subscriptionID)
	if err == sql.ErrNoRows {
		_, err = tx.Exec(`
			INSERT INTO platform_subscriptions(
				tenant_id, plan_code, status, starts_at, ends_at,
				provider, provider_subscription_ref, user_limit
			) VALUES($1,$2,'ACTIVE',NOW(),NULL,'e2e','e2e-smoke',$3)
		`, e2eSmokeTenantID, e2eSmokePlan, planLimit)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		if _, err = tx.Exec(`
			UPDATE platform_subscriptions
			SET plan_code=$2, status='ACTIVE', starts_at=COALESCE(starts_at,NOW()),
				ends_at=NULL, provider='e2e',
				provider_subscription_ref='e2e-smoke', user_limit=$3, updated_at=NOW()
			WHERE id=$1
		`, subscriptionID, e2eSmokePlan, planLimit); err != nil {
			return err
		}
		if _, err = tx.Exec(
			"DELETE FROM platform_subscriptions WHERE tenant_id=$1 AND id<>$2",
			e2eSmokeTenantID, subscriptionID,
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec(`
		INSERT INTO platform_registration_registry(email_id,first_registered,last_tenant_id)
		VALUES($1,NOW(),$2)
		ON CONFLICT(email_id) DO UPDATE
		SET last_tenant_id=EXCLUDED.last_tenant_id
	`, e2eSmokeEmail, e2eSmokeTenantID); err != nil {
		return err
	}

	_, err = tx.Exec(
		"DELETE FROM platform_pending_registrations WHERE lower(email)=lower($1)",
		e2eSmokeEmail,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

type e2eBootstrapConflictError struct {
	tenantID string
}

func (e *e2eBootstrapConflictError) Error() string {
	return "E2E smoke tenant identity conflict: " + e.tenantID
}
