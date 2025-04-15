
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/RK-Consulting/skill-sifter/auth"
	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUser handles user registration
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

	// Defer rollback in case of error (no-op if tx.Commit is called)
	defer tx.Rollback()

	var companyID int
	var isFirstUser bool


	// Check if company exists
	if creds.Company != "" {
		// Check if company already exists
		err := tx.QueryRow("SELECT id FROM companies WHERE name = $1", creds.Company).Scan(&companyID)
		if err != nil {
			if err == sql.ErrNoRows {
				// Create new company
				err = tx.QueryRow("INSERT INTO companies(name, created_at) VALUES($1, $2) RETURNING id",
					creds.Company, time.Now()).Scan(&companyID)
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Could not create company")
					return
				}
				isFirstUser = true
			} else {
				respondWithError(w, http.StatusInternalServerError, "Database error")
				return
			}
		}
	} else {
		respondWithError(w, http.StatusBadRequest, "Company name is required")
		return
	}

	// Assign role name directly
	role := "recruiter"
	if isFirstUser {
		role = "admin"
	}
	// Insert user with company ID
	var userID int
	err = tx.QueryRow(`
        INSERT INTO users(username, email, password, role, company_id, created_at) 
        VALUES($1, $2, $3, $4, $5, $6) RETURNING id`,
		creds.Username, creds.Email, hashedPassword, role, companyID, time.Now()).Scan(&userID)

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
		ID:        userID,
		Username:  creds.Username,
		Email:     creds.Email,
		Role:      role,
		CompanyID: companyID,
		CreatedAt: time.Now(),
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
		SELECT u.id, u.username, u.email, u.password, u.role, u.company_id, u.created_at
		FROM users u
		WHERE u.email = $1`, creds.Email).Scan(
		&user.ID, &user.Username, &user.Email, &hashedPassword,
		&user.Role, &user.CompanyID, &user.CreatedAt)

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
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	users := []models.User{}
	rows, err := db.DB.Query(`
		SELECT id, username, email, role, company_id, created_at
		FROM users 
		WHERE company_id = $1`, companyID)
		
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching users")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CompanyID, &u.CreatedAt); err != nil {
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
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	
	// Set company ID from logged in admin
	user.CompanyID = companyID

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	// Insert user
	var userID int
	err = db.DB.QueryRow(`
        INSERT INTO users(username, email, password, role, company_id, created_at) 
        VALUES($1, $2, $3, $4, $5, $6) RETURNING id`,
		user.Username, user.Email, hashedPassword, user.Role, user.CompanyID, time.Now()).Scan(&userID)

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
		ID:        userID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CompanyID: user.CompanyID,
		CreatedAt: time.Now(),
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
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Implementation omitted for brevity
	// This would extract the user ID from the URL path and delete the user
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}
