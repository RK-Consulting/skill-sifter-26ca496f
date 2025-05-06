
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

	// Root route handlers - respond to both root paths
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Welcome to SkillSifter API"}`))
	}).Methods("GET", "OPTIONS")

	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Welcome to SkillSifter API"}`))
	}).Methods("GET", "OPTIONS")
	
	// IMPORTANT: Health check routes at both root and with /api prefix
	// These need to be registered in both ways for Nginx proxy compatibility
	r.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"OK"}`))
	}).Methods("GET", "OPTIONS")
	
	r.HandleFunc("/api/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"OK"}`))
	}).Methods("GET", "OPTIONS")
	
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"pong"}`))
	}).Methods("GET", "OPTIONS")
	
	r.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"pong"}`))
	}).Methods("GET", "OPTIONS")

	// Auth routes need to be registered for both prefixed and non-prefixed paths
	r.HandleFunc("/auth/register", handlers.RegisterUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/register", handlers.RegisterUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/login", handlers.LoginUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/login", handlers.LoginUser).Methods("POST", "OPTIONS")

	// Protected routes with API prefix
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(auth.AuthMiddleware)

	// Admin-only routes
	adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(auth.RoleMiddleware("admin"))
	adminRouter.HandleFunc("/users", handlers.GetUsers).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/users", handlers.CreateUser).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT", "OPTIONS")
	adminRouter.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE", "OPTIONS")

	// Manager and Admin routes
	managerRouter := apiRouter.PathPrefix("/manager").Subrouter()
	managerRouter.Use(auth.RoleMiddleware("manager", "admin"))
	// Add any manager-specific endpoints here if needed

	// General API Routes (accessible by all roles)
	apiRouter.HandleFunc("/candidates", handlers.GetCandidates).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/candidates", handlers.AddCandidate).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/candidates/{id}", handlers.GetCandidateByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/candidates/{id}", handlers.UpdateCandidate).Methods("PUT", "OPTIONS")
	apiRouter.HandleFunc("/candidates/{id}", handlers.DeleteCandidate).Methods("DELETE", "OPTIONS")

	apiRouter.HandleFunc("/jobs", handlers.GetJobs).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/jobs", handlers.AddJob).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/jobs/{id}", handlers.GetJobByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/jobs/{id}", handlers.UpdateJob).Methods("PUT", "OPTIONS")
	apiRouter.HandleFunc("/jobs/{id}", handlers.DeleteJob).Methods("DELETE", "OPTIONS")

	apiRouter.HandleFunc("/daily-jobs", handlers.GetDailyJobs).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/daily-jobs", handlers.AddDailyJob).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/daily-jobs/{id}", handlers.GetDailyJobByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/daily-jobs/{id}", handlers.UpdateDailyJob).Methods("PUT", "OPTIONS")
	apiRouter.HandleFunc("/daily-jobs/{id}", handlers.DeleteDailyJob).Methods("DELETE", "OPTIONS")

	apiRouter.HandleFunc("/interviews", handlers.GetInterviews).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/interviews", handlers.ScheduleInterview).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/interviews/{id}", handlers.GetInterviewByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/interviews/{id}", handlers.UpdateInterview).Methods("PUT", "OPTIONS")
	apiRouter.HandleFunc("/interviews/{id}", handlers.DeleteInterview).Methods("DELETE", "OPTIONS")

	// Also register non-prefixed routes to work both with and without /api prefix
	nonPrefixRouter := r.NewRoute().Subrouter()
	nonPrefixRouter.Use(auth.AuthMiddleware)
	
	// Register the same routes without the /api prefix
	nonPrefixRouter.HandleFunc("/candidates", handlers.GetCandidates).Methods("GET", "OPTIONS")
	nonPrefixRouter.HandleFunc("/candidates", handlers.AddCandidate).Methods("POST", "OPTIONS")
	nonPrefixRouter.HandleFunc("/candidates/{id}", handlers.GetCandidateByID).Methods("GET", "OPTIONS")
	nonPrefixRouter.HandleFunc("/candidates/{id}", handlers.UpdateCandidate).Methods("PUT", "OPTIONS")
	nonPrefixRouter.HandleFunc("/candidates/{id}", handlers.DeleteCandidate).Methods("DELETE", "OPTIONS")
	
	nonPrefixRouter.HandleFunc("/jobs", handlers.GetJobs).Methods("GET", "OPTIONS")
	// ... register other routes similarly to support both prefixed and non-prefixed versions

	// Setup CORS - Updated with broader configuration
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://skillsifter.in",
			"https://www.skillsifter.in",
			"https://api.skillsifter.in",
			"http://localhost:5173",
			"http://localhost:3000",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:3000",
			"*", // Allow all origins for development
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
	
	// Always respond to OPTIONS requests for all routes
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Wrap the router with CORS handler
	handler := c.Handler(r)
	
	// Start HTTP server
	port := db.GetEnv("PORT", "8080")
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
