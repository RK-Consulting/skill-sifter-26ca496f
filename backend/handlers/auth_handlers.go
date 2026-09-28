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
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate required fields
	if creds.Email == "" || creds.Password == "" || creds.Username == "" {
		respondWithError(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	// Start a transaction
	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not start transaction")
		return
	}

	defer tx.Rollback()

	// Check if company name is provided
	if creds.CompanyName == "" {
		respondWithError(w, http.StatusBadRequest, "Company name is required")
		return
	}

	// Registration creates a new SkillSifter tenant. Existing tenants must
	// create additional users through their tenant admin rather than public
	// registration, so a public caller cannot join another tenant or choose
	// an elevated role.
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM companies WHERE name = $1)", creds.CompanyName).Scan(&exists)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		respondWithError(w, http.StatusConflict, "Company already has a SkillSifter account; ask its administrator to create your user")
		return
	}

	// Resolve the authoritative tenant identity (companies.id, per ADR 0001)
	// for this registration. This used to only be computed for brand-new
	// companies — an existing company's id was never looked up, so a
	// second user joining an existing company had no way to be linked to
	// its tenant identity. Both branches now resolve companyID explicitly.
	// Create the new tenant with an immutable platform identity.
	companyID := fmt.Sprintf("comp_%s", strings.ReplaceAll(strings.ToLower(creds.CompanyName), " ", "_"))
	_, err = tx.Exec("INSERT INTO companies(id, name, created_at) VALUES($1, $2, $3)",
		companyID, creds.CompanyName, time.Now())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create company")
		return
	}

	// The first account created through public registration is always the
	// tenant administrator. Role assignment for existing tenants belongs to
	// the authenticated tenant-admin workflow.
	role := "admin"

	// Insert user with tenant_id (authoritative) and company_name
	// (display/compatibility). tenant_id is always server-resolved above,
	// never taken from the request payload.
	var userID int
	err = tx.QueryRow(`
        INSERT INTO users(username, email, password, role, tenant_id, company_name, created_at) 
        VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		creds.Username, creds.Email, hashedPassword, role, companyID, creds.CompanyName, time.Now()).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not register user")
		return
	}

	// Bridge the newly registered account into the platform/control-plane layer.
	// "legacy" is a compatibility subscription until the real external
	// subscription checkout flow is introduced in Phase 9E.
	_, err = tx.Exec(`
        INSERT INTO platform_tenants(tenant_id, company_name, account_status, provisioning_status)
        VALUES($1, $2, 'ACTIVE', 'READY')
        ON CONFLICT (tenant_id) DO NOTHING`, companyID, creds.CompanyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create platform tenant")
		return
	}

	_, err = tx.Exec(`
        INSERT INTO platform_subscriptions(tenant_id, plan_code, status)
        SELECT $1, 'legacy', 'ACTIVE'
        WHERE NOT EXISTS (
            SELECT 1 FROM platform_subscriptions
            WHERE tenant_id = $1
              AND status IN ('TRIAL', 'ACTIVE')
        )`, companyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create platform subscription")
		return
	}

	_, err = tx.Exec(`
        INSERT INTO platform_user_accounts(user_id, tenant_id, email, role)
        VALUES($1, $2, $3, $4)
        ON CONFLICT (user_id) DO UPDATE
        SET tenant_id = EXCLUDED.tenant_id, email = EXCLUDED.email, role = EXCLUDED.role, updated_at = NOW()`,
		userID, companyID, creds.Email, role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create platform user account")
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not commit transaction")
		return
	}

	// Create user object for response (without password)
	user := models.User{
		ID:          userID,
		Username:    creds.Username,
		Email:       creds.Email,
		Role:        role,
		TenantID:    companyID,
		CompanyName: creds.CompanyName,
		CreatedAt:   time.Now(),
	}

	// Create JWT token
	tokenString, err := auth.GenerateToken(user, role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	// Return token and user info
	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "User registered successfully",
		Data: models.TokenResponse{
			Token: tokenString,
			User:  user,
		},
	})
}

// LoginUser handles user login
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Get user from database
	var user models.User
	var hashedPassword string

	err = db.DB.QueryRow(`
		SELECT u.id, u.username, u.email, u.password, u.role, u.tenant_id, u.company_name, u.created_at
		FROM users u
		WHERE u.email = $1`, creds.Email).Scan(
		&user.ID, &user.Username, &user.Email, &hashedPassword,
		&user.Role, &user.TenantID, &user.CompanyName, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Compare provided password with stored hash
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	user.Password = "" // Don't return the password

	// Resolve tenant, subscription and RBAC from the trusted platform layer.
	// The client cannot supply or override any of these values.
	access, err := platformaccess.ResolveLoginAccess(db.DB, user.ID, user.TenantID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Tenant subscription or access is not active")
		return
	}
	user.Role = access.Role
	user.TenantID = access.TenantID

	// Create JWT token
	tokenString, err := auth.GenerateToken(user, user.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	// Return token and user info
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

// GetCurrentAccount returns the trusted platform access context for the authenticated user.
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
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Account access retrieved successfully",
		Data: map[string]interface{}{
			"userId": userID,
			"tenantId": access.TenantID,
			"companyName": companyName,
			"role": access.Role,
			"accountStatus": access.AccountStatus,
			"subscriptionStatus": access.SubscriptionStatus,
			"planCode": access.PlanCode,
		},
	})
}

// GetUsers fetches all users for a tenant (admin only). Scoped by the
// authenticated tenant_id (ADR 0001), not by company_name.
func GetUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)

	users := []models.User{}
	rows, err := db.DB.Query(`
		SELECT id, username, email, role, tenant_id, company_name, created_at
		FROM users 
		WHERE tenant_id = $1`, tenantID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching users")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.TenantID, &u.CompanyName, &u.CreatedAt); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning user row")
			return
		}
		users = append(users, u)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

// CreateUser creates a new user (admin only). The new user's tenant is
// always the authenticated admin's tenant — tenant_id and company_name in
// the request body (if any) are ignored and overwritten, so a client can
// never place a new user into a different tenant (ADR 0001: "client-provided
// tenant identifiers or names cannot override the authenticated tenant").
func CreateUser(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)
	companyName := r.Context().Value("companyName").(string)

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if input.Username == "" || input.Email == "" || input.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	validRoles := map[string]bool{"manager": true, "recruiter": true, "team_leader": true}
	if !validRoles[input.Role] {
		respondWithError(w, http.StatusBadRequest, "Additional users must have role manager, recruiter, or team_leader")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not start transaction")
		return
	}
	defer tx.Rollback()

	var userLimit, userCount int
	err = tx.QueryRow(`
		SELECT s.user_limit, COUNT(u.id)
		FROM platform_subscriptions s
		LEFT JOIN users u ON u.tenant_id = s.tenant_id
		WHERE s.tenant_id = $1
		  AND s.status IN ('TRIAL', 'ACTIVE')
		  AND (s.ends_at IS NULL OR s.ends_at >= NOW())
		GROUP BY s.id, s.user_limit
		ORDER BY s.starts_at DESC
		LIMIT 1
	`, tenantID).Scan(&userLimit, &userCount)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusForbidden, "No active subscription for this tenant")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read subscription user limit")
		return
	}
	if userCount >= userLimit {
		respondWithError(w, http.StatusConflict, "Subscription user limit reached")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	var userID int
	err = tx.QueryRow(`
		INSERT INTO users(username, email, password, role, tenant_id, company_name, created_at)
		VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		input.Username, input.Email, hashedPassword, input.Role, tenantID, companyName, time.Now()).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	_, err = tx.Exec(`
		INSERT INTO platform_user_accounts(user_id, tenant_id, email, role)
		VALUES($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET tenant_id = EXCLUDED.tenant_id, email = EXCLUDED.email, role = EXCLUDED.role, updated_at = NOW()`,
		userID, tenantID, input.Email, input.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create platform user account")
		return
	}

	if err = tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not commit transaction")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "User created successfully",
		Data: models.User{
			ID: userID, Username: input.Username, Email: input.Email,
			Role: input.Role, TenantID: tenantID, CompanyName: companyName,
			CreatedAt: time.Now(),
		},
	})
}

// UpdateUser updates a non-admin user's profile and fixed V1 role.
// Only the tenant admin may change user roles or user profile data.
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars["id"]
	tenantID := r.Context().Value("tenantID").(string)

	var currentRole string
	err := db.DB.QueryRow(
		`SELECT role FROM users WHERE id = $1 AND tenant_id = $2`,
		targetID, tenantID,
	).Scan(&currentRole)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error looking up user")
		return
	}
	if currentRole == "admin" {
		respondWithError(w, http.StatusForbidden, "The tenant admin cannot be edited through this endpoint")
		return
	}

	var update struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	validRoles := map[string]bool{"manager": true, "recruiter": true, "team_leader": true}
	if update.Role != "" && !validRoles[update.Role] {
		respondWithError(w, http.StatusBadRequest, "User role must be manager, recruiter, or team_leader")
		return
	}

	result, err := db.DB.Exec(
		`UPDATE users SET
			username = COALESCE(NULLIF($1, ''), username),
			email = COALESCE(NULLIF($2, ''), email),
			role = COALESCE(NULLIF($3, ''), role)
		WHERE id = $4 AND tenant_id = $5`,
		update.Username, update.Email, update.Role, targetID, tenantID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user")
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	_, err = db.DB.Exec(
		`UPDATE platform_user_accounts
		SET email = COALESCE(NULLIF($1, ''), email),
		    role = COALESCE(NULLIF($2, ''), role),
		    updated_at = NOW()
		WHERE user_id = $3 AND tenant_id = $4`,
		update.Email, update.Role, targetID, tenantID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating platform account")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

// DeleteUser removes a non-admin tenant user.
// Admin is the only role allowed to manage tenant users in V1.
// The admin account itself can never be deleted through this endpoint.
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars["id"]
	tenantID := r.Context().Value("tenantID").(string)

	var targetRole string
	err := db.DB.QueryRow(
		`SELECT role FROM users WHERE id = $1 AND tenant_id = $2`,
		targetID, tenantID,
	).Scan(&targetRole)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error looking up user")
		return
	}
	if targetRole == "admin" {
		respondWithError(w, http.StatusForbidden, "The tenant admin cannot be deleted")
		return
	}

	result, err := db.DB.Exec(`DELETE FROM users WHERE id = $1 AND tenant_id = $2`, targetID, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting user")
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}
