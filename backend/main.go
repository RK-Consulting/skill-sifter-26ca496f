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

func loggingMiddleware(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){log.Printf("%s %s %s",r.RemoteAddr,r.Method,r.URL.Path);next.ServeHTTP(w,r)}) }
func rootHandler(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","application/json");w.Write([]byte(`{"message":"Welcome to SkillSifter API"}`))}
func apiRootHandler(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","application/json");w.Write([]byte(`{"message":"Welcome to SkillSifter API"}`))}
func healthCheckHandler(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","application/json");w.Write([]byte(`{"status":"OK"}`))}
func pingHandler(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","application/json");w.Write([]byte(`{"message":"pong"}`))}
func setupCORS()*cors.Cors{return cors.New(cors.Options{AllowedOrigins:[]string{"https://skillsifter.in","https://www.skillsifter.in","https://api.skillsifter.in","https://*.skill-sifter-26ca496f.pages.dev","http://localhost:5173","http://localhost:3000","http://127.0.0.1:5173","http://127.0.0.1:3000"},AllowedMethods:[]string{"GET","POST","PUT","DELETE","OPTIONS","HEAD","PATCH"},AllowedHeaders:[]string{"Content-Type","Authorization","Origin","Accept","X-Requested-With","X-CSRF-Token"},ExposedHeaders:[]string{"Content-Length","Content-Type"},AllowCredentials:true,MaxAge:86400})}
func setupPublicRoutes(r *mux.Router){r.HandleFunc("/",rootHandler).Methods("GET","OPTIONS");r.HandleFunc("/api",apiRootHandler).Methods("GET","OPTIONS");r.HandleFunc("/health-check",healthCheckHandler).Methods("GET","OPTIONS");r.HandleFunc("/api/health-check",healthCheckHandler).Methods("GET","OPTIONS");r.HandleFunc("/ping",pingHandler).Methods("GET","OPTIONS");r.HandleFunc("/api/ping",pingHandler).Methods("GET","OPTIONS");r.HandleFunc("/auth/register",handlers.RegisterUser).Methods("POST","OPTIONS");r.HandleFunc("/api/auth/register",handlers.RegisterUser).Methods("POST","OPTIONS");r.HandleFunc("/auth/login",handlers.LoginUser).Methods("POST","OPTIONS");r.HandleFunc("/api/auth/login",handlers.LoginUser).Methods("POST","OPTIONS")}
func setupResourceRoutes(router *mux.Router,path string,getAll,create,getOne,update,del http.HandlerFunc){router.HandleFunc(path,getAll).Methods("GET","OPTIONS");router.HandleFunc(path,create).Methods("POST","OPTIONS");router.HandleFunc(path+"/{id}",getOne).Methods("GET","OPTIONS");router.HandleFunc(path+"/{id}",update).Methods("PUT","OPTIONS");router.HandleFunc(path+"/{id}",del).Methods("DELETE","OPTIONS")}
func managerOnly(h http.HandlerFunc)http.HandlerFunc{return auth.RoleMiddleware("admin","manager")(h).ServeHTTP}
func setupProtectedRoutes(r *mux.Router){
 api:=r.PathPrefix("/api").Subrouter();api.Use(auth.AuthMiddleware)
 admin:=api.PathPrefix("/admin").Subrouter();admin.Use(auth.RoleMiddleware("admin"));admin.HandleFunc("/users",handlers.GetUsers).Methods("GET","OPTIONS");admin.HandleFunc("/users",handlers.CreateUser).Methods("POST","OPTIONS");admin.HandleFunc("/users/{id}",handlers.UpdateUser).Methods("PUT","OPTIONS");admin.HandleFunc("/users/{id}",handlers.DeleteUser).Methods("DELETE","OPTIONS")
 manager:=api.PathPrefix("/manager").Subrouter();manager.Use(auth.RoleMiddleware("manager","admin"));manager.HandleFunc("/users",handlers.GetUsers).Methods("GET","OPTIONS");manager.HandleFunc("/users/{id}",handlers.UpdateUser).Methods("PUT","OPTIONS");manager.HandleFunc("/users/{id}",handlers.DeleteUser).Methods("DELETE","OPTIONS")
 api.HandleFunc("/company-users",handlers.GetUsers).Methods("GET","OPTIONS")
 setupResourceRoutes(api,"/candidates",handlers.GetCandidates,handlers.AddCandidate,handlers.GetCandidateByID,handlers.UpdateCandidate,handlers.DeleteCandidate)
 setupResourceRoutes(api,"/jobs",handlers.GetJobs,managerOnly(handlers.AddJob),handlers.GetJobByID,managerOnly(handlers.UpdateJob),managerOnly(handlers.DeleteJob))
 setupResourceRoutes(api,"/daily-jobs",handlers.GetDailyJobs,handlers.AddDailyJob,handlers.GetDailyJobByID,handlers.UpdateDailyJob,handlers.DeleteDailyJob)
 setupResourceRoutes(api,"/interviews",handlers.GetInterviews,handlers.ScheduleInterview,handlers.GetInterviewByID,handlers.UpdateInterview,handlers.DeleteInterview)
 setupResourceRoutes(api,"/business-dev",handlers.GetBusinessDevs,handlers.AddBusinessDev,handlers.GetBusinessDevByID,handlers.UpdateBusinessDev,handlers.DeleteBusinessDev)
 api.HandleFunc("/reports/hiring",handlers.GetHiringReport).Methods("GET","OPTIONS");api.HandleFunc("/reports/sources",handlers.GetSourceReport).Methods("GET","OPTIONS");api.HandleFunc("/reports/activity",handlers.GetRecentActivity).Methods("GET","OPTIONS");api.HandleFunc("/reports/periodic",handlers.GetPeriodicReport).Methods("GET","OPTIONS");api.HandleFunc("/reports/activity-log",handlers.GetActivityLog).Methods("GET","OPTIONS")
 api.HandleFunc("/resume-ai/upload",handlers.UploadResumes).Methods("POST","OPTIONS");api.HandleFunc("/resume-ai/search",handlers.SearchResumes).Methods("GET","OPTIONS");api.HandleFunc("/resume-ai/resumes",handlers.ListResumes).Methods("GET","OPTIONS");api.HandleFunc("/resume-ai/health",handlers.GetResumeHealth).Methods("GET","OPTIONS")

 // Issue #34 / ADR 0002: Client and Requirement domain. New V1 domain work
 // is introduced under /api/v1 per ADR 0008 ("New V1 endpoints must be
 // introduced under /api/v1/..."); the existing /api namespace above is
 // untouched, and `jobs` remains available there unchanged (ADR 0002:
 // jobs is retained as a temporary compatibility model, not replaced).
 apiV1:=r.PathPrefix("/api/v1").Subrouter();apiV1.Use(auth.AuthMiddleware)
 setupResourceRoutes(apiV1,"/clients",handlers.GetClients,managerOnly(handlers.AddClient),handlers.GetClientByID,managerOnly(handlers.UpdateClient),managerOnly(handlers.DeleteClient))
 setupResourceRoutes(apiV1,"/requirements",handlers.GetRequirements,managerOnly(handlers.AddRequirement),handlers.GetRequirementByID,managerOnly(handlers.UpdateRequirement),managerOnly(handlers.DeleteRequirement))

 // Issue #35 / ADR 0003 checkpoint 3: HTTP API only. Lifecycle transition
 // endpoints, snapshot capture, and audit-event writing are explicitly
 // deferred to later checkpoints — UpdateAssignment here only supports
 // reassigning owner_user_id (see handlers/assignment_handlers.go).
 setupResourceRoutes(apiV1,"/assignments",handlers.GetAssignments,managerOnly(handlers.AddAssignment),handlers.GetAssignmentByID,managerOnly(handlers.UpdateAssignment),managerOnly(handlers.DeleteAssignment))
 // Checkpoint 4: dedicated lifecycle-transition endpoint, deliberately
 // separate from PUT /assignments/{id} so owner mutation and lifecycle
 // transition stay two distinct concepts. Snapshot capture and audit
 // events remain deferred.
 apiV1.HandleFunc("/assignments/{id}/transition",managerOnly(handlers.TransitionAssignment)).Methods("POST","OPTIONS")
}
func main(){db.InitDB();defer db.DB.Close();if err:=db.ApplyMigrations();err!=nil{log.Fatalf("Migration failed: %v",err)};r:=mux.NewRouter();r.Use(loggingMiddleware);setupPublicRoutes(r);setupProtectedRoutes(r);r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(http.StatusNoContent)});port:=db.GetEnv("PORT","8080");fmt.Printf("SkillSifter API running at http://localhost:%s\n",port);log.Fatal(http.ListenAndServe(":"+port,setupCORS().Handler(r)))}
