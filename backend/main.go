
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

// Database connection
var db *sql.DB

// Candidate represents a job candidate
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
}

// Job represents a job posting
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
}

// DailyJob represents a daily task assignment
type DailyJob struct {
	ID           int       `json:"id"`
	JdNo         int       `json:"jdNo"`
	Instructions string    `json:"instructions"`
	AssignedUser int       `json:"assignedUser"`
	AssignedDate time.Time `json:"assignedDate"`
	LastModified time.Time `json:"lastModified"`
}

// Interview represents a candidate interview
type Interview struct {
	ID            int       `json:"id"`
	CandidateID   int       `json:"candidateId"`
	CandidateName string    `json:"candidateName"`
	Position      string    `json:"position"`
	InterviewDate time.Time `json:"interviewDate"`
	Status        string    `json:"status"`
	Feedback      string    `json:"feedback"`
	LastModified  time.Time `json:"lastModified"`
}

// User represents a system user
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// ApiResponse represents a standard API response
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	// Initialize database connection
	initDB()
	defer db.Close()

	// Create router
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/candidates", getCandidates).Methods("GET")
	r.HandleFunc("/api/candidates", addCandidate).Methods("POST")
	r.HandleFunc("/api/candidates/{id}", getCandidateByID).Methods("GET")

	r.HandleFunc("/api/jobs", getJobs).Methods("GET")
	r.HandleFunc("/api/jobs", addJob).Methods("POST")
	r.HandleFunc("/api/jobs/{id}", getJobByID).Methods("GET")

	r.HandleFunc("/api/daily-jobs", getDailyJobs).Methods("GET")
	r.HandleFunc("/api/daily-jobs", addDailyJob).Methods("POST")
	r.HandleFunc("/api/daily-jobs/{id}", getDailyJobByID).Methods("GET")

	r.HandleFunc("/api/interviews", getInterviews).Methods("GET")
	r.HandleFunc("/api/interviews", scheduleInterview).Methods("POST")
	r.HandleFunc("/api/interviews/{id}", getInterviewByID).Methods("GET")

	r.HandleFunc("/api/auth/register", registerUser).Methods("POST")
	r.HandleFunc("/api/auth/login", loginUser).Methods("POST")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
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

// Candidate endpoints
func getCandidates(w http.ResponseWriter, r *http.Request) {
	candidates := []Candidate{}
	
	rows, err := db.Query("SELECT id, name, email, phone, position, status, date_applied FROM candidates")
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

	var c Candidate
	err := db.QueryRow("SELECT id, name, email, phone, position, status, date_applied, resume_url, cover_letter FROM candidates WHERE id = $1", id).
		Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, &c.Status, &c.DateApplied, &c.ResumeURL, &c.CoverLetter)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Candidate not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error fetching candidate")
		return
	}

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

	stmt, err := db.Prepare("INSERT INTO candidates(name, email, phone, position, status, date_applied, resume_url, cover_letter) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error preparing statement")
		return
	}

	var id int
	err = stmt.QueryRow(c.Name, c.Email, c.Phone, c.Position, c.Status, c.DateApplied, c.ResumeURL, c.CoverLetter).Scan(&id)
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

// Job endpoints
func getJobs(w http.ResponseWriter, r *http.Request) {
	jobs := []Job{}
	
	rows, err := db.Query("SELECT id, title, department, location, status, date_posted FROM jobs")
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

	var j Job
	err := db.QueryRow("SELECT id, title, department, location, status, date_posted, description, requirements FROM jobs WHERE id = $1", id).
		Scan(&j.ID, &j.Title, &j.Department, &j.Location, &j.Status, &j.DatePosted, &j.Description, &j.Requirements)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Job not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error fetching job")
		return
	}

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

	stmt, err := db.Prepare("INSERT INTO jobs(title, department, location, status, date_posted, description, requirements) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error preparing statement")
		return
	}

	var id int
	err = stmt.QueryRow(j.Title, j.Department, j.Location, j.Status, j.DatePosted, j.Description, j.Requirements).Scan(&id)
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

// DailyJob endpoints
func getDailyJobs(w http.ResponseWriter, r *http.Request) {
	dailyJobs := []DailyJob{}
	
	rows, err := db.Query("SELECT id, jd_no, instructions, assigned_user, assigned_date FROM daily_jobs")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching daily jobs")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var d DailyJob
		if err := rows.Scan(&d.ID, &d.JdNo, &d.Instructions, &d.AssignedUser, &d.AssignedDate); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning daily job row")
			return
		}
		dailyJobs = append(dailyJobs, d)
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

	var d DailyJob
	err := db.QueryRow("SELECT id, jd_no, instructions, assigned_user, assigned_date FROM daily_jobs WHERE id = $1", id).
		Scan(&d.ID, &d.JdNo, &d.Instructions, &d.AssignedUser, &d.AssignedDate)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Daily job not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error fetching daily job")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Daily job retrieved successfully",
		Data:    d,
	})
}

func addDailyJob(w http.ResponseWriter, r *http.Request) {
	var d DailyJob
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&d); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	stmt, err := db.Prepare("INSERT INTO daily_jobs(jd_no, instructions, assigned_user, assigned_date) VALUES($1, $2, $3, $4) RETURNING id")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error preparing statement")
		return
	}

	var id int
	err = stmt.QueryRow(d.JdNo, d.Instructions, d.AssignedUser, time.Now()).Scan(&id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error inserting daily job")
		return
	}

	d.ID = id
	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "Daily job created successfully",
		Data:    d,
	})
}

// Interview endpoints
func getInterviews(w http.ResponseWriter, r *http.Request) {
	interviews := []Interview{}
	
	rows, err := db.Query("SELECT id, candidate_id, candidate_name, position, interview_date, status, feedback FROM interviews")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching interviews")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var i Interview
		if err := rows.Scan(&i.ID, &i.CandidateID, &i.CandidateName, &i.Position, &i.InterviewDate, &i.Status, &i.Feedback); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning interview row")
			return
		}
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

	var i Interview
	err := db.QueryRow("SELECT id, candidate_id, candidate_name, position, interview_date, status, feedback FROM interviews WHERE id = $1", id).
		Scan(&i.ID, &i.CandidateID, &i.CandidateName, &i.Position, &i.InterviewDate, &i.Status, &i.Feedback)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Interview not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error fetching interview")
		return
	}

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

	stmt, err := db.Prepare("INSERT INTO interviews(candidate_id, candidate_name, position, interview_date, status, feedback) VALUES($1, $2, $3, $4, $5, $6) RETURNING id")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error preparing statement")
		return
	}

	var id int
	err = stmt.QueryRow(i.CandidateID, i.CandidateName, i.Position, i.InterviewDate, i.Status, i.Feedback).Scan(&id)
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

// Auth endpoints
func registerUser(w http.ResponseWriter, r *http.Request) {
	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// In a real app, you would hash the password here
	stmt, err := db.Prepare("INSERT INTO users(username, email, password, created_at) VALUES($1, $2, $3, $4) RETURNING id")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error preparing statement")
		return
	}

	var id int
	err = stmt.QueryRow(u.Username, u.Email, u.Password, time.Now()).Scan(&id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error registering user")
		return
	}

	u.ID = id
	u.Password = "" // Don't return the password
	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    u,
	})
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&credentials); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	var u User
	err := db.QueryRow("SELECT id, username, email FROM users WHERE email = $1 AND password = $2", 
		credentials.Email, credentials.Password).Scan(&u.ID, &u.Username, &u.Email)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error during login")
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Login successful",
		Data:    u,
	})
}
