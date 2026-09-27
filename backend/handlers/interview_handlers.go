package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

var interviewStatuses = map[string]bool{
	"scheduled":   true,
	"completed":   true,
	"cancelled":   true,
	"rescheduled": true,
	"no_show":     true,
}

func interviewTenant(r *http.Request) string {
	return r.Context().Value("tenantID").(string)
}

func validateInterviewReferences(candidateID int, requirementID int, tenantID string) (string, string, string, error) {
	var candidateName, jobID, requirementTitle string
	err := db.DB.QueryRow(
		`SELECT c.name, r.job_id, r.title
		 FROM candidates c
		 JOIN requirements r ON r.tenant_id = c.tenant_id
		 WHERE c.id = $1 AND r.id = $2
		   AND c.tenant_id = $3 AND r.tenant_id = $3`,
		candidateID, requirementID, tenantID,
	).Scan(&candidateName, &jobID, &requirementTitle)
	if err != nil {
		return "", "", "", err
	}
	if strings.TrimSpace(jobID) == "" {
		return "", "", "", errMissingJobID{}
	}
	return candidateName, jobID, requirementTitle, nil
}

type errMissingJobID struct{}

func (errMissingJobID) Error() string { return "selected requirement does not have a Job ID" }

func validateInterviewStatus(status string) bool {
	return interviewStatuses[status]
}

// GetInterviews retrieves all current interview records for the authenticated
// tenant. Historical interview rows are retained; the endpoint does not
// expose deletion as part of the Phase 5 workflow.
func GetInterviews(w http.ResponseWriter, r *http.Request) {
	tenantID := interviewTenant(r)

	rows, err := db.DB.Query(`
		SELECT i.id, i.candidate_id, i.candidate_name, i.requirement_id,
			COALESCE(r.job_id, ''), COALESCE(r.title, ''), i.position,
			i.round, i.interview_date, i.duration_minutes, i.status, i.outcome,
			i.feedback, i.candidate_feedback, i.next_action, i.created_at,
			i.last_modified, i.tenant_id, i.company_name
		FROM interviews i
		LEFT JOIN requirements r ON r.id = i.requirement_id AND r.tenant_id = i.tenant_id
		WHERE i.tenant_id = $1
		ORDER BY i.interview_date DESC, i.id DESC`, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching interviews")
		return
	}
	defer rows.Close()

	interviews := []models.Interview{}
	for rows.Next() {
		var i models.Interview
		if err := rows.Scan(
			&i.ID, &i.CandidateID, &i.CandidateName, &i.RequirementID,
			&i.JobID, &i.RequirementTitle, &i.Position, &i.Round,
			&i.InterviewDate, &i.DurationMinutes, &i.Status, &i.Outcome,
			&i.Feedback, &i.CandidateFeedback, &i.NextAction, &i.CreatedAt,
			&i.LastModified, &i.TenantID, &i.CompanyName,
		); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning interview row")
			return
		}
		interviews = append(interviews, i)
	}

	if err := rows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading interview rows")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Interviews retrieved successfully",
		Data:    interviews,
	})
}

// GetInterviewByID retrieves one interview, always scoped to the authenticated
// tenant. Cross-tenant IDs are indistinguishable from nonexistent IDs.
func GetInterviewByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}

	tenantID := interviewTenant(r)
	var interview models.Interview
	err = db.DB.QueryRow(`
		SELECT i.id, i.candidate_id, i.candidate_name, i.requirement_id,
			COALESCE(r.job_id, ''), COALESCE(r.title, ''), i.position,
			i.round, i.interview_date, i.duration_minutes, i.status, i.outcome,
			i.feedback, i.candidate_feedback, i.next_action, i.created_at,
			i.last_modified, i.tenant_id, i.company_name
		FROM interviews i
		LEFT JOIN requirements r ON r.id = i.requirement_id AND r.tenant_id = i.tenant_id
		WHERE i.id = $1 AND i.tenant_id = $2`,
		id, tenantID,
	).Scan(
		&interview.ID, &interview.CandidateID, &interview.CandidateName,
		&interview.RequirementID, &interview.JobID, &interview.RequirementTitle,
		&interview.Position, &interview.Round, &interview.InterviewDate,
		&interview.DurationMinutes, &interview.Status, &interview.Outcome,
		&interview.Feedback, &interview.CandidateFeedback, &interview.NextAction,
		&interview.CreatedAt, &interview.LastModified, &interview.TenantID,
		&interview.CompanyName,
	)
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

// GetAssignmentInterviews returns the preserved interview history for one
// Candidate × Requirement recruitment transaction.
func GetAssignmentInterviews(w http.ResponseWriter, r *http.Request) {
	assignmentID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid assignment ID")
		return
	}
	tenantID := interviewTenant(r)

	var candidateID, requirementID int
	if err := db.DB.QueryRow(
		`SELECT candidate_id, requirement_id
		 FROM recruitment_assignments
		 WHERE id = $1 AND tenant_id = $2`,
		assignmentID, tenantID,
	).Scan(&candidateID, &requirementID); err != nil {
		respondWithError(w, http.StatusNotFound, "Assignment not found")
		return
	}

	rows, err := db.DB.Query(`
		SELECT i.id, i.candidate_id, i.candidate_name, i.requirement_id,
			COALESCE(r.job_id, ''), COALESCE(r.title, ''), i.position,
			i.round, i.interview_date, i.duration_minutes, i.status, i.outcome,
			i.feedback, i.candidate_feedback, i.next_action, i.created_at,
			i.last_modified, i.tenant_id, i.company_name
		FROM interviews i
		LEFT JOIN requirements r ON r.id = i.requirement_id AND r.tenant_id = i.tenant_id
		WHERE i.candidate_id = $1 AND i.requirement_id = $2 AND i.tenant_id = $3
		ORDER BY i.round ASC, i.interview_date DESC, i.id DESC`,
		candidateID, requirementID, tenantID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching interview history")
		return
	}
	defer rows.Close()

	history := []models.Interview{}
	for rows.Next() {
		var i models.Interview
		if err := rows.Scan(
			&i.ID, &i.CandidateID, &i.CandidateName, &i.RequirementID,
			&i.JobID, &i.RequirementTitle, &i.Position, &i.Round,
			&i.InterviewDate, &i.DurationMinutes, &i.Status, &i.Outcome,
			&i.Feedback, &i.CandidateFeedback, &i.NextAction, &i.CreatedAt,
			&i.LastModified, &i.TenantID, &i.CompanyName,
		); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning interview history")
			return
		}
		history = append(history, i)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Interview history retrieved successfully",
		Data:    history,
	})
}

// ScheduleInterview creates an interview only for an existing Candidate ×
// Requirement assignment. The assignment must already be submitted or
// interviewing. Creating the first interview advances submitted -> interviewing.
func ScheduleInterview(w http.ResponseWriter, r *http.Request) {
	var interview models.Interview
	if err := json.NewDecoder(r.Body).Decode(&interview); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	tenantID := interviewTenant(r)
	interview.TenantID = tenantID
	interview.CompanyName = r.Context().Value("companyName").(string)

	if interview.CandidateID <= 0 || interview.RequirementID == nil || *interview.RequirementID <= 0 {
		respondWithError(w, http.StatusBadRequest, "candidateId and requirementId are required")
		return
	}
	if interview.Round <= 0 {
		interview.Round = 1
	}
	if interview.InterviewDate.IsZero() {
		respondWithError(w, http.StatusBadRequest, "interviewDate is required")
		return
	}
	if interview.Status == "" {
		interview.Status = "scheduled"
	}
	if !validateInterviewStatus(interview.Status) {
		respondWithError(w, http.StatusBadRequest, "Invalid interview status")
		return
	}

	candidateName, jobID, requirementTitle, err := validateInterviewReferences(
		interview.CandidateID, *interview.RequirementID, tenantID,
	)
	if err != nil {
		if _, ok := err.(errMissingJobID); ok {
			respondWithError(w, http.StatusUnprocessableEntity, "Selected requirement does not have a Job ID")
			return
		}
		respondWithError(w, http.StatusNotFound, "Candidate or requirement not found")
		return
	}

	var assignmentID int
	var assignmentStatus string
	if err := db.DB.QueryRow(
		`SELECT id, status
		 FROM recruitment_assignments
		 WHERE candidate_id = $1 AND requirement_id = $2 AND tenant_id = $3`,
		interview.CandidateID, *interview.RequirementID, tenantID,
	).Scan(&assignmentID, &assignmentStatus); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Candidate is not assigned to the selected requirement")
		return
	}
	if assignmentStatus != "submitted" && assignmentStatus != "interviewing" {
		respondWithError(w, http.StatusUnprocessableEntity, "Assignment must be submitted before scheduling an interview")
		return
	}

	interview.CandidateName = candidateName
	interview.JobID = jobID
	interview.RequirementTitle = requirementTitle
	interview.Position = requirementTitle

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error starting interview transaction")
		return
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`INSERT INTO interviews (
			candidate_id, candidate_name, requirement_id, position, round,
			interview_date, duration_minutes, status, outcome, feedback,
			candidate_feedback, next_action, tenant_id, company_name
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, last_modified`,
		interview.CandidateID, interview.CandidateName, *interview.RequirementID,
		interview.Position, interview.Round, interview.InterviewDate,
		interview.DurationMinutes, interview.Status, interview.Outcome,
		interview.Feedback, interview.CandidateFeedback, interview.NextAction,
		tenantID, interview.CompanyName,
	).Scan(&interview.ID, &interview.CreatedAt, &interview.LastModified)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error scheduling interview")
		return
	}

	if assignmentStatus == "submitted" {
		if _, err := tx.Exec(
			`UPDATE recruitment_assignments
			 SET status = 'interviewing', last_modified = NOW()
			 WHERE id = $1 AND tenant_id = $2`,
			assignmentID, tenantID,
		); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error advancing recruitment assignment")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error committing interview")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Interview scheduled successfully",
		Data:    interview,
	})
}

// UpdateInterview updates the interview event while preserving its identity
// and Candidate × Requirement relationship.
func UpdateInterview(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}

	var interview models.Interview
	if err := json.NewDecoder(r.Body).Decode(&interview); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	tenantID := interviewTenant(r)
	if interview.Round <= 0 {
		interview.Round = 1
	}
	if interview.InterviewDate.IsZero() {
		respondWithError(w, http.StatusBadRequest, "interviewDate is required")
		return
	}
	if interview.Status == "" {
		interview.Status = "scheduled"
	}
	if !validateInterviewStatus(interview.Status) {
		respondWithError(w, http.StatusBadRequest, "Invalid interview status")
		return
	}
	if interview.RequirementID == nil || *interview.RequirementID <= 0 || interview.CandidateID <= 0 {
		respondWithError(w, http.StatusBadRequest, "candidateId and requirementId are required")
		return
	}

	candidateName, jobID, requirementTitle, err := validateInterviewReferences(
		interview.CandidateID, *interview.RequirementID, tenantID,
	)
	if err != nil {
		if _, ok := err.(errMissingJobID); ok {
			respondWithError(w, http.StatusUnprocessableEntity, "Selected requirement does not have a Job ID")
			return
		}
		respondWithError(w, http.StatusNotFound, "Candidate or requirement not found")
		return
	}

	interview.ID = id
	interview.TenantID = tenantID
	interview.CandidateName = candidateName
	interview.JobID = jobID
	interview.RequirementTitle = requirementTitle
	interview.Position = requirementTitle

	result, err := db.DB.Exec(
		`UPDATE interviews SET
			candidate_id = $1, candidate_name = $2, requirement_id = $3,
			position = $4, round = $5, interview_date = $6,
			duration_minutes = $7, status = $8, outcome = $9,
			feedback = $10, candidate_feedback = $11, next_action = $12,
			last_modified = NOW()
		WHERE id = $13 AND tenant_id = $14`,
		interview.CandidateID, interview.CandidateName, *interview.RequirementID,
		interview.Position, interview.Round, interview.InterviewDate,
		interview.DurationMinutes, interview.Status, interview.Outcome,
		interview.Feedback, interview.CandidateFeedback, interview.NextAction,
		id, tenantID,
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
