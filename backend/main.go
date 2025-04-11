package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
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
	ID           int       `json:"id" db:"id,primarykey,autoincrement"`
	Name         string    `json:"name" db:"name,notnull"`
	Email        string    `json:"email" db:"email,notnull"`
	Phone        string    `json:"phone" db:"phone"`
	Position     string    `json:"position" db:"position"`
	Status       string    `json:"status" db:"status,default:'applied'"`
	DateApplied  time.Time `json:"dateApplied" db:"date_applied,default:CURRENT_TIMESTAMP"`
	ResumeURL    string    `json:"resumeUrl,omitempty" db:"resume_url"`
	CoverLetter  string    `json:"coverLetter,omitempty" db:"cover_letter"`
	LastModified time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	CompanyID    int       `json:"companyId" db:"company_id,notnull,foreignkey:companies.id"`
}

type Job struct {
	ID           int       `json:"id" db:"id,primarykey,autoincrement"`
	Title        string    `json:"title" db:"title,notnull"`
	Department   string    `json:"department" db:"department"`
	Location     string    `json:"location" db:"location"`
	Status       string    `json:"status" db:"status,default:'open'"`
	DatePosted   time.Time `json:"datePosted" db:"date_posted,default:CURRENT_TIMESTAMP"`
	Description  string    `json:"description,omitempty" db:"description"`
	Requirements string    `json:"requirements,omitempty" db:"requirements"`
	LastModified time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	CompanyID    int       `json:"companyId" db:"company_id,notnull,foreignkey:companies.id"`
}

type DailyJob struct {
	ID           int       `json:"id" db:"id,primarykey,autoincrement"`
	JdNo         int       `json:"jdNo" db:"jd_no,notnull"`
	Instructions string    `json:"instructions" db:"instructions"`
	AssignedUser int       `json:"assignedUser" db:"assigned_user,foreignkey:users.id"`
	AssignedDate time.Time `json:"assignedDate" db:"assigned_date,default:CURRENT_TIMESTAMP"`
	LastModified time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	CompanyID    int       `json:"companyId" db:"company_id,notnull,foreignkey:companies.id"`
}

type Interview struct {
	ID            int       `json:"id" db:"id,primarykey,autoincrement"`
	CandidateID   int       `json:"candidateId" db:"candidate_id,foreignkey:candidates.id"`
	CandidateName string    `json:"candidateName" db:"candidate_name,notnull"`
	Position      string    `json:"position" db:"position"`
	InterviewDate time.Time `json:"interviewDate" db:"interview_date,notnull"`
	Status        string    `json:"status" db:"status,default:'scheduled'"`
	Feedback      string    `json:"feedback" db:"feedback"`
	LastModified  time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	CompanyID     int       `json:"companyId" db:"company_id,notnull,foreignkey:companies.id"`
}

// New Company/Tenant model
type Company struct {
	ID        int       `json:"id" db:"id,primarykey,autoincrement"`
	Name      string    `json:"name" db:"name,notnull,unique"`
	CreatedAt time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
}

// Role model
type Role struct {
	ID          int      `json:"id" db:"id,primarykey,autoincrement"`
	Name        string   `json:"name" db:"name,notnull,unique"`
	Permissions []string `json:"permissions" db:"permissions,type:jsonb,default:'[]'::jsonb"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
}

// Enhanced User model with role
type User struct {
	ID        int       `json:"id" db:"id,primarykey,autoincrement"`
	Username  string    `json:"username" db:"username,notnull"`
	Email     string    `json:"email" db:"email,notnull,unique"`
	Password  string    `json:"password,omitempty" db:"password,notnull"`
	RoleID    int       `json:"roleId" db:"role_id,notnull,foreignkey:roles.id"`
	Role      string    `json:"role,omitempty"` // Not stored in DB, for frontend
	CompanyID int       `json:"companyId" db:"company_id,notnull,foreignkey:companies.id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
}

// Schema version tracking
type SchemaVersion struct {
	Version   int       `json:"version" db:"version,primarykey"`
	CreatedAt time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
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
	
	// Initialize schema (one-time operation)
	if err := initializeSchema(); err != nil {
		log.Fatalf("Schema initialization failed: %v", err)
	}

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

// Initialize the database connection
func initDB() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "skillsifter")
	password := getEnv("DB_PASSWORD", "ROOT")
	dbname := getEnv("DB_NAME", "postgres")

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

// One-time schema initialization function
func initializeSchema() error {
	// Check if schema_version table exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'schema_version'
		)
	`).Scan(&exists)
	
	if err != nil {
		return fmt.Errorf("error checking schema_version table: %v", err)
	}
	
	// If schema_version table doesn't exist, create it and initialize schema
	if !exists {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("error starting transaction: %v", err)
		}
		defer tx.Rollback()
		
		// Create schema_version table
		_, err = tx.Exec(`
			CREATE TABLE schema_version (
				version INTEGER PRIMARY KEY,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("error creating schema_version table: %v", err)
		}
		
		// Create tables in correct order based on dependencies
		fmt.Println("Initializing database schema...")
		
		// 1. Companies table (no dependencies)
		if err := createTableFromStruct(tx, Company{}, "companies"); err != nil {
			return err
		}
		
		// 2. Roles table (no dependencies)
		if err := createTableFromStruct(tx, Role{}, "roles"); err != nil {
			return err
		}
		
		// 3. Users table (depends on roles and companies)
		if err := createTableFromStruct(tx, User{}, "users"); err != nil {
			return err
		}
		
		// 4. Candidates table (depends on companies)
		if err := createTableFromStruct(tx, Candidate{}, "candidates"); err != nil {
			return err
		}
		
		// 5. Jobs table (depends on companies)
		if err := createTableFromStruct(tx, Job{}, "jobs"); err != nil {
			return err
		}
		
		// 6. Daily Jobs table (depends on companies and users)
		if err := createTableFromStruct(tx, DailyJob{}, "daily_jobs"); err != nil {
			return err
		}
		
		// 7. Interviews table (depends on companies and candidates)
		if err := createTableFromStruct(tx, Interview{}, "interviews"); err != nil {
			return err
		}
		
		// Create indexes for performance
		_, err = tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_candidates_company ON candidates(company_id);
			CREATE INDEX IF NOT EXISTS idx_jobs_company ON jobs(company_id);
			CREATE INDEX IF NOT EXISTS idx_daily_jobs_company ON daily_jobs(company_id);
			CREATE INDEX IF NOT EXISTS idx_interviews_company ON interviews(company_id);
			CREATE INDEX IF NOT EXISTS idx_users_company ON users(company_id);
			CREATE INDEX IF NOT EXISTS idx_users_role ON users(role_id);
		`)
		
		if err != nil {
			return fmt.Errorf("error creating indexes: %v", err)
		}
		
		// Insert default roles
		_, err = tx.Exec(`
			INSERT INTO roles (name, permissions, created_at) VALUES 
			('admin', '["all"]', NOW()),
			('manager', '["manage_candidates", "manage_jobs", "manage_interviews", "view_reports"]', NOW()),
			('recruiter', '["view_candidates", "add_candidates", "view_jobs", "schedule_interviews"]', NOW()),
			('team_leader', '["view_candidates", "add_candidates", "view_jobs", "manage_team"]', NOW())
			ON CONFLICT (name) DO NOTHING;
		`)
		
		if err != nil {
			return fmt.Errorf("error inserting default roles: %v", err)
		}
		
		// Insert schema version
		_, err = tx.Exec("INSERT INTO schema_version (version) VALUES (1)")
		if err != nil {
			return fmt.Errorf("error inserting schema version: %v", err)
		}
		
		// Commit the transaction
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("error committing schema transaction: %v", err)
		}
		
		fmt.Println("Database schema initialized successfully!")
	} else {
		fmt.Println("Database schema already initialized, skipping...")
	}
	
	return nil
}

// Helper function to create a table from a struct
func createTableFromStruct(tx *sql.Tx, model interface{}, tableName string) error {
	// Check if table already exists
	var exists bool
	err := tx.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`, tableName).Scan(&exists)
	
	if err != nil {
		return fmt.Errorf("error checking if %s table exists: %v", tableName, err)
	}
	
	if exists {
		fmt.Printf("Table %s already exists, skipping creation\n", tableName)
		return nil
	}
	
	fmt.Printf("Creating table %s...\n", tableName)
	
	t := reflect.TypeOf(model)
	var columns []string
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		
		if dbTag == "" {
			continue // Skip fields without db tag
		}
		
		parts := strings.Split(dbTag, ",")
		colName := parts[0]
		
		if colName == "-" {
			continue // Skip explicitly ignored fields
		}
		
		colType := getPostgresType(field.Type)
		colDef := fmt.Sprintf("%s %s", colName, colType)
		
		// Add constraints from tags
		for _, part := range parts[1:] {
			switch {
			case part == "primarykey":
				colDef += " PRIMARY KEY"
			case part == "autoincrement":
				// For Postgres, we change int to SERIAL when it has autoincrement
				if strings.Contains(colType, "INTEGER") {
					colDef = fmt.Sprintf("%s SERIAL PRIMARY KEY", colName)
				}
			case part == "notnull":
				colDef += " NOT NULL"
			case part == "unique":
				colDef += " UNIQUE"
			case strings.HasPrefix(part, "default:"):
				defaultVal := strings.TrimPrefix(part, "default:")
				colDef += fmt.Sprintf(" DEFAULT %s", defaultVal)
			case strings.HasPrefix(part, "foreignkey:"):
				fkRef := strings.TrimPrefix(part, "foreignkey:")
				columns = append(columns, fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s", colName, fkRef))
				continue // Don't add this part to the column definition
			case strings.HasPrefix(part, "type:"):
				// Override detected type with explicit type
				typeVal := strings.TrimPrefix(part, "type:")
				colDef = fmt.Sprintf("%s %s", colName, typeVal)
			}
		}
		
		columns = append(columns, colDef)
	}
	
	createSQL := fmt.Sprintf("CREATE TABLE %s (\n\t%s\n)", tableName, strings.Join(columns, ",\n\t"))
	_, err = tx.Exec(createSQL)
	
	if err != nil {
		return fmt.Errorf("error creating %s table: %v\nSQL: %s", tableName, err, createSQL)
	}
	
	return nil
}

// Helper function to convert Go types to PostgreSQL types
func getPostgresType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "NUMERIC"
	case reflect.String:
		return "VARCHAR(255)"
	default:
		// Handle special types
		typeName := t.String()
		switch typeName {
		case "time.Time":
			return "TIMESTAMP"
		case "[]string":
			return "TEXT[]"
		case "map[string]interface {}", "map[string]interface{}":
			return "JSONB"
		default:
			return "TEXT"
		}
	}
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
		respondWithError(w, http.StatusInternalServerError, "
