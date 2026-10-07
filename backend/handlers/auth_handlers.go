package handlers

import (
	"database/sql"
	"encoding/json"
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

// LoginUser handles user login
func legacyLoginUser(w http.ResponseWriter, r *http.Request) {
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

// GetUsers fetches all users for a tenant (admin only). Scoped by the
// authenticated tenant_id (ADR 0001), not by company_name.
func GetUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, tenantDB, err := tenantUserDB(r)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Tenant database is not ready")
		return
	}

	users := []models.User{}
	rows, err := tenantDB.Query(`
		SELECT id, username, email, role, tenant_id, created_at
		FROM users
		WHERE tenant_id=$1
		ORDER BY id
	`, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching users")
		return
	}
	defer rows.Close()

	companyName, _ := r.Context().Value("companyName").(string)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.TenantID, &u.CreatedAt); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning user row")
			return
		}
		u.CompanyName = companyName
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
func legacyCreateUser(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)
	companyName := r.Context().Value("companyName").(string)

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Always derive tenant identity from the authenticated context, never
	// from the request body.
	user.TenantID = tenantID
	user.CompanyName = companyName

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	// Insert user
	var userID int
	err = db.DB.QueryRow(`
        INSERT INTO users(username, email, password, role, tenant_id, company_name, created_at) 
        VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		user.Username, user.Email, hashedPassword, user.Role, user.TenantID, user.CompanyName, time.Now()).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	// Create user object for response (without password)
	createdUser := models.User{
		ID:          userID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		TenantID:    user.TenantID,
		CompanyName: user.CompanyName,
		CreatedAt:   time.Now(),
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "User created successfully",
		Data:    createdUser,
	})
}

// UpdateUser updates an existing user (admin only)
// UpdateUser updates a user's username, email, and/or role.
// Reuses the exact same access rule as DeleteUser (docs/architecture.md
// section 13.3), since both are "can this caller touch this user record"
// questions: admin's own record can never be edited via this endpoint,
// manager's record can only be edited by admin, recruiter/team_leader can
// be edited by admin or manager.
// Additionally guards against privilege escalation: only an admin may set
// a target user's role to "admin" (a manager could otherwise create a
// second admin, or edit their own downstream reports upward, bypassing the
// hierarchy this endpoint is meant to protect).
func legacyUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars["id"]

	tenantID := r.Context().Value("tenantID").(string)
	requesterRole := r.Context().Value("role").(string)

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
		respondWithError(w, http.StatusForbidden, "Admin users cannot be edited via this endpoint")
		return
	}
	if currentRole == "manager" && requesterRole != "admin" {
		respondWithError(w, http.StatusForbidden, "Only an admin can edit a manager")
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

	if update.Role == "admin" && requesterRole != "admin" {
		respondWithError(w, http.StatusForbidden, "Only an admin can promote a user to admin")
		return
	}

	validRoles := map[string]bool{"admin": true, "manager": true, "recruiter": true, "team_leader": true}
	if update.Role != "" && !validRoles[update.Role] {
		respondWithError(w, http.StatusBadRequest, "Invalid role")
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

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

// DeleteUser deletes a user, enforcing the role hierarchy defined in
// docs/architecture.md section 13.3:
//   - Admin can never be deleted, by anyone, under any circumstance.
//   - Manager can only be deleted by Admin.
//   - Recruiter/Team Leader can be deleted by Admin or Manager.
//
// This check is based on the TARGET user's actual role, so it is safe
// regardless of whether it's reached via the admin-only or manager-accessible
// route (both are wired to this same handler).
func legacyDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars["id"]

	tenantID := r.Context().Value("tenantID").(string)
	requesterRole := r.Context().Value("role").(string)

	// Look up the target user's role, scoped to the same company (tenant safety —
	// never allow deleting a user from a different company via this endpoint).
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
		respondWithError(w, http.StatusForbidden, "Admin users cannot be deleted")
		return
	}
	if targetRole == "manager" && requesterRole != "admin" {
		respondWithError(w, http.StatusForbidden, "Only an admin can delete a manager")
		return
	}
	// Any remaining case (target is recruiter/team_leader, requester is admin or
	// manager) is allowed, matching the hierarchy table.

	result, err := db.DB.Exec(
		`DELETE FROM users WHERE id = $1 AND tenant_id = $2`,
		targetID, tenantID,
	)
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
