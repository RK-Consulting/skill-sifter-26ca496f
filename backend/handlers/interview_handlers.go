package handlers

import (
	"database/sql"
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

func validateInterviewStatus(status string) bool {
	return interviewStatuses[status]
}

type errMissingJobID struct{}

func (errMissingJobID) Error() string { return "selected requirement does not have a Job ID" }

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

func scanInterviewRows(rows *sql.Rows) ([]models.Interview, error) {
	result := []models.Interview{}
	for rows.Next() {
		var i models.Interview
		if err := rows.Scan(
			&i.ID, &i.CandidateID, &i.CandidateName, &i.RequirementID,
			&i.JobID, &i.RequirementTitle, &i.Position, &i.Round,
			&i.InterviewDate, &i.Status, &i.Outcome,
			&i.Feedback, &i.CandidateFeedback, &i.NextAction,
			&i.LastModified, &i.TenantID, &i.CompanyName,
		); err != nil {
			return nil, err
		}
		result = append(result, i)
	}
	return result, rows.Err()
}

func GetInterviews(w http.ResponseWriter, r *http.Request) {
	tenantID := interviewTenant(r)
	rows, err := db.DB.Query(`
		SELECT i.id, i.candidate_id, i.candidate_name, i.requirement_id,
			COALESCE(r.job_id, ''), COALESCE(r.title, ''), i.position,
			i.round, i.interview_date, i.status, i.outcome,
			i.feedback, i.candidate_feedback, i.next_action,
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

	interviews, err := scanInterviewRows(rows)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading interview rows")
		return
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true, Message: "Interviews retrieved successfully", Data: interviews,
	})
}

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
			i.round, i.interview_date, i.status, i.outcome,
			i.feedback, i.candidate_feedback, i.next_action,
			i.last_modified, i.tenant_id, i.company_name
		FROM interviews i
		LEFT JOIN requirements r ON r.id = i.requirement_id AND r.tenant_id = i.tenant_id
		WHERE i.id = $1 AND i.tenant_id = $2`, id, tenantID).Scan(
		&interview.ID, &interview.CandidateID, &interview.CandidateName,
		&interview.RequirementID, &interview.JobID, &interview.RequirementTitle,
		&interview.Position, &interview.Round, &interview.InterviewDate,
		&interview.Status, &interview.Outcome,
		&interview.Feedback, &interview.CandidateFeedback, &interview.NextAction,
		&interview.LastModified, &interview.TenantID, &interview.CompanyName,
	)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Interview not found")
		return
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true, Message: "Interview retrieved successfully", Data: interview,
	})
}

// GetCandidateInterviews returns interview history directly by candidate.
// Requirement IDs in each history row identify the client opportunity.
func GetCandidateInterviews(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}
	tenantID := interviewTenant(r)

	rows, err := db.DB.Query(`
		SELECT i.id, i.candidate_id, i.candidate_name, i.requirement_id,
			COALESCE(r.job_id, ''), COALESCE(r.title, ''), i.position,
			i.round, i.interview_date, i.status, i.outcome,
			i.feedback, i.candidate_feedback, i.next_action,
			i.last_modified, i.tenant_id, i.company_name
		FROM interviews i
		LEFT JOIN requirements r ON r.id = i.requirement_id AND r.tenant_id = i.tenant_id
		WHERE i.candidate_id = $1 AND i.tenant_id = $2
		ORDER BY i.interview_date DESC, i.id DESC`, candidateID, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching candidate interview history")
		return
	}
	defer rows.Close()

	history, err := scanInterviewRows(rows)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading interview history")
		return
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true, Message: "Candidate interview history retrieved successfully", Data: history,
	})
}

func ScheduleInterview(w http.ResponseWriter, r *http.Request) {
	var interview models.Interview
	if err := json.NewDecoder(r.Body).Decode(&interview); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	tenantID := interviewTenant(r)
	companyName, _ := r.Context().Value("companyName").(string)

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

	interview.CandidateName = candidateName
	interview.JobID = jobID
	interview.RequirementTitle = requirementTitle
	interview.Position = requirementTitle
	interview.TenantID = tenantID
	interview.CompanyName = companyName

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error starting interview transaction")
		return
	}
	defer tx.Rollback()

	// Candidate row is the single persistent concurrency gate. The conditional
	// UPDATE is atomic: only an unlocked candidate can acquire the interview lock.
	var lockedCandidateID int
	err = tx.QueryRow(`
		UPDATE candidates
		SET interview_locked = TRUE
		WHERE id = $1 AND tenant_id = $2 AND interview_locked = FALSE
		RETURNING id`, interview.CandidateID, tenantID).Scan(&lockedCandidateID)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusConflict, "Candidate is already locked by an active interview")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error acquiring interview lock")
		return
	}

	err = tx.QueryRow(`
		INSERT INTO interviews (
			candidate_id, candidate_name, requirement_id, position, round,
			interview_date, status, outcome, feedback,
			candidate_feedback, next_action, tenant_id, company_name
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, last_modified`,
		interview.CandidateID, interview.CandidateName, *interview.RequirementID,
		interview.Position, interview.Round, interview.InterviewDate,
		interview.Status, interview.Outcome, interview.Feedback,
		interview.CandidateFeedback, interview.NextAction, tenantID, companyName,
	).Scan(&interview.ID, &interview.LastModified)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error scheduling interview")
		return
	}

	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error committing interview")
		return
	}
	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true, Message: "Interview scheduled successfully", Data: interview,
	})
}

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

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error starting interview update")
		return
	}
	defer tx.Rollback()

	var existingCandidateID, existingRequirementID int
	var existingStatus, existingOutcome string
	if err := tx.QueryRow(`
		SELECT candidate_id, requirement_id, status, COALESCE(outcome, '')
		FROM interviews WHERE id = $1 AND tenant_id = $2 FOR UPDATE`,
		id, tenantID).Scan(&existingCandidateID, &existingRequirementID, &existingStatus, &existingOutcome); err != nil {
		respondWithError(w, http.StatusNotFound, "Interview not found")
		return
	}
	if interview.CandidateID != existingCandidateID || *interview.RequirementID != existingRequirementID {
		respondWithError(w, http.StatusBadRequest, "Interview Candidate and Requirement cannot be changed")
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

	wasActive := existingStatus == "scheduled" || existingStatus == "rescheduled"
	willBeActive := interview.Status == "scheduled" || interview.Status == "rescheduled"
	if !wasActive && willBeActive {
		var lockedID int
		if err := tx.QueryRow(`
			UPDATE candidates
			SET interview_locked = TRUE
			WHERE id = $1 AND tenant_id = $2 AND interview_locked = FALSE
			RETURNING id`, existingCandidateID, tenantID).Scan(&lockedID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondWithError(w, http.StatusConflict, "Candidate is already locked by another active interview")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Error acquiring interview lock")
			return
		}
	}
	if wasActive && (!willBeActive || strings.EqualFold(strings.TrimSpace(interview.Outcome), "rejected")) {
		if _, err := tx.Exec(`UPDATE candidates SET interview_locked = FALSE WHERE id = $1 AND tenant_id = $2`, existingCandidateID, tenantID); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error releasing interview lock")
			return
		}
	}

	interview.ID = id
	interview.TenantID = tenantID
	interview.CandidateName = candidateName
	interview.JobID = jobID
	interview.RequirementTitle = requirementTitle
	interview.Position = requirementTitle

	_, err = tx.Exec(`
		UPDATE interviews SET
			round = $1, interview_date = $2, status = $3, outcome = $4,
			feedback = $5, candidate_feedback = $6, next_action = $7,
			last_modified = NOW()
		WHERE id = $8 AND tenant_id = $9`,
		interview.Round, interview.InterviewDate, interview.Status,
		interview.Outcome, interview.Feedback, interview.CandidateFeedback,
		interview.NextAction, id, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating interview")
		return
	}

	if err := tx.QueryRow(`SELECT last_modified FROM interviews WHERE id = $1 AND tenant_id = $2`, id, tenantID).Scan(&interview.LastModified); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading updated interview")
		return
	}
	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error committing interview update")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true, Message: "Interview updated successfully", Data: interview,
	})
}
