package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/RK-Consulting/skill-sifter/auth"
	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/platformaccess"
	"github.com/RK-Consulting/skill-sifter/models"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if creds.Email == "" || creds.Password == "" || creds.Username == "" {
		respondWithError(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}
	if creds.CompanyName == "" || creds.PlanCode == "" {
		respondWithError(w, http.StatusBadRequest, "Company name and plan are required")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not start transaction")
		return
	}
	defer tx.Rollback()

	var exists bool
	if err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM platform_tenants WHERE company_name = $1)", creds.CompanyName).Scan(&exists); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		respondWithError(w, http.StatusConflict, "Company already has a SkillSifter account; ask its administrator to create your user")
		return
	}

	companyID := fmt.Sprintf("comp_%s", strings.ReplaceAll(strings.ToLower(creds.CompanyName), " ", "_"))
	if _, err = tx.Exec(
		"INSERT INTO platform_tenants(tenant_id, company_name, account_status, provisioning_status) VALUES($1, $2, 'ACTIVE', 'PENDING') ON CONFLICT (tenant_id) DO NOTHING",
		companyID, creds.CompanyName,
	); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create platform tenant")
		return
	}

	role := "admin"

	if _, err = tx.Exec(`
		INSERT INTO platform_tenants(tenant_id, company_name, account_status, provisioning_status)
		VALUES($1, $2, 'ACTIVE', 'PENDING')
		ON CONFLICT (tenant_id) DO NOTHING
	`, companyID, creds.CompanyName); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create platform tenant")
		return
	}

	var planLimit int
	if err = tx.QueryRow(
		"SELECT user_limit FROM platform_plans WHERE code=$1 AND active=TRUE",
		creds.PlanCode,
	).Scan(&planLimit); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusBadRequest, "Selected subscription plan is not available")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not validate subscription plan")
		return
	}

	if _, err = tx.Exec(`
		INSERT INTO platform_subscriptions(tenant_id, plan_code, status, user_limit)
		VALUES($1, $2, 'TRIAL', $3)
	`, companyID, creds.PlanCode, planLimit); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create platform subscription")
		return
	}

	if err = tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not commit transaction")
		return
	}

	databaseName, userID, err := db.ProvisionTenantDatabase(db.DB, companyID, creds.CompanyName, &db.TenantUser{Username: creds.Username, Email: creds.Email, Password: hashedPassword, Role: role})
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Tenant database provisioning failed; administrator can retry provisioning")
		return
	}

	if _, err = db.DB.Exec(`INSERT INTO platform_user_accounts(tenant_id,user_id,email,role) VALUES($1,$2,$3,$4) ON CONFLICT (tenant_id,user_id) DO UPDATE SET email=EXCLUDED.email, role=EXCLUDED.role, updated_at=NOW()`, companyID, userID, creds.Email, role); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create platform user account")
		return
	}

	user := models.User{
		ID:          userID,
		Username:    creds.Username,
		Email:       creds.Email,
		Role:        role,
		TenantID:    companyID,
		CompanyName: creds.CompanyName,
		CreatedAt:   time.Now(),
	}
	tokenString, err := auth.GenerateToken(user, role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "User registered successfully",
		Data: models.TokenResponse{
			Token:              tokenString,
			User:               user,
			SubscriptionStatus: "TRIAL",
			PlanCode:           creds.PlanCode,
		},
	})
}

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

	user.Password = ""
	access, err := platformaccess.ResolveLoginAccess(db.DB, user.ID, user.TenantID)
	if err != nil {
		access, err = platformaccess.ResolveProvisioningAccess(db.DB, user.ID, user.TenantID)
		if err != nil {
			respondWithError(w, http.StatusForbidden, "Tenant subscription or access is not active")
			return
		}
	}
	user.Role = access.Role
	user.TenantID = access.TenantID

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
	if err := db.DB.QueryRow(`
		SELECT COUNT(*), s.user_limit
		FROM users u
		JOIN platform_subscriptions s ON s.tenant_id = u.tenant_id
		WHERE u.tenant_id = $1
		  AND s.status IN ('TRIAL', 'ACTIVE')
		  AND (s.ends_at IS NULL OR s.ends_at >= NOW())
		GROUP BY s.id, s.user_limit
		ORDER BY s.starts_at DESC
		LIMIT 1
	`, tenantID).Scan(&userCount, &userLimit); err != nil {
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

	databaseName, err := db.ProvisionTenantDatabase(db.DB, tenantID, companyName)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Tenant database provisioning failed")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Tenant database is ready",
		Data: map[string]interface{}{
			"tenantId": tenantID,
			"database": databaseName,
			"status":   "READY",
		},
	})
}
