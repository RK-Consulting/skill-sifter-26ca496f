
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

// ----------- Middleware -----------

func loggingMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
                next.ServeHTTP(w, r)
        })
}

// ----------- Basic Handlers -----------

func rootHandler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"message":"Welcome to SkillSifter API"}`))
}

func apiRootHandler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"message":"Welcome to SkillSifter API"}`))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"OK"}`))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"message":"pong"}`))
}

// ----------- CORS Setup -----------

func setupCORS() *cors.Cors {
        corsOptions := cors.Options{
                AllowedOrigins: []string{
                        "https://skillsifter.in",
                        "https://www.skillsifter.in",
                        "https://api.skillsifter.in",
                        "http://localhost:5173",
                        "http://localhost:3000",
                        "http://127.0.0.1:5173",
                        "http://127.0.0.1:3000",
                        // "*", // For development only
                },
                AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
                AllowedHeaders:   []string{"Content-Type", "Authorization", "Origin", "Accept", "X-Requested-With", "X-CSRF-Token"},
                ExposedHeaders:   []string{"Content-Length", "Content-Type"},
                AllowCredentials: true,
                MaxAge:           86400,
        }
        return cors.New(corsOptions)
}

// ----------- Public Routes -----------

func setupPublicRoutes(r *mux.Router) {
        r.HandleFunc("/", rootHandler).Methods("GET", "OPTIONS")
        r.HandleFunc("/api", apiRootHandler).Methods("GET", "OPTIONS")

        r.HandleFunc("/health-check", healthCheckHandler).Methods("GET", "OPTIONS")
        r.HandleFunc("/api/health-check", healthCheckHandler).Methods("GET", "OPTIONS")

        r.HandleFunc("/ping", pingHandler).Methods("GET", "OPTIONS")
        r.HandleFunc("/api/ping", pingHandler).Methods("GET", "OPTIONS")

        r.HandleFunc("/auth/register", handlers.RegisterUser).Methods("POST", "OPTIONS")
        r.HandleFunc("/api/auth/register", handlers.RegisterUser).Methods("POST", "OPTIONS")
        r.HandleFunc("/auth/login", handlers.LoginUser).Methods("POST", "OPTIONS")
        r.HandleFunc("/api/auth/login", handlers.LoginUser).Methods("POST", "OPTIONS")
}

// ----------- Protected Routes -----------

func setupProtectedRoutes(r *mux.Router) {
        apiRouter := r.PathPrefix("/api").Subrouter()
        apiRouter.Use(auth.AuthMiddleware)

        // Admin Routes
        adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
        adminRouter.Use(auth.RoleMiddleware("admin"))
        adminRouter.HandleFunc("/users", handlers.GetUsers).Methods("GET", "OPTIONS")
        adminRouter.HandleFunc("/users", handlers.CreateUser).Methods("POST", "OPTIONS")
        adminRouter.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT", "OPTIONS")
        adminRouter.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE", "OPTIONS")

        // Manager Routes
        managerRouter := apiRouter.PathPrefix("/manager").Subrouter()
        managerRouter.Use(auth.RoleMiddleware("manager", "admin"))
        // Add manager-specific routes if needed

        // General API Resources
        setupResourceRoutes(apiRouter, "/candidates", handlers.GetCandidates, handlers.AddCandidate,
                handlers.GetCandidateByID, handlers.UpdateCandidate, handlers.DeleteCandidate)

        setupResourceRoutes(apiRouter, "/jobs", handlers.GetJobs, handlers.AddJob,
                handlers.GetJobByID, handlers.UpdateJob, handlers.DeleteJob)

        setupResourceRoutes(apiRouter, "/daily-jobs", handlers.GetDailyJobs, handlers.AddDailyJob,
                handlers.GetDailyJobByID, handlers.UpdateDailyJob, handlers.DeleteDailyJob)

        setupResourceRoutes(apiRouter, "/interviews", handlers.GetInterviews, handlers.ScheduleInterview,
                handlers.GetInterviewByID, handlers.UpdateInterview, handlers.DeleteInterview)
                
        // Add business-dev routes with the /api prefix
        setupResourceRoutes(apiRouter, "/business-dev", handlers.GetBusinessDevs, handlers.AddBusinessDev,
                handlers.GetBusinessDevByID, handlers.UpdateBusinessDev, handlers.DeleteBusinessDev)

        // Duplicate for root-level paths (no /api prefix)
        nonApi := r.NewRoute().Subrouter()
        nonApi.Use(auth.AuthMiddleware)

        setupResourceRoutes(nonApi, "/candidates", handlers.GetCandidates, handlers.AddCandidate,
                handlers.GetCandidateByID, handlers.UpdateCandidate, handlers.DeleteCandidate)

        setupResourceRoutes(nonApi, "/jobs", handlers.GetJobs, handlers.AddJob,
                handlers.GetJobByID, handlers.UpdateJob, handlers.DeleteJob)

        setupResourceRoutes(nonApi, "/daily-jobs", handlers.GetDailyJobs, handlers.AddDailyJob,
                handlers.GetDailyJobByID, handlers.UpdateDailyJob, handlers.DeleteDailyJob)

        setupResourceRoutes(nonApi, "/interviews", handlers.GetInterviews, handlers.ScheduleInterview,
                handlers.GetInterviewByID, handlers.UpdateInterview, handlers.DeleteInterview)
                
        // Add business-dev routes without the /api prefix
        setupResourceRoutes(nonApi, "/business-dev", handlers.GetBusinessDevs, handlers.AddBusinessDev,
                handlers.GetBusinessDevByID, handlers.UpdateBusinessDev, handlers.DeleteBusinessDev)
}

// ----------- Resource Router Helper -----------

func setupResourceRoutes(router *mux.Router, path string,
        getAll http.HandlerFunc, create http.HandlerFunc,
        getOne http.HandlerFunc, update http.HandlerFunc, delete http.HandlerFunc) {

        router.HandleFunc(path, getAll).Methods("GET", "OPTIONS")
        router.HandleFunc(path, create).Methods("POST", "OPTIONS")
        router.HandleFunc(path+"/{id}", getOne).Methods("GET", "OPTIONS")
        router.HandleFunc(path+"/{id}", update).Methods("PUT", "OPTIONS")
        router.HandleFunc(path+"/{id}", delete).Methods("DELETE", "OPTIONS")
}

// ----------- Main Entry Point -----------

func main() {
        // Init DB
        db.InitDB()
        defer db.DB.Close()

        if err := db.InitializeSchema(); err != nil {
                log.Fatalf("Schema initialization failed: %v", err)
        }

        // Setup Router
        r := mux.NewRouter()
        r.Use(loggingMiddleware)

        setupPublicRoutes(r)
        setupProtectedRoutes(r)

        // Global OPTIONS handler for all unmatched OPTIONS requests
        r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusNoContent)
        })

        // Apply CORS
        corsHandler := setupCORS().Handler(r)

        // Start Server
        port := db.GetEnv("PORT", "8080")
        fmt.Printf("✅ SkillSifter API running at http://localhost:%s\n", port)
        log.Fatal(http.ListenAndServe(":"+port, corsHandler))
}
