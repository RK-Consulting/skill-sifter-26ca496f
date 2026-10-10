package handlers

import (
	"net/http"
	"strings"

	"github.com/RK-Consulting/skill-sifter/db"
	"golang.org/x/crypto/bcrypt"
)

const (
	e2eSmokeTenantID = "e2e_smoke_tenant"
	e2eSmokeEmail    = "e2e-admin@skillsifter.in"
)

// // ResetE2ESmokeTenantData removes all tenant business data from the dedicated
// production smoke tenant while preserving its roles and administrator account.
// It is authenticated with the same CI-only E2E administrator credentials used
// by the production smoke workflow.
func ResetE2ESmokeTenantData(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Email != e2eSmokeEmail || len(input.Password) < 6 {
		respondWithError(w, http.StatusUnauthorized, "E2E smoke credentials are invalid")
		return
	}

	tenantDB, err := db.OpenDatabase(db.TenantDatabaseName(e2eSmokeTenantID))
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Could not open E2E smoke tenant database")
		return
	}
	defer tenantDB.Close()

	var passwordHash string
	if err := tenantDB.QueryRow(`
		SELECT password
		FROM users
		WHERE lower(email)=lower($1) AND tenant_id=$2 AND role='admin'
		LIMIT 1
	`, input.Email, e2eSmokeTenantID).Scan(&passwordHash); err != nil {
		respondWithError(w, http.StatusUnauthorized, "E2E smoke credentials are invalid")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "E2E smoke credentials are invalid")
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
