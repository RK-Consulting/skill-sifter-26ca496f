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

var interviewStatuses = map[string]bool{"scheduled": true, "completed": true, "cancelled": true, "rescheduled": true, "no_show": true}

func interviewTenant(r *http.Request) string     { return r.Context().Value("tenantID").(string) }
func validateInterviewStatus(status string) bool { return interviewStatuses[status] }

type errMissingJobID struct{}

func (errMissingJobID) Error() string { return "selected requirement does not have a Job ID" }

func validateInterviewReferences(r *http.Request, candidateID, requirementID int, tenantID string) (string, string, string, error) {
	var candidateName, jobID, requirementTitle string
	err := db.RequestDB(r.Context()).QueryRow(`SELECT c.name,r.job_id,r.title FROM candidates c JOIN requirements r ON r.tenant_id=c.tenant_id
		WHERE c.id=$1 AND r.id=$2 AND c.tenant_id=$3 AND r.tenant_id=$3`, candidateID, requirementID, tenantID).Scan(&candidateName, &jobID, &requirementTitle)
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
		if err := rows.Scan(&i.ID, &i.CandidateID, &i.CandidateName, &i.RequirementID, &i.JobID, &i.RequirementTitle, &i.Position, &i.Round, &i.InterviewDate, &i.Status, &i.Outcome, &i.Feedback, &i.CandidateFeedback, &i.NextAction, &i.LastModified, &i.TenantID, &i.CompanyName); err != nil {
			return nil, err
		}
		result = append(result, i)
	}
	return result, rows.Err()
}

func GetInterviews(w http.ResponseWriter, r *http.Request) {
	tenantID := interviewTenant(r)
	rows, err := db.RequestDB(r.Context()).Query(`SELECT i.id,i.candidate_id,i.candidate_name,i.requirement_id,COALESCE(r.job_id,''),COALESCE(r.title,''),i.position,i.round,i.interview_date,i.status,i.outcome,i.feedback,i.candidate_feedback,i.next_action,i.last_modified,i.tenant_id,i.company_name FROM interviews i LEFT JOIN requirements r ON r.id=i.requirement_id AND r.tenant_id=i.tenant_id WHERE i.tenant_id=$1 ORDER BY i.interview_date DESC,i.id DESC`, tenantID)
	if err != nil {
		respondWithError(w, 500, "Error fetching interviews")
		return
	}
	defer rows.Close()
	items, err := scanInterviewRows(rows)
	if err != nil {
		respondWithError(w, 500, "Error reading interview rows")
		return
	}
	respondWithJSON(w, 200, models.ApiResponse{Success: true, Message: "Interviews retrieved successfully", Data: items})
}

func GetInterviewByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, 400, "Invalid interview ID")
		return
	}
	tenantID := interviewTenant(r)
	var i models.Interview
	err = db.RequestDB(r.Context()).QueryRow(`SELECT i.id,i.candidate_id,i.candidate_name,i.requirement_id,COALESCE(r.job_id,''),COALESCE(r.title,''),i.position,i.round,i.interview_date,i.status,i.outcome,i.feedback,i.candidate_feedback,i.next_action,i.last_modified,i.tenant_id,i.company_name FROM interviews i LEFT JOIN requirements r ON r.id=i.requirement_id AND r.tenant_id=i.tenant_id WHERE i.id=$1 AND i.tenant_id=$2`, id, tenantID).Scan(&i.ID, &i.CandidateID, &i.CandidateName, &i.RequirementID, &i.JobID, &i.RequirementTitle, &i.Position, &i.Round, &i.InterviewDate, &i.Status, &i.Outcome, &i.Feedback, &i.CandidateFeedback, &i.NextAction, &i.LastModified, &i.TenantID, &i.CompanyName)
	if err != nil {
		respondWithError(w, 404, "Interview not found")
		return
	}
	respondWithJSON(w, 200, models.ApiResponse{Success: true, Message: "Interview retrieved successfully", Data: i})
}

func GetCandidateInterviews(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil {
		respondWithError(w, 400, "Invalid candidate ID")
		return
	}
	tenantID := interviewTenant(r)
	rows, err := db.RequestDB(r.Context()).Query(`SELECT i.id,i.candidate_id,i.candidate_name,i.requirement_id,COALESCE(r.job_id,''),COALESCE(r.title,''),i.position,i.round,i.interview_date,i.status,i.outcome,i.feedback,i.candidate_feedback,i.next_action,i.last_modified,i.tenant_id,i.company_name FROM interviews i LEFT JOIN requirements r ON r.id=i.requirement_id AND r.tenant_id=i.tenant_id WHERE i.candidate_id=$1 AND i.tenant_id=$2 ORDER BY i.interview_date DESC,i.id DESC`, candidateID, tenantID)
	if err != nil {
		respondWithError(w, 500, "Error fetching candidate interview history")
		return
	}
	defer rows.Close()
	items, err := scanInterviewRows(rows)
	if err != nil {
		respondWithError(w, 500, "Error reading interview history")
		return
	}
	respondWithJSON(w, 200, models.ApiResponse{Success: true, Message: "Candidate interview history retrieved successfully", Data: items})
}

func ScheduleInterview(w http.ResponseWriter, r *http.Request) {
	var i models.Interview
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		respondWithError(w, 400, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	tenantID := interviewTenant(r)
	companyName, _ := r.Context().Value("companyName").(string)
	if i.CandidateID <= 0 || i.RequirementID == nil || *i.RequirementID <= 0 {
		respondWithError(w, 400, "candidateId and requirementId are required")
		return
	}
	if i.Round <= 0 {
		i.Round = 1
	}
	if i.InterviewDate.IsZero() {
		respondWithError(w, 400, "interviewDate is required")
		return
	}
	if i.Status == "" {
		i.Status = "scheduled"
	}
	if !validateInterviewStatus(i.Status) {
		respondWithError(w, 400, "Invalid interview status")
		return
	}
	candidateName, jobID, title, err := validateInterviewReferences(r, i.CandidateID, *i.RequirementID, tenantID)
	if err != nil {
		if _, ok := err.(errMissingJobID); ok {
			respondWithError(w, 422, "Selected requirement does not have a Job ID")
			return
		}
		respondWithError(w, 404, "Candidate or requirement not found")
		return
	}
	var submitted bool
	if err := db.RequestDB(r.Context()).QueryRow(`SELECT EXISTS(
		SELECT 1 FROM recruitment_submissions
		WHERE candidate_id=$1 AND requirement_id=$2 AND tenant_id=$3
	)`, i.CandidateID, *i.RequirementID, tenantID).Scan(&submitted); err != nil {
		respondWithError(w, 500, "Error validating submission state")
		return
	}
	if !submitted {
		respondWithError(w, 422, "A submission is required before scheduling an interview")
		return
	}

	var feedbackRecorded bool
	if err := db.RequestDB(r.Context()).QueryRow(`SELECT EXISTS(
		SELECT 1 FROM recruitment_submission_feedback f
		JOIN recruitment_submissions s ON s.id=f.submission_id AND s.tenant_id=f.tenant_id
		WHERE s.candidate_id=$1 AND s.requirement_id=$2 AND s.tenant_id=$3
	)`, i.CandidateID, *i.RequirementID, tenantID).Scan(&feedbackRecorded); err != nil {
		respondWithError(w, 500, "Error validating feedback state")
		return
	}
	if !feedbackRecorded {
		respondWithError(w, 422, "Client feedback is required before scheduling an interview")
		return
	}

	var active bool
	if err := db.RequestDB(r.Context()).QueryRow(`SELECT EXISTS(SELECT 1 FROM interviews WHERE candidate_id=$1 AND requirement_id=$2 AND tenant_id=$3 AND status IN ('scheduled','rescheduled'))`, i.CandidateID, *i.RequirementID, tenantID).Scan(&active); err != nil {
		respondWithError(w, 500, "Error validating interview state")
		return
	}
	if active {
		respondWithError(w, 409, "An active interview already exists for this candidate and requirement")
		return
	}
	i.CandidateName = candidateName
	i.JobID = jobID
	i.RequirementTitle = title
	i.Position = title
	i.TenantID = tenantID
	i.CompanyName = companyName
	err = db.RequestDB(r.Context()).QueryRow(`INSERT INTO interviews(candidate_id,candidate_name,requirement_id,position,round,interview_date,status,outcome,feedback,candidate_feedback,next_action,tenant_id,company_name) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id,last_modified`, i.CandidateID, i.CandidateName, *i.RequirementID, i.Position, i.Round, i.InterviewDate, i.Status, i.Outcome, i.Feedback, i.CandidateFeedback, i.NextAction, tenantID, companyName).Scan(&i.ID, &i.LastModified)
	if err != nil {
		respondWithError(w, 500, "Error scheduling interview")
		return
	}
	respondWithJSON(w, 201, models.ApiResponse{Success: true, Message: "Interview scheduled successfully", Data: i})
}

func UpdateInterview(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, 400, "Invalid interview ID")
		return
	}
	var i models.Interview
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		respondWithError(w, 400, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	tenantID := interviewTenant(r)
	if i.Round <= 0 {
		i.Round = 1
	}
	if i.InterviewDate.IsZero() {
		respondWithError(w, 400, "interviewDate is required")
		return
	}
	if i.Status == "" {
		i.Status = "scheduled"
	}
	if !validateInterviewStatus(i.Status) {
		respondWithError(w, 400, "Invalid interview status")
		return
	}
	if i.CandidateID <= 0 || i.RequirementID == nil || *i.RequirementID <= 0 {
		respondWithError(w, 400, "candidateId and requirementId are required")
		return
	}
	var existingCandidateID, existingRequirementID int
	if err := db.RequestDB(r.Context()).QueryRow(`SELECT candidate_id,requirement_id FROM interviews WHERE id=$1 AND tenant_id=$2`, id, tenantID).Scan(&existingCandidateID, &existingRequirementID); err != nil {
		respondWithError(w, 404, "Interview not found")
		return
	}
	if i.CandidateID != existingCandidateID || *i.RequirementID != existingRequirementID {
		respondWithError(w, 400, "Interview Candidate and Requirement cannot be changed")
		return
	}
	if i.Status == "scheduled" || i.Status == "rescheduled" {
		var active bool
		if err := db.RequestDB(r.Context()).QueryRow(`SELECT EXISTS(SELECT 1 FROM interviews WHERE candidate_id=$1 AND requirement_id=$2 AND tenant_id=$3 AND id<>$4 AND status IN ('scheduled','rescheduled'))`, i.CandidateID, *i.RequirementID, tenantID, id).Scan(&active); err != nil {
			respondWithError(w, 500, "Error validating interview state")
			return
		}
		if active {
			respondWithError(w, 409, "An active interview already exists for this candidate and requirement")
			return
		}
	}
	name, jobID, title, err := validateInterviewReferences(r, i.CandidateID, *i.RequirementID, tenantID)
	if err != nil {
		respondWithError(w, 404, "Candidate or requirement not found")
		return
	}
	i.ID = id
	i.TenantID = tenantID
	i.CandidateName = name
	i.JobID = jobID
	i.RequirementTitle = title
	i.Position = title
	_, err = db.RequestDB(r.Context()).Exec(`UPDATE interviews SET round=$1,interview_date=$2,status=$3,outcome=$4,feedback=$5,candidate_feedback=$6,next_action=$7,last_modified=NOW() WHERE id=$8 AND tenant_id=$9`, i.Round, i.InterviewDate, i.Status, i.Outcome, i.Feedback, i.CandidateFeedback, i.NextAction, id, tenantID)
	if err != nil {
		respondWithError(w, 500, "Error updating interview")
		return
	}
	if err := db.RequestDB(r.Context()).QueryRow(`SELECT last_modified FROM interviews WHERE id=$1 AND tenant_id=$2`, id, tenantID).Scan(&i.LastModified); err != nil {
		respondWithError(w, 500, "Error reading updated interview")
		return
	}
	respondWithJSON(w, 200, models.ApiResponse{Success: true, Message: "Interview updated successfully", Data: i})
}
