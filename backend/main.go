
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

// Database connection
var db *sql.DB

// JWT secret key
var jwtKey = []byte("skill_sifter_secret_key") // In production, use env variable

// Model definitions
type Candidate struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Position     string    `json:"position"`
	Status       string    `json:"status"`
	DateApplied  time.Time `json:"dateApplied"`
	ResumeURL    string    `json:"resumeUrl,omitempty"`
	CoverLetter  string    `json:"coverLetter,omitempty"`
	LastModified time.Time `json:"lastModified"`
	CompanyID    int       `json:"companyId"`
}

type Job struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Department   string    `json:"department"`
	Location     string    `json:"location"`
	Status       string    `json:"status"`
	DatePosted   time.Time `json:"datePosted"`
	Description  string    `json:"description,omitempty"`
	Requirements string    `json:"requirements,omitempty"`
	LastModified time.Time `json:"lastModified"`
	CompanyID    int       `json:"companyId"`
}

type DailyJob struct {
	ID           int       `json:"id"`
	JdNo         int       `json:"jdNo"`
	Instructions string    `json:"instructions"`
	AssignedUser int       `json:"assignedUser"`
	AssignedDate time.Time `json:"assignedDate"`
	LastModified time.Time `json:"lastModified"`
	CompanyID    int       `json:"companyId"`
}

type Interview struct {
	ID            int       `json:"id"`
	CandidateID   int       `json:"candidateId"`
	CandidateName string    `json:"candidateName"`
	Position      string    `json:"position"`
	InterviewDate time.Time `json:"interviewDate"`
	Status        string    `json:"status"`
	Feedback      string    `json:"feedback"`
	LastModified  time.Time `json:"lastModified"`
	CompanyID     int       `json:"companyId"`
}

// New Company/Tenant model
type Company struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// Enhanced User model with role
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	RoleID    int       `json:"roleId"`
	Role      string    `json:"role,omitempty"` // Role name for frontend
	CompanyID int       `json:"companyId"`
	CreatedAt time.Time `json:"createdAt"`
}

// Role model
type Role struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// Credentials for login/register
type Credentials struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Username  string `json:"username,omitempty"`
	CompanyID int    `json:"companyId,omitempty"`
	Company   string `json:"company,omitempty"` // For registration
}

// Claims for JWT
type Claims struct {
	UserID    int    `json:"userId"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CompanyID int    `json:"companyId"`
	jwt.RegisteredClaims
}

// ApiResponse represents a standard API response
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// TokenResponse for authentication
type TokenResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

func main() {
	// Initialize database connection
	initDB()
	defer db.Close()

	// Create router
	r := mux.NewRouter()

	// Authentication routes
	r.HandleFunc("/api/auth/register", registerUser).Methods("POST")
	r.HandleFunc("/api/auth/login", loginUser).Methods("POST")

	// Protected routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware)

	// Admin-only routes
	adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(roleMiddleware("admin"))
	adminRouter.HandleFunc("/users", getUsers).Methods("GET")
	adminRouter.HandleFunc("/users", createUser).Methods("POST")
	adminRouter.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	adminRouter.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")
	
	// Manager and Admin routes
	managerRouter := apiRouter.PathPrefix("/manager").Subrouter()
	managerRouter.Use(roleMiddleware("manager", "admin"))
	// Add manager-specific routes here if needed

	// API routes (protected but accessible by all roles)
	apiRouter.HandleFunc("/candidates", getCandidates).Methods("GET")
	apiRouter.HandleFunc("/candidates", addCandidate).Methods("POST")
	apiRouter.HandleFunc("/candidates/{id}", getCandidateByID).Methods("GET")
	apiRouter.HandleFunc("/candidates/{id}", updateCandidate).Methods("PUT")
	apiRouter.HandleFunc("/candidates/{id}", deleteCandidate).Methods("DELETE")

	apiRouter.HandleFunc("/jobs", getJobs).Methods("GET")
	apiRouter.HandleFunc("/jobs", addJob).Methods("POST")
	apiRouter.HandleFunc("/jobs/{id}", getJobByID).Methods("GET")
	apiRouter.HandleFunc("/jobs/{id}", updateJob).Methods("PUT")
	apiRouter.HandleFunc("/jobs/{id}", deleteJob).Methods("DELETE")

	apiRouter.HandleFunc("/daily-jobs", getDailyJobs).Methods("GET")
	apiRouter.HandleFunc("/daily-jobs", addDailyJob).Methods("POST")
	apiRouter.HandleFunc("/daily-jobs/{id}", getDailyJobByID).Methods("GET")
	apiRouter.HandleFunc("/daily-jobs/{id}", updateDailyJob).Methods("PUT")
	apiRouter.HandleFunc("/daily-jobs/{id}", deleteDailyJob).Methods("DELETE")

	apiRouter.HandleFunc("/interviews", getInterviews).Methods("GET")
	apiRouter.HandleFunc("/interviews", scheduleInterview).Methods("POST")
	apiRouter.HandleFunc("/interviews/{id}", getInterviewByID).Methods("GET")
	apiRouter.HandleFunc("/interviews/{id}", updateInterview).Methods("PUT")
	apiRouter.HandleFunc("/interviews/{id}", deleteInterview).Methods("DELETE")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Start server
	port := getEnv("PORT", "8080")
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, c.Handler(r)))
}

func initDB() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "skillsifter")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}

	fmt.Println("Successfully connected to database")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"message":"Error marshalling JSON"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ApiResponse{
		Success: false,
		Message: message,
	})
}

// Authentication middleware
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Check if the header has the format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header format must be 'Bearer <token>'")
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Store claims in context for further use
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "companyID", claims.CompanyID)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Role-based middleware
func roleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role from context (set by authMiddleware)
			role := r.Context().Value("role").(string)
			
			// Check if the role is allowed
			allowed := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					allowed = true
					break
				}
			}
			
			if !allowed {
				respondWithError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// Registration handler
func registerUser(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
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
	tx, err := db.Begin()
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

	// Get appropriate role ID
	var roleID int
	if isFirstUser {
		// First user is admin
		err = tx.QueryRow("SELECT id FROM roles WHERE name = 'admin'").Scan(&roleID)
	} else {
		// Default role is recruiter
		err = tx.QueryRow("SELECT id FROM roles WHERE name = 'recruiter'").Scan(&roleID)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not determine role")
		return
	}

	// Insert user with company ID
	var userID int
	err = tx.QueryRow(`
        INSERT INTO users(username, email, password, role_id, company_id, created_at) 
        VALUES($1, $2, $3, $4, $5, $6) RETURNING id`,
		creds.Username, creds.Email, hashedPassword, roleID, companyID, time.Now()).Scan(&userID)

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

	// Get role name for response
	var roleName string
	err = db.QueryRow("SELECT name FROM roles WHERE id = $1", roleID).Scan(&roleName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch role")
		return
	}

	// Create user object for response (without password)
	user := User{
		ID:        userID,
		Username:  creds.Username,
		Email:     creds.Email,
		RoleID:    roleID,
		Role:      roleName,
		CompanyID: companyID,
		CreatedAt: time.Now(),
	}

	// Create JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      roleName,
		CompanyID: user.CompanyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	// Return token and user info
	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "User registered successfully",
		Data: TokenResponse{
			Token: tokenString,
			User:  user,
		},
	})
}

// Login handler
func loginUser(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Get user from database
	var user User
	var hashedPassword string
	var roleName string

	err = db.QueryRow(`
		SELECT u.id, u.username, u.email, u.password, u.role_id, r.name, u.company_id, u.created_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1`, creds.Email).Scan(
		&user.ID, &user.Username, &user.Email, &hashedPassword,
		&user.RoleID, &roleName, &user.CompanyID, &user.CreatedAt)

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

	// Set role name in user object
	user.Role = roleName
	user.Password = "" // Don't return the password

	// Create JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      roleName,
		CompanyID: user.CompanyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	// Return token and user info
	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Login successful",
		Data: TokenResponse{
			Token: tokenString,
			User:  user,
		},
	})
}

// User management handlers (admin only)
func getUsers(w http.ResponseWriter, r *http.Request) {
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	users := []User{}
	rows, err := db.Query(`
		SELECT u.id, u.username, u.email, u.role_id, r.name, u.company_id, u.created_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.company_id = $1`, companyID)
		
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching users")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.RoleID, &u.Role, &u.CompanyID, &u.CreatedAt); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning user row")
			return
		}
		users = append(users, u)
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

func createUser(w http.ResponseWriter, r *http.Request) {
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	var user User
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
	err = db.QueryRow(`
        INSERT INTO users(username, email, password, role_id, company_id, created_at) 
        VALUES($1, $2, $3, $4, $5, $6) RETURNING id`,
		user.Username, user.Email, hashedPassword, user.RoleID, user.CompanyID, time.Now()).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	// Get role name for response
	var roleName string
	err = db.QueryRow("SELECT name FROM roles WHERE id = $1", user.RoleID).Scan(&roleName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch role")
		return
	}

	// Create user object for response (without password)
	user.ID = userID
	user.Role = roleName
	user.Password = ""

	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	})
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Check if user exists and belongs to this company
	var existingCompanyID int
	err = db.QueryRow("SELECT company_id FROM users WHERE id = $1", id).Scan(&existingCompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if existingCompanyID != companyID {
		respondWithError(w, http.StatusForbidden, "Cannot modify user from another company")
		return
	}

	// Update query parts
	updateParts := []string{}
	updateArgs := []interface{}{id} // First arg is the ID
	argPosition := 2 // Start at position 2

	// Add fields to update
	if user.Username != "" {
		updateParts = append(updateParts, fmt.Sprintf("username = $%d", argPosition))
		updateArgs = append(updateArgs, user.Username)
		argPosition++
	}

	if user.Email != "" {
		updateParts = append(updateParts, fmt.Sprintf("email = $%d", argPosition))
		updateArgs = append(updateArgs, user.Email)
		argPosition++
	}

	if user.Password != "" {
		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not hash password")
			return
		}
		updateParts = append(updateParts, fmt.Sprintf("password = $%d", argPosition))
		updateArgs = append(updateArgs, string(hashedPassword))
		argPosition++
	}

	if user.RoleID != 0 {
		updateParts = append(updateParts, fmt.Sprintf("role_id = $%d", argPosition))
		updateArgs = append(updateArgs, user.RoleID)
		argPosition++
	}

	// If nothing to update
	if len(updateParts) == 0 {
		respondWithJSON(w, http.StatusOK, ApiResponse{
			Success: true,
			Message: "No fields to update",
		})
		return
	}

	// Build and execute update query
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $1", strings.Join(updateParts, ", "))
	result, err := db.Exec(query, updateArgs...)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not update user")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusInternalServerError, "User update failed")
		return
	}

	// Fetch updated user data
	var updatedUser User
	var roleName string
	err = db.QueryRow(`
		SELECT u.id, u.username, u.email, u.role_id, r.name, u.company_id, u.created_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1`, id).Scan(
		&updatedUser.ID, &updatedUser.Username, &updatedUser.Email,
		&updatedUser.RoleID, &roleName, &updatedUser.CompanyID, &updatedUser.CreatedAt)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch updated user")
		return
	}

	updatedUser.Role = roleName

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    updatedUser,
	})
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	// Check if user exists and belongs to this company
	var existingCompanyID int
	err := db.QueryRow("SELECT company_id FROM users WHERE id = $1", id).Scan(&existingCompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if existingCompanyID != companyID {
		respondWithError(w, http.StatusForbidden, "Cannot delete user from another company")
		return
	}

	// Delete the user
	result, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete user")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusInternalServerError, "User deletion failed")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// Company/tenant-specific handlers for existing entities
func getCandidates(w http.ResponseWriter, r *http.Request) {
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	candidates := []Candidate{}
	
	rows, err := db.Query("SELECT id, name, email, phone, position, status, date_applied FROM candidates WHERE company_id = $1", companyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching candidates")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var c Candidate
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, &c.Status, &c.DateApplied); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning candidate row")
			return
		}
		c.CompanyID = companyID
		candidates = append(candidates, c)
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Candidates retrieved successfully",
		Data:    candidates,
	})
}

func getCandidateByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)

	var c Candidate
	err := db.QueryRow(`
		SELECT id, name, email, phone, position, status, date_applied, resume_url, cover_letter 
		FROM candidates 
		WHERE id = $1 AND company_id = $2`, id, companyID).
		Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, &c.Status, &c.DateApplied, &c.ResumeURL, &c.CoverLetter)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Candidate not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error fetching candidate")
		return
	}

	c.CompanyID = companyID
	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Candidate retrieved successfully",
		Data:    c,
	})
}

func addCandidate(w http.ResponseWriter, r *http.Request) {
	var c Candidate
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Set company ID from context
	c.CompanyID = r.Context().Value("companyID").(int)

	stmt, err := db.Prepare(`
		INSERT INTO candidates(name, email, phone, position, status, date_applied, resume_url, cover_letter, company_id) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id`)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error preparing statement")
		return
	}

	var id int
	err = stmt.QueryRow(c.Name, c.Email, c.Phone, c.Position, c.Status, c.DateApplied, c.ResumeURL, c.CoverLetter, c.CompanyID).Scan(&id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error inserting candidate")
		return
	}

	c.ID = id
	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "Candidate created successfully",
		Data:    c,
	})
}

func updateCandidate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	var c Candidate
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Check if candidate exists and belongs to this company
	var existingCompanyID int
	err := db.QueryRow("SELECT company_id FROM candidates WHERE id = $1", id).Scan(&existingCompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Candidate not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if existingCompanyID != companyID {
		respondWithError(w, http.StatusForbidden, "Cannot modify candidate from another company")
		return
	}

	// Update the candidate
	_, err = db.Exec(`
		UPDATE candidates 
		SET name = $1, email = $2, phone = $3, position = $4, status = $5, 
		    resume_url = $6, cover_letter = $7, last_modified = $8
		WHERE id = $9 AND company_id = $10`,
		c.Name, c.Email, c.Phone, c.Position, c.Status,
		c.ResumeURL, c.CoverLetter, time.Now(), id, companyID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating candidate")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Candidate updated successfully",
		Data:    c,
	})
}

func deleteCandidate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	// Check if candidate exists and belongs to this company
	var existingCompanyID int
	err := db.QueryRow("SELECT company_id FROM candidates WHERE id = $1", id).Scan(&existingCompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Candidate not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if existingCompanyID != companyID {
		respondWithError(w, http.StatusForbidden, "Cannot delete candidate from another company")
		return
	}

	// Delete the candidate
	result, err := db.Exec("DELETE FROM candidates WHERE id = $1 AND company_id = $2", id, companyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete candidate")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusInternalServerError, "Candidate deletion failed")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Candidate deleted successfully",
	})
}

// Jobs handlers
func getJobs(w http.ResponseWriter, r *http.Request) {
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	jobs := []Job{}
	
	rows, err := db.Query("SELECT id, title, department, location, status, date_posted FROM jobs WHERE company_id = $1", companyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching jobs")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.Title, &j.Department, &j.Location, &j.Status, &j.DatePosted); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning job row")
			return
		}
		j.CompanyID = companyID
		jobs = append(jobs, j)
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Jobs retrieved successfully",
		Data:    jobs,
	})
}

func getJobByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)

	var j Job
	err := db.QueryRow(`
		SELECT id, title, department, location, status, date_posted, description, requirements
		FROM jobs 
		WHERE id = $1 AND company_id = $2`, id, companyID).
		Scan(&j.ID, &j.Title, &j.Department, &j.Location, &j.Status, &j.DatePosted, &j.Description, &j.Requirements)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Job not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error fetching job")
		return
	}

	j.CompanyID = companyID
	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Job retrieved successfully",
		Data:    j,
	})
}

func addJob(w http.ResponseWriter, r *http.Request) {
	var j Job
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&j); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Set company ID from context
	j.CompanyID = r.Context().Value("companyID").(int)

	stmt, err := db.Prepare(`
		INSERT INTO jobs(title, department, location, status, date_posted, description, requirements, company_id) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id`)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error preparing statement")
		return
	}

	var id int
	err = stmt.QueryRow(j.Title, j.Department, j.Location, j.Status, time.Now(), j.Description, j.Requirements, j.CompanyID).Scan(&id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error inserting job")
		return
	}

	j.ID = id
	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "Job created successfully",
		Data:    j,
	})
}

func updateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	var j Job
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&j); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Check if job exists and belongs to this company
	var existingCompanyID int
	err := db.QueryRow("SELECT company_id FROM jobs WHERE id = $1", id).Scan(&existingCompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Job not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if existingCompanyID != companyID {
		respondWithError(w, http.StatusForbidden, "Cannot modify job from another company")
		return
	}

	// Update the job
	_, err = db.Exec(`
		UPDATE jobs 
		SET title = $1, department = $2, location = $3, status = $4, 
		    description = $5, requirements = $6, last_modified = $7
		WHERE id = $8 AND company_id = $9`,
		j.Title, j.Department, j.Location, j.Status,
		j.Description, j.Requirements, time.Now(), id, companyID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating job")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Job updated successfully",
		Data:    j,
	})
}

func deleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	// Check if job exists and belongs to this company
	var existingCompanyID int
	err := db.QueryRow("SELECT company_id FROM jobs WHERE id = $1", id).Scan(&existingCompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Job not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if existingCompanyID != companyID {
		respondWithError(w, http.StatusForbidden, "Cannot delete job from another company")
		return
	}

	// Delete the job
	result, err := db.Exec("DELETE FROM jobs WHERE id = $1 AND company_id = $2", id, companyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete job")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusInternalServerError, "Job deletion failed")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Job deleted successfully",
	})
}

// Daily Jobs handlers
func getDailyJobs(w http.ResponseWriter, r *http.Request) {
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	dailyJobs := []DailyJob{}
	
	rows, err := db.Query("SELECT id, jd_no, instructions, assigned_user, assigned_date FROM daily_jobs WHERE company_id = $1", companyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching daily jobs")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var dj DailyJob
		if err := rows.Scan(&dj.ID, &dj.JdNo, &dj.Instructions, &dj.AssignedUser, &dj.AssignedDate); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning daily job row")
			return
		}
		dj.CompanyID = companyID
		dailyJobs = append(dailyJobs, dj)
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Daily jobs retrieved successfully",
		Data:    dailyJobs,
	})
}

func getDailyJobByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)

	var dj DailyJob
	err := db.QueryRow(`
		SELECT id, jd_no, instructions, assigned_user, assigned_date
		FROM daily_jobs 
		WHERE id = $1 AND company_id = $2`, id, companyID).
		Scan(&dj.ID, &dj.JdNo, &dj.Instructions, &dj.AssignedUser, &dj.AssignedDate)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Daily job not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error fetching daily job")
		return
	}

	dj.CompanyID = companyID
	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Daily job retrieved successfully",
		Data:    dj,
	})
}

func addDailyJob(w http.ResponseWriter, r *http.Request) {
	var dj DailyJob
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&dj); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Set company ID from context
	dj.CompanyID = r.Context().Value("companyID").(int)

	// Check if assigned user belongs to the same company
	if dj.AssignedUser != 0 {
		var userCompanyID int
		err := db.QueryRow("SELECT company_id FROM users WHERE id = $1", dj.AssignedUser).Scan(&userCompanyID)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusBadRequest, "Assigned user not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}

		if userCompanyID != dj.CompanyID {
			respondWithError(w, http.StatusForbidden, "Cannot assign job to user from another company")
			return
		}
	}

	stmt, err := db.Prepare(`
		INSERT INTO daily_jobs(jd_no, instructions, assigned_user, assigned_date, company_id) 
		VALUES($1, $2, $3, $4, $5) 
		RETURNING id`)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error preparing statement")
		return
	}

	var id int
	err = stmt.QueryRow(dj.JdNo, dj.Instructions, dj.AssignedUser, time.Now(), dj.CompanyID).Scan(&id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error inserting daily job")
		return
	}

	dj.ID = id
	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "Daily job created successfully",
		Data:    dj,
	})
}

func updateDailyJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	var dj DailyJob
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&dj); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Check if daily job exists and belongs to this company
	var existingCompanyID int
	err := db.QueryRow("SELECT company_id FROM daily_jobs WHERE id = $1", id).Scan(&existingCompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Daily job not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if existingCompanyID != companyID {
		respondWithError(w, http.StatusForbidden, "Cannot modify daily job from another company")
		return
	}

	// Check if assigned user belongs to the same company
	if dj.AssignedUser != 0 {
		var userCompanyID int
		err := db.QueryRow("SELECT company_id FROM users WHERE id = $1", dj.AssignedUser).Scan(&userCompanyID)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusBadRequest, "Assigned user not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}

		if userCompanyID != companyID {
			respondWithError(w, http.StatusForbidden, "Cannot assign job to user from another company")
			return
		}
	}

	// Update the daily job
	_, err = db.Exec(`
		UPDATE daily_jobs 
		SET jd_no = $1, instructions = $2, assigned_user = $3, last_modified = $4
		WHERE id = $5 AND company_id = $6`,
		dj.JdNo, dj.Instructions, dj.AssignedUser, time.Now(), id, companyID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating daily job")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Daily job updated successfully",
		Data:    dj,
	})
}

func deleteDailyJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	// Check if daily job exists and belongs to this company
	var existingCompanyID int
	err := db.QueryRow("SELECT company_id FROM daily_jobs WHERE id = $1", id).Scan(&existingCompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Daily job not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if existingCompanyID != companyID {
		respondWithError(w, http.StatusForbidden, "Cannot delete daily job from another company")
		return
	}

	// Delete the daily job
	result, err := db.Exec("DELETE FROM daily_jobs WHERE id = $1 AND company_id = $2", id, companyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete daily job")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusInternalServerError, "Daily job deletion failed")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Daily job deleted successfully",
	})
}

// Interview handlers
func getInterviews(w http.ResponseWriter, r *http.Request) {
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	interviews := []Interview{}
	
	rows, err := db.Query("SELECT id, candidate_id, candidate_name, position, interview_date, status FROM interviews WHERE company_id = $1", companyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching interviews")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var i Interview
		if err := rows.Scan(&i.ID, &i.CandidateID, &i.CandidateName, &i.Position, &i.InterviewDate, &i.Status); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning interview row")
			return
		}
		i.CompanyID = companyID
		interviews = append(interviews, i)
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Interviews retrieved successfully",
		Data:    interviews,
	})
}

func getInterviewByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)

	var i Interview
	err := db.QueryRow(`
		SELECT id, candidate_id, candidate_name, position, interview_date, status, feedback
		FROM interviews 
		WHERE id = $1 AND company_id = $2`, id, companyID).
		Scan(&i.ID, &i.CandidateID, &i.CandidateName, &i.Position, &i.InterviewDate, &i.Status, &i.Feedback)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Interview not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error fetching interview")
		return
	}

	i.CompanyID = companyID
	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Interview retrieved successfully",
		Data:    i,
	})
}

func scheduleInterview(w http.ResponseWriter, r *http.Request) {
	var i Interview
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&i); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Set company ID from context
	i.CompanyID = r.Context().Value("companyID").(int)

	// Check if candidate belongs to the same company if candidate ID is provided
	if i.CandidateID != 0 {
		var candidateCompanyID int
		err := db.QueryRow("SELECT company_id FROM candidates WHERE id = $1", i.CandidateID).Scan(&candidateCompanyID)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusBadRequest, "Candidate not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}

		if candidateCompanyID != i.CompanyID {
			respondWithError(w, http.StatusForbidden, "Cannot schedule interview for candidate from another company")
			return
		}
	}

	stmt, err := db.Prepare(`
		INSERT INTO interviews(candidate_id, candidate_name, position, interview_date, status, feedback, company_id) 
		VALUES($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id`)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error preparing statement")
		return
	}

	var id int
	err = stmt.QueryRow(i.CandidateID, i.CandidateName, i.Position, i.InterviewDate, i.Status, i.Feedback, i.CompanyID).Scan(&id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error scheduling interview")
		return
	}

	i.ID = id
	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "Interview scheduled successfully",
		Data:    i,
	})
}

func updateInterview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	var i Interview
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&i); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Check if interview exists and belongs to this company
	var existingCompanyID int
	err := db.QueryRow("SELECT company_id FROM interviews WHERE id = $1", id).Scan(&existingCompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Interview not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if existingCompanyID != companyID {
		respondWithError(w, http.StatusForbidden, "Cannot modify interview from another company")
		return
	}

	// Update the interview
	_, err = db.Exec(`
		UPDATE interviews 
		SET candidate_name = $1, position = $2, interview_date = $3, 
		    status = $4, feedback = $5, last_modified = $6
		WHERE id = $7 AND company_id = $8`,
		i.CandidateName, i.Position, i.InterviewDate,
		i.Status, i.Feedback, time.Now(), id, companyID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating interview")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Interview updated successfully",
		Data:    i,
	})
}

func deleteInterview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(int)
	
	// Check if interview exists and belongs to this company
	var existingCompanyID int
	err := db.QueryRow("SELECT company_id FROM interviews WHERE id = $1", id).Scan(&existingCompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Interview not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if existingCompanyID != companyID {
		respondWithError(w, http.StatusForbidden, "Cannot delete interview from another company")
		return
	}

	// Delete the interview
	result, err := db.Exec("DELETE FROM interviews WHERE id = $1 AND company_id = $2", id, companyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete interview")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusInternalServerError, "Interview deletion failed")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Interview deleted successfully",
	})
}
