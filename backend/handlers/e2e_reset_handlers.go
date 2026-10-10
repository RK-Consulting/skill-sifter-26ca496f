package handlers

import (
	"net/http"

	"github.com/RK-Consulting/skill-sifter/db"
)

const e2eSmokeTenantID = "e2e_smoke_tenant"

// ResetE2ESmokeTenantData is a narrowly scoped maintenance operation for the
// dedicated production smoke tenant. The route is registered behind normal
// JWT authentication and the admin role; this handler independently enforces
// the fixed tenant boundary as defense in depth.
func ResetE2ESmokeTenantData(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID != e2eSmokeTenantID {
		respondWithError(w, http.StatusForbidden, "This operation is restricted to the dedicated smoke tenant")
		return
	}
	role, ok := r.Context().Value("role").(string)
	if !ok || role != "admin" {
		respondWithError(w, http.StatusForbidden, "Tenant administrator role required")
		return
	}

	tenantDB := db.RequestDB(r)
	if tenantDB == nil {
		respondWithError(w, http.StatusServiceUnavailable, "Smoke tenant database is not ready")
		return
	}

	if _, err := tenantDB.Exec(`
		TRUNCATE TABLE
			clients,
			requirements,
			candidates,
			candidate_language_expertise,
			candidate_expertise,
			resumes,
			resume_search_logs,
			candidate_professional_profiles,
			candidate_employment_history,
			candidate_education,
			candidate_certifications,
			candidate_projects,
			interviews,
			recruitment_screenings,
			recruitment_submissions,
			recruitment_submission_feedback,
			recruitment_selections,
			recruitment_offers,
			recruitment_joinings,
			recruitment_billings,
			audit_events
		RESTART IDENTITY CASCADE
	`); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not reset E2E smoke tenant data")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "E2E smoke tenant data reset",
		"data": map[string]interface{}{
			"tenantId": e2eSmokeTenantID,
			"status":   "RESET",
		},
	})
}
