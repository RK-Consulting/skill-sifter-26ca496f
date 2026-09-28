package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/submission"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

func submissionService(r *http.Request) *submission.Service {
	return submission.NewService(submission.NewPostgresRepository(db.RequestDB(r)), db.RequestDB(r))
}

type submissionRequest struct {
	RecipientType     string `json:"recipientType"`
	RecipientClientID *int   `json:"recipientClientId,omitempty"`
	RecipientUserID   *int   `json:"recipientUserId,omitempty"`
	RecipientName     string `json:"recipientName,omitempty"`
	RecipientEmail    string `json:"recipientEmail,omitempty"`
	SubmissionContext string `json:"submissionContext,omitempty"`
	RecruiterNotes    string `json:"recruiterNotes,omitempty"`
}

type submissionResponse struct {
	ID                  int             `json:"id"`
	TenantID            string          `json:"tenantId"`
	CandidateID         int             `json:"candidateId"`
	RequirementID       int             `json:"requirementId"`
	SubmittedByUserID   int             `json:"submittedByUserId"`
	RecipientType       string          `json:"recipientType"`
	RecipientClientID   *int            `json:"recipientClientId,omitempty"`
	RecipientUserID     *int            `json:"recipientUserId,omitempty"`
	RecipientName       string          `json:"recipientName,omitempty"`
	RecipientEmail      string          `json:"recipientEmail,omitempty"`
	SubmissionContext   string          `json:"submissionContext,omitempty"`
	RecruiterNotes      string          `json:"recruiterNotes,omitempty"`
	CandidateSnapshot   json.RawMessage `json:"candidateSnapshot"`
	RequirementSnapshot json.RawMessage `json:"requirementSnapshot"`
	SubmittedAt         string          `json:"submittedAt"`
}

func toSubmissionResponse(s *submission.Submission) submissionResponse {
	return submissionResponse{ID: s.ID, TenantID: s.TenantID, CandidateID: s.CandidateID, RequirementID: s.RequirementID, SubmittedByUserID: s.SubmittedByUserID,
		RecipientType: string(s.RecipientType), RecipientClientID: s.RecipientClientID, RecipientUserID: s.RecipientUserID, RecipientName: s.RecipientName,
		RecipientEmail: s.RecipientEmail, SubmissionContext: s.SubmissionContext, RecruiterNotes: s.RecruiterNotes, CandidateSnapshot: s.CandidateSnapshot,
		RequirementSnapshot: s.RequirementSnapshot, SubmittedAt: s.SubmittedAt.Format("2006-01-02T15:04:05Z07:00")}
}

func AddCandidateRequirementSubmission(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil {
		respondWithError(w, 400, "Invalid candidate ID")
		return
	}
	requirementID, err := strconv.Atoi(mux.Vars(r)["requirementId"])
	if err != nil {
		respondWithError(w, 400, "Invalid requirement ID")
		return
	}
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, 401, "Tenant context is required")
		return
	}
	actorUserID, ok := r.Context().Value("userID").(int)
	if !ok || actorUserID == 0 {
		respondWithError(w, 401, "Authenticated user is required")
		return
	}
	var req submissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, 400, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	record, err := submissionService(r).Submit(tenantID, submission.CreateInput{CandidateID: candidateID, RequirementID: requirementID, SubmittedByUserID: actorUserID,
		RecipientType: submission.RecipientType(req.RecipientType), RecipientClientID: req.RecipientClientID, RecipientUserID: req.RecipientUserID, RecipientName: req.RecipientName,
		RecipientEmail: req.RecipientEmail, SubmissionContext: req.SubmissionContext, RecruiterNotes: req.RecruiterNotes})
	if err != nil {
		switch {
		case errors.Is(err, submission.ErrCandidateNotFound):
			respondWithError(w, 404, "Candidate or requirement not found")
		case errors.Is(err, submission.ErrClientNotFound), errors.Is(err, submission.ErrRecipientNotFound):
			respondWithError(w, 404, "Submission recipient not found")
		case errors.Is(err, submission.ErrAlreadySubmitted):
			respondWithError(w, 409, "Candidate has already been submitted for this requirement")
		case errors.Is(err, submission.ErrInvalidRecipient):
			respondWithError(w, 400, "A valid submission recipient is required")
		default:
			respondWithError(w, 422, err.Error())
		}
		return
	}
	respondWithJSON(w, 201, models.ApiResponse{Success: true, Message: "Candidate submitted successfully", Data: toSubmissionResponse(record)})
}

func GetCandidateRequirementSubmissions(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil {
		respondWithError(w, 400, "Invalid candidate ID")
		return
	}
	requirementID, err := strconv.Atoi(mux.Vars(r)["requirementId"])
	if err != nil {
		respondWithError(w, 400, "Invalid requirement ID")
		return
	}
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, 401, "Tenant context is required")
		return
	}
	records, err := submissionService(r).ListByCandidateRequirement(tenantID, candidateID, requirementID)
	if err != nil {
		if errors.Is(err, submission.ErrCandidateNotFound) {
			respondWithError(w, 404, "Candidate or requirement not found")
			return
		}
		respondWithError(w, 500, "Error retrieving submission history")
		return
	}
	responses := make([]submissionResponse, 0, len(records))
	for _, record := range records {
		responses = append(responses, toSubmissionResponse(record))
	}
	respondWithJSON(w, 200, models.ApiResponse{Success: true, Message: "Submission history retrieved successfully", Data: responses})
}
