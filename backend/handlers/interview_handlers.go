package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetInterviews retrieves all interviews for the authenticated tenant.
// Scoped by tenant_id (ADR 0001), not company_name.
func GetInterviews(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)

	interviews := []models.Interview{}
	rows, err := db.DB.Query(`
		SELECT i.id, i.candidate_id, i.candidate_name, i.requirement_id,
			COALESCE(r.job_id, ''), COALESCE(r.title, ''), i.position, i.interview_date,
			 i.status, i.feedback, i.last_modified, i.tenant_id, i.company_name
		FROM interviews i
		LEFT JOIN requirements r ON r.id = i.requirement_id AND r.tenant_id = i.tenant_id
		WHERE i.tenant_id = $1`, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching interviews")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Interview
		err := rows.Scan(&i.ID, &i.CandidateID, &i.CandidateName, &i.RequirementID, &i.JobID, &i.RequirementTitle, &i.Position,
			&i.InterviewDate, &i.Status, &i.Feedback, &i.LastModified, &i.TenantID, &i.CompanyName)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning interview row")
			return
		}
		interviews = append(interviews, i)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Interviews retrieved successfully",
		Data:    interviews,
	})
}

// GetInterviewByID retrieves a single interview by ID, scoped to the
// authenticated tenant. An interview ID belonging to another tenant returns
// 404, identically to a nonexistent ID (ADR 0001).
func GetInterviewByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	var interview models.Interview
	err = db.DB.QueryRow(`
		SELECT i.id, i.candidate_id, i.candidate_name, i.requirement_id,
			COALESCE(r.job_id, ''), COALESCE(r.title, ''), i.position, i.interview_date,
			 i.status, i.feedback, i.last_modified, i.tenant_id, i.company_name
		FROM interviews i
		LEFT JOIN requirements r ON r.id = i.requirement_id AND r.tenant_id = i.tenant_id
		WHERE i.id = $1 AND i.tenant_id = $2`,
		id, tenantID,
	).Scan(&interview.ID, &interview.CandidateID, &interview.CandidateName, &interview.Position,
		&interview.InterviewDate, &interview.Status, &interview.Feedback,
		&interview.LastModified, &interview.TenantID, &interview.CompanyName)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Interview not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Interview retrieved successfully",
		Data:    interview,
	})
}

// ScheduleInterview creates a new interview under the authenticated tenant.
// tenant_id is always derived from context, never from the request payload.
func ScheduleInterview(w http.ResponseWriter, r *http.Request) {
	var interview models.Interview
	err := json.NewDecoder(r.Body).Decode(&interview)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	interview.TenantID = r.Context().Value("tenantID").(string)
	interview.CompanyName = r.Context().Value("companyName").(string)

	if interview.RequirementID == nil || *interview.RequirementID <= 0 {
		respondWithError(w, http.StatusBadRequest, "requirementId is required")
		return
	}

	var jobID, requirementTitle string
	err = db.DB.QueryRow(
		`SELECT job_id, title FROM requirements WHERE id = $1 AND tenant_id = $2`,
		*interview.RequirementID, interview.TenantID,
	).Scan(&jobID, &requirementTitle)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Requirement not found")
		return
	}
	if jobID == "" {
		respondWithError(w, http.StatusUnprocessableEntity, "Selected requirement does not have a Job ID")
		return
	}
	interview.JobID = jobID
	interview.RequirementTitle = requirementTitle
	interview.Position = requirementTitle

	var id int
	err = db.DB.QueryRow(
		`INSERT INTO interviews (candidate_id, candidate_name, requirement_id, position, interview_date, status, feedback, tenant_id, company_name) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id`,
		interview.CandidateID, interview.CandidateName, *interview.RequirementID, interview.Position,
		interview.InterviewDate, interview.Status, interview.Feedback,
		interview.TenantID, interview.CompanyName,
	).Scan(&id)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error scheduling interview")
		return
	}

	interview.ID = id

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Interview scheduled successfully",
		Data:    interview,
	})
}

// UpdateInterview updates an existing interview, scoped to the
// authenticated tenant. An interview ID belonging to another tenant affects
// zero rows.
func UpdateInterview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}

	var interview models.Interview
	err = json.NewDecoder(r.Body).Decode(&interview)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	tenantID := r.Context().Value("tenantID").(string)
	interview.TenantID = tenantID
	interview.ID = id

	if interview.RequirementID == nil || *interview.RequirementID <= 0 {
		respondWithError(w, http.StatusBadRequest, "requirementId is required")
		return
	}

	var jobID, requirementTitle string
	err = db.DB.QueryRow(
		`SELECT job_id, title FROM requirements WHERE id = $1 AND tenant_id = $2`,
		*interview.RequirementID, tenantID,
	).Scan(&jobID, &requirementTitle)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Requirement not found")
		return
	}
	if jobID == "" {
		respondWithError(w, http.StatusUnprocessableEntity, "Selected requirement does not have a Job ID")
		return
	}
	interview.JobID = jobID
	interview.RequirementTitle = requirementTitle
	interview.Position = requirementTitle

	result, err := db.DB.Exec(
		`UPDATE interviews 
		SET candidate_id = $1, candidate_name = $2, requirement_id = $3, position = $4, interview_date = $5, 
			status = $6, feedback = $7, last_modified = NOW() 
		WHERE id = $8 AND tenant_id = $9`,
		interview.CandidateID, interview.CandidateName, *interview.RequirementID, interview.Position,
		interview.InterviewDate, interview.Status, interview.Feedback,
		interview.ID, tenantID,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating interview")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Interview not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Interview updated successfully",
		Data:    interview,
	})
}

// DeleteInterview deletes an interview, scoped to the authenticated tenant.
func DeleteInterview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	result, err := db.DB.Exec(
		"DELETE FROM interviews WHERE id = $1 AND tenant_id = $2",
		id, tenantID,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting interview")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Interview not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Interview deleted successfully",
	})
}
