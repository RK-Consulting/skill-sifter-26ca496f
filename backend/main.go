package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/RK-Consulting/skill-sifter/auth"
	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
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
func setupCORS() *cors.Cors {
	return cors.New(cors.Options{AllowedOrigins: []string{"https://skillsifter.in", "https://www.skillsifter.in", "https://api.skillsifter.in", "https://*.skill-sifter-26ca496f.pages.dev", "http://localhost:5173", "http://localhost:3000", "http://127.0.0.1:5173", "http://127.0.0.1:3000"}, AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"}, AllowedHeaders: []string{"Content-Type", "Authorization", "Origin", "Accept", "X-Requested-With", "X-CSRF-Token"}, ExposedHeaders: []string{"Content-Length", "Content-Type"}, AllowCredentials: true, MaxAge: 86400})
}
func setupPublicRoutes(r *mux.Router) {
	// The API is canonical under /api. Keep the site root separate from API routes.
	r.HandleFunc("/", rootHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api", apiRootHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/health-check", healthCheckHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/ping", pingHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/auth/register", handlers.StartRegistration).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/login", handlers.LoginUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/e2e/bootstrap", handlers.BootstrapE2ESmokeAccount).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/e2e/reset", handlers.ResetE2ESmokeTenantData).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/account/plans", handlers.GetSubscriptionPlans).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/auth/register/verify-email", handlers.VerifyRegistrationEmail).Methods("POST", "OPTIONS")
}
func setupResourceRoutes(router *mux.Router, path string, getAll, create, getOne, update, del http.HandlerFunc) {
	router.HandleFunc(path, getAll).Methods("GET", "OPTIONS")
	router.HandleFunc(path, create).Methods("POST", "OPTIONS")
	router.HandleFunc(path+"/{id}", getOne).Methods("GET", "OPTIONS")
	router.HandleFunc(path+"/{id}", update).Methods("PUT", "OPTIONS")
	router.HandleFunc(path+"/{id}", del).Methods("DELETE", "OPTIONS")
}
func managerOnly(h http.HandlerFunc) http.HandlerFunc {
	return auth.RoleMiddleware("admin", "manager")(h).ServeHTTP
}
func setupProtectedRoutes(r *mux.Router) {
	api := r.PathPrefix("/api").Subrouter()
	api.Use(auth.AuthMiddleware)
	api.Use(auth.TenantDBMiddleware)
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(auth.RoleMiddleware("admin"))
	admin.HandleFunc("/users", handlers.GetUsers).Methods("GET", "OPTIONS")
	admin.HandleFunc("/users", handlers.CreateUser).Methods("POST", "OPTIONS")
	admin.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE", "OPTIONS")
	manager := api.PathPrefix("/manager").Subrouter()
	manager.Use(auth.RoleMiddleware("manager", "admin"))
	manager.HandleFunc("/users", handlers.GetUsers).Methods("GET", "OPTIONS")
	api.HandleFunc("/account", handlers.GetCurrentAccount).Methods("GET", "OPTIONS")
	api.HandleFunc("/account/subscription", handlers.GetSubscriptionAccount).Methods("GET", "OPTIONS")
	api.HandleFunc("/account/subscription/phone/send", auth.RoleMiddleware("admin")(http.HandlerFunc(handlers.SendPhoneVerificationCode)).ServeHTTP).Methods("POST", "OPTIONS")
	api.HandleFunc("/account/subscription/phone/verify", auth.RoleMiddleware("admin")(http.HandlerFunc(handlers.VerifyPhoneVerificationCode)).ServeHTTP).Methods("POST", "OPTIONS")
	api.HandleFunc("/account/subscription/checkout", auth.RoleMiddleware("admin")(http.HandlerFunc(handlers.StartSubscriptionCheckout)).ServeHTTP).Methods("POST", "OPTIONS")
	api.HandleFunc("/account/subscription/cancel", auth.RoleMiddleware("admin")(http.HandlerFunc(handlers.CancelSubscription)).ServeHTTP).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/subscriptions/webhook/razorpay", handlers.RazorpaySubscriptionWebhook).Methods("POST", "OPTIONS")
	api.HandleFunc("/admin/tenant/provision", auth.RoleMiddleware("admin")(http.HandlerFunc(handlers.ProvisionCurrentTenant)).ServeHTTP).Methods("POST", "OPTIONS")
	api.HandleFunc("/candidates", handlers.GetCandidates).Methods("GET", "OPTIONS")
	api.HandleFunc("/candidates", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.AddCandidate)).ServeHTTP).Methods("POST", "OPTIONS")
	api.HandleFunc("/candidates/{id}", handlers.GetCandidateByID).Methods("GET", "OPTIONS")
	api.HandleFunc("/candidates/{id}", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.UpdateCandidate)).ServeHTTP).Methods("PUT", "OPTIONS")
	api.HandleFunc("/candidates/{id}", auth.RoleMiddleware("admin", "manager")(http.HandlerFunc(handlers.DeleteCandidate)).ServeHTTP).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/candidates/{id}/resume", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.UploadCandidateResume)).ServeHTTP).Methods("POST", "OPTIONS")
	api.HandleFunc("/candidates/{id}/resume", handlers.GetCandidateResume).Methods("GET", "OPTIONS")

	// Client and Requirement are the authoritative V1 recruitment-demand domain.
	// Requirements replace the legacy Jobs resource.
	apiV1 := r.PathPrefix("/api/v1").Subrouter()
	apiV1.Use(auth.AuthMiddleware)
	apiV1.Use(auth.TenantDBMiddleware)

	// Phase 5: agency-first interview workflow. Interview history is preserved;
	// deletion is intentionally not exposed as a core workflow operation.
	// Keep the legacy routes for existing UI compatibility while exposing the
	// authoritative V1 API under /api/v1.
	api.HandleFunc("/interviews", handlers.GetInterviews).Methods("GET", "OPTIONS")
	api.HandleFunc("/interviews", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.ScheduleInterview)).ServeHTTP).Methods("POST", "OPTIONS")
	api.HandleFunc("/interviews/{id}", handlers.GetInterviewByID).Methods("GET", "OPTIONS")
	api.HandleFunc("/interviews/{id}", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.UpdateInterview)).ServeHTTP).Methods("PUT", "OPTIONS")
	apiV1.HandleFunc("/interviews", handlers.GetInterviews).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/interviews", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.ScheduleInterview)).ServeHTTP).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/interviews/{id}", handlers.GetInterviewByID).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/interviews/{id}", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.UpdateInterview)).ServeHTTP).Methods("PUT", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/interviews", handlers.GetCandidateInterviews).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/screenings", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.CreateCandidateScreening)).ServeHTTP).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/screenings", handlers.GetCandidateScreenings).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/screenings/{screeningId}", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.UpdateCandidateScreening)).ServeHTTP).Methods("PUT", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/selection", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.CreateCandidateRequirementSelection)).ServeHTTP).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/selection", handlers.GetCandidateRequirementSelection).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/offer", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.CreateCandidateRequirementOffer)).ServeHTTP).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/offer", handlers.GetCandidateRequirementOffer).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/offer", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.UpdateCandidateRequirementOffer)).ServeHTTP).Methods("PUT", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/joining", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.CreateCandidateRequirementJoining)).ServeHTTP).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/joining", handlers.GetCandidateRequirementJoining).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/joining", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.UpdateCandidateRequirementJoining)).ServeHTTP).Methods("PUT", "OPTIONS")
	apiV1.HandleFunc("/billing", auth.RoleMiddleware("admin", "manager", "team_leader")(http.HandlerFunc(handlers.GetBillingWorklist)).ServeHTTP).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/billing", auth.RoleMiddleware("admin", "manager")(http.HandlerFunc(handlers.CreateCandidateRequirementBilling)).ServeHTTP).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/billing", handlers.GetCandidateRequirementBilling).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/submissions", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.AddCandidateRequirementSubmission)).ServeHTTP).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/candidates/{candidateId}/requirements/{requirementId}/submissions", handlers.GetCandidateRequirementSubmissions).Methods("GET", "OPTIONS")
	api.HandleFunc("/reports/hiring", handlers.GetHiringReport).Methods("GET", "OPTIONS")
	api.HandleFunc("/reports/pipeline", handlers.GetPipelineReport).Methods("GET", "OPTIONS")
	api.HandleFunc("/reports/sources", handlers.GetSourceReport).Methods("GET", "OPTIONS")
	api.HandleFunc("/reports/activity", handlers.GetRecentActivity).Methods("GET", "OPTIONS")
	api.HandleFunc("/reports/periodic", handlers.GetPeriodicReport).Methods("GET", "OPTIONS")
	api.HandleFunc("/reports/activity-log", handlers.GetActivityLog).Methods("GET", "OPTIONS")
	api.HandleFunc("/resume-ai/upload", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.UploadResumes)).ServeHTTP).Methods("POST", "OPTIONS")
	api.HandleFunc("/resume-ai/search", handlers.SearchResumes).Methods("GET", "OPTIONS")
	api.HandleFunc("/resume-ai/resumes", handlers.ListResumes).Methods("GET", "OPTIONS")
	api.HandleFunc("/resume-ai/resumes/{id}", handlers.GetResumeDetail).Methods("GET", "OPTIONS")
	api.HandleFunc("/resume-ai/resumes/{id}/file", handlers.DownloadResume).Methods("GET", "OPTIONS")
	api.HandleFunc("/resume-ai/resumes/{id}/retry", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.RetryResume)).ServeHTTP).Methods("POST", "OPTIONS")
	api.HandleFunc("/resume-ai/health", handlers.GetResumeHealth).Methods("GET", "OPTIONS")

	apiV1.HandleFunc("/clients", handlers.GetClients).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/clients", auth.RoleMiddleware("admin", "manager", "team_leader", "recruiter")(http.HandlerFunc(handlers.AddClient)).ServeHTTP).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/clients/{id}", handlers.GetClientByID).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/clients/{id}", auth.RoleMiddleware("admin", "manager", "team_leader", "recruiter")(http.HandlerFunc(handlers.UpdateClient)).ServeHTTP).Methods("PUT", "OPTIONS")
	apiV1.HandleFunc("/clients/{id}", managerOnly(handlers.DeleteClient)).Methods("DELETE", "OPTIONS")
	apiV1.HandleFunc("/requirements", handlers.GetRequirements).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/requirements", auth.RoleMiddleware("admin", "manager", "team_leader", "recruiter")(http.HandlerFunc(handlers.AddRequirement)).ServeHTTP).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/requirements/{id}", handlers.GetRequirementByID).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/requirements/{id}", auth.RoleMiddleware("admin", "manager", "team_leader", "recruiter")(http.HandlerFunc(handlers.UpdateRequirement)).ServeHTTP).Methods("PUT", "OPTIONS")
	apiV1.HandleFunc("/requirements/{id}", managerOnly(handlers.DeleteRequirement)).Methods("DELETE", "OPTIONS")
	apiV1.HandleFunc("/requirements/{id}/matches", handlers.GetRequirementMatches).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/requirements/{id}/matches/{candidateId}", handlers.GetRequirementCandidateMatch).Methods("GET", "OPTIONS")

	apiV1.HandleFunc("/submissions/{submissionId}/feedback", auth.RoleMiddleware("admin", "manager", "recruiter", "team_leader")(http.HandlerFunc(handlers.AddSubmissionFeedback)).ServeHTTP).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/submissions/{submissionId}/feedback", handlers.GetSubmissionFeedback).Methods("GET", "OPTIONS")
}
func main() {
	db.InitDB()
	defer db.DB.Close()
	defer db.CloseTenantDatabases()
	if err := db.InitializeSchema(); err != nil {
		log.Fatalf("Schema initialization failed: %v", err)
	}
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	setupPublicRoutes(r)
	setupProtectedRoutes(r)
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			handlers.CleanupExpiredTrials()
		}
	}()
	handlers.CleanupExpiredTrials()
	port := db.GetEnv("PORT", "8080")
	fmt.Printf("SkillSifter API running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, setupCORS().Handler(r)))
}
