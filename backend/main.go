
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

// Handler for root route
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Welcome to SkillSifter API"}`))
}

// Handler for API root route
func apiRootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Welcome to SkillSifter API"}`))
}

// Handler for health check
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"OK"}`))
}

// Handler for ping
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"pong"}`))
}

// Setup CORS configuration
func setupCORS() *cors.Cors {
	return cors.New(cors.Options{
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
}

// Setup public routes that don't require authentication
func setupPublicRoutes(r *mux.Router) {
	// Root route handlers - respond to both root paths
	r.HandleFunc("/", rootHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api", apiRootHandler).Methods("GET", "OPTIONS")
	
	// Health check routes
	r.HandleFunc("/health-check", healthCheckHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/health-check", healthCheckHandler).Methods("GET", "OPTIONS")
	
	// Ping routes
	r.HandleFunc("/ping", pingHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/ping", pingHandler).Methods("GET", "OPTIONS")

	// Auth routes
	r.HandleFunc("/auth/register", handlers.RegisterUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/register", handlers.RegisterUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/login", handlers.LoginUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/login", handlers.LoginUser).Methods("POST", "OPTIONS")
}

// Setup protected routes that require authentication
func setupProtectedRoutes(r *mux.Router) {
	// API Router (with authentication)
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
	setupResourceRoutes(apiRouter, "/candidates", handlers.GetCandidates, handlers.AddCandidate, 
		handlers.GetCandidateByID, handlers.UpdateCandidate, handlers.DeleteCandidate)
	
	setupResourceRoutes(apiRouter, "/jobs", handlers.GetJobs, handlers.AddJob, 
		handlers.GetJobByID, handlers.UpdateJob, handlers.DeleteJob)
	
	setupResourceRoutes(apiRouter, "/daily-jobs", handlers.GetDailyJobs, handlers.AddDailyJob, 
		handlers.GetDailyJobByID, handlers.UpdateDailyJob, handlers.DeleteDailyJob)
	
	setupResourceRoutes(apiRouter, "/interviews", handlers.GetInterviews, handlers.ScheduleInterview, 
		handlers.GetInterviewByID, handlers.UpdateInterview, handlers.DeleteInterview)

	// Also register non-prefixed routes to work both with and without /api prefix
	nonPrefixRouter := r.NewRoute().Subrouter()
	nonPrefixRouter.Use(auth.AuthMiddleware)
	
	// Register the same routes without the /api prefix for compatibility
	setupResourceRoutes(nonPrefixRouter, "/candidates", handlers.GetCandidates, handlers.AddCandidate, 
		handlers.GetCandidateByID, handlers.UpdateCandidate, handlers.DeleteCandidate)
	
	setupResourceRoutes(nonPrefixRouter, "/jobs", handlers.GetJobs, handlers.AddJob, 
		handlers.GetJobByID, handlers.UpdateJob, handlers.DeleteJob)
	
	setupResourceRoutes(nonPrefixRouter, "/daily-jobs", handlers.GetDailyJobs, handlers.AddDailyJob, 
		handlers.GetDailyJobByID, handlers.UpdateDailyJob, handlers.DeleteDailyJob)
	
	setupResourceRoutes(nonPrefixRouter, "/interviews", handlers.GetInterviews, handlers.ScheduleInterview, 
		handlers.GetInterviewByID, handlers.UpdateInterview, handlers.DeleteInterview)
}

// Helper function to set up CRUD routes for a resource
func setupResourceRoutes(router *mux.Router, path string, 
	getAll http.HandlerFunc, create http.HandlerFunc,
	getOne http.HandlerFunc, update http.HandlerFunc, delete http.HandlerFunc) {
	
	router.HandleFunc(path, getAll).Methods("GET", "OPTIONS")
	router.HandleFunc(path, create).Methods("POST", "OPTIONS")
	router.HandleFunc(path+"/{id}", getOne).Methods("GET", "OPTIONS")
	router.HandleFunc(path+"/{id}", update).Methods("PUT", "OPTIONS")
	router.HandleFunc(path+"/{id}", delete).Methods("DELETE", "OPTIONS")
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

	// Setup routes
	setupPublicRoutes(r)
	setupProtectedRoutes(r)
	
	// Always respond to OPTIONS requests for all routes
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Setup CORS
	c := setupCORS()
	
	// Log startup configuration
	log.Println("CORS Configuration:")
	log.Println("- Allowed Origins:", c.Options.AllowedOrigins)
	log.Println("- Allowed Methods:", c.Options.AllowedMethods)
	log.Println("- Allowed Headers:", c.Options.AllowedHeaders)
	
	// Wrap the router with CORS handler
	handler := c.Handler(r)
	
	// Start HTTP server
	port := db.GetEnv("PORT", "8080")
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
