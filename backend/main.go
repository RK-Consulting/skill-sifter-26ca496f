package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/RK-Consulting/skill-sifter/auth"
	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/handlers"
	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Logging middleware to track requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize database connection
	db.InitDB()
	defer db.DB.Close()

	// Initialize schema (one-time operation)
	if err := db.InitializeSchema(); err != nil {
		log.Fatalf("Schema initialization failed: %v", err)
	}

	// Create main router
	r := mux.NewRouter()
	r.Use(loggingMiddleware) // Add logging middleware

	// Root route handler - Respond to both root paths
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"SkillSifter API root"}`))
	}).Methods("GET")
	
	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"SkillSifter API root"}`))
	}).Methods("GET")
	
	// IMPORTANT: Health check routes at root level
	// These need to be registered WITHOUT any prefixes since Nginx strips /api/
	r.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"OK"}`))
	}).Methods("GET")
	
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"pong"}`))
	}).Methods("GET")

	// IMPORTANT: Direct auth routes without /api prefix
	// Since Nginx strips /api prefix, we need these registered at root level
	r.HandleFunc("/auth/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/auth/login", handlers.LoginUser).Methods("POST")

	// Also register health checks and auth routes WITH /api prefix for direct access
	r.HandleFunc("/api/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"OK"}`))
	}).Methods("GET")
	
	r.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"pong"}`))
	}).Methods("GET")

	r.HandleFunc("/api/auth/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/api/auth/login", handlers.LoginUser).Methods("POST")

	// Protected routes - maintain /api prefix for consistency
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(auth.AuthMiddleware)

	// Admin-only routes
	adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(auth.RoleMiddleware("admin"))
	adminRouter.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	adminRouter.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	adminRouter.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT")
	adminRouter.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE")

	// Manager and Admin routes
	managerRouter := apiRouter.PathPrefix("/manager").Subrouter()
	managerRouter.Use(auth.RoleMiddleware("manager", "admin"))
	// Add any manager-specific endpoints here if needed

	// General API Routes (accessible by all roles)
	apiRouter.HandleFunc("/candidates", handlers.GetCandidates).Methods("GET")
	apiRouter.HandleFunc("/candidates", handlers.AddCandidate).Methods("POST")
	apiRouter.HandleFunc("/candidates/{id}", handlers.GetCandidateByID).Methods("GET")
	apiRouter.HandleFunc("/candidates/{id}", handlers.UpdateCandidate).Methods("PUT")
	apiRouter.HandleFunc("/candidates/{id}", handlers.DeleteCandidate).Methods("DELETE")

	apiRouter.HandleFunc("/jobs", handlers.GetJobs).Methods("GET")
	apiRouter.HandleFunc("/jobs", handlers.AddJob).Methods("POST")
	apiRouter.HandleFunc("/jobs/{id}", handlers.GetJobByID).Methods("GET")
	apiRouter.HandleFunc("/jobs/{id}", handlers.UpdateJob).Methods("PUT")
	apiRouter.HandleFunc("/jobs/{id}", handlers.DeleteJob).Methods("DELETE")

	apiRouter.HandleFunc("/daily-jobs", handlers.GetDailyJobs).Methods("GET")
	apiRouter.HandleFunc("/daily-jobs", handlers.AddDailyJob).Methods("POST")
	apiRouter.HandleFunc("/daily-jobs/{id}", handlers.GetDailyJobByID).Methods("GET")
	apiRouter.HandleFunc("/daily-jobs/{id}", handlers.UpdateDailyJob).Methods("PUT")
	apiRouter.HandleFunc("/daily-jobs/{id}", handlers.DeleteDailyJob).Methods("DELETE")

	apiRouter.HandleFunc("/interviews", handlers.GetInterviews).Methods("GET")
	apiRouter.HandleFunc("/interviews", handlers.ScheduleInterview).Methods("POST")
	apiRouter.HandleFunc("/interviews/{id}", handlers.GetInterviewByID).Methods("GET")
	apiRouter.HandleFunc("/interviews/{id}", handlers.UpdateInterview).Methods("PUT")
	apiRouter.HandleFunc("/interviews/{id}", handlers.DeleteInterview).Methods("DELETE")

	// Setup CORS - Updated to allow ALL necessary origins and methods
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://skillsifter.in",
			"https://www.skillsifter.in",
			"https://api.skillsifter.in",
			"http://localhost:5173",
			"http://localhost:3000",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:3000",
			"*", // Allow all origins as a fallback (remove in production if possible)
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Origin", "Accept", "X-Requested-With", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours for preflight cache
	})

	// Log startup configuration
	log.Println("CORS Configuration:")
	log.Println("- Allowed Origins:", c.Options.AllowedOrigins)
	log.Println("- Allowed Methods:", c.Options.AllowedMethods)
	log.Println("- Allowed Headers:", c.Options.AllowedHeaders)

	// Wrap the router
	handler := c.Handler(r)
	
	// Start HTTP server
	port := db.GetEnv("PORT", "8080")
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
