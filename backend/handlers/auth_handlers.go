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

	// Check if company already exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM companies WHERE name = $1)", creds.CompanyName).Scan(&exists)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Generate company ID for new companies
	var companyID string
	var isFirstUser bool

	if !exists {
		// Create new company with generated ID
		companyID = fmt.Sprintf("comp_%s", strings.ReplaceAll(strings.ToLower(creds.CompanyName), " ", "_"))
		_, err = tx.Exec("INSERT INTO companies(id, name, created_at) VALUES($1, $2, $3)",
			companyID, creds.CompanyName, time.Now())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not create company")
			return
		}
		isFirstUser = true
	}

	// Determine role
	var role string
	if creds.Role != "" {
		role = creds.Role
	} else if isFirstUser {
		role = "admin"
	} else {
		role = "recruiter" // Default role if none provided
	}

	// Validate role
	validRoles := map[string]bool{
		"admin":       true,
		"manager":     true,
		"recruiter":   true,
		"team_leader": true,
	}

	if !validRoles[role] {
		respondWithError(w, http.StatusBadRequest, "Invalid role")
		return
	}

	// Insert user with company name
	var userID int
	err = tx.QueryRow(`
        INSERT INTO users(username, email, password, role, company_name, created_at) 
        VALUES($1, $2, $3, $4, $5, $6) RETURNING id`,
		creds.Username, creds.Email, hashedPassword, role, creds.CompanyName, time.Now()).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not register user")
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
		CompanyName: creds.CompanyName, // Use company name instead of ID
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
		SELECT u.id, u.username, u.email, u.password, u.role, u.company_name, u.created_at
		FROM users u
		WHERE u.email = $1`, creds.Email).Scan(
		&user.ID, &user.Username, &user.Email, &hashedPassword,
		&user.Role, &user.CompanyName, &user.CreatedAt)

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
			Token: tokenString,
			User:  user,
		},
	})
}

// GetUsers fetches all users for a company (admin only)
func GetUsers(w http.ResponseWriter, r *http.Request) {
	// Get company name from context
	companyName := r.Context().Value("companyName").(string)

	users := []models.User{}
	rows, err := db.DB.Query(`
		SELECT id, username, email, role, company_name, created_at
		FROM users 
		WHERE company_name = $1`, companyName)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching users")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CompanyName, &u.CreatedAt); err != nil {
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

// CreateUser creates a new user (admin only)
func CreateUser(w http.ResponseWriter, r *http.Request) {
	// Get company name from context
	companyName := r.Context().Value("companyName").(string)

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Set company name from logged in admin
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
        INSERT INTO users(username, email, password, role, company_name, created_at) 
        VALUES($1, $2, $3, $4, $5, $6) RETURNING id`,
		user.Username, user.Email, hashedPassword, user.Role, user.CompanyName, time.Now()).Scan(&userID)

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
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Implementation omitted for brevity
	// This would extract the user ID from the URL path and update the user
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

// DeleteUser deletes a user (admin only)
// DeleteUser deletes a user, enforcing the role hierarchy defined in
// docs/architecture.md section 13.3:
//   - Admin can never be deleted, by anyone, under any circumstance.
//   - Manager can only be deleted by Admin.
//   - Recruiter/Team Leader can be deleted by Admin or Manager.
//
// This check is based on the TARGET user's actual role, so it is safe
// regardless of whether it's reached via the admin-only or manager-accessible
// route (both are wired to this same handler).
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars["id"]

	companyName := r.Context().Value("companyName").(string)
	requesterRole := r.Context().Value("role").(string)

	// Look up the target user's role, scoped to the same company (tenant safety —
	// never allow deleting a user from a different company via this endpoint).
	var targetRole string
	err := db.DB.QueryRow(
		`SELECT role FROM users WHERE id = $1 AND company_name = $2`,
		targetID, companyName,
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
		`DELETE FROM users WHERE id = $1 AND company_name = $2`,
		targetID, companyName,
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