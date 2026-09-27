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

func submissionService() *submission.Service {
	return submission.NewService(submission.NewPostgresRepository(db.DB), db.DB)
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
	AssignmentID        int             `json:"assignmentId"`
	SubmittedByUserID   int             `json:"submittedByUserId"`
	RecipientType       string          `json:"recipientType"`
	RecipientClientID   *int            `json:"recipientClientId,omitempty"`
	RecipientUserID     *int            `json:"recipientUserId,omitempty"`
	RecipientName       string          `json:"recipientName,omitempty"`
	RecipientEmail      string          `json:"recipientEmail,omitempty"`
	SubmissionContext   string          `json:"submissionContext,omitempty"`
	RecruiterNotes      string          `json:"recruiterNotes,omitempty"`
	CandidateSnapshot   json.RawMessage  `json:"candidateSnapshot"`
	RequirementSnapshot json.RawMessage  `json:"requirementSnapshot"`
	SubmittedAt         string          `json:"submittedAt"`
}

func toSubmissionResponse(s *submission.Submission) submissionResponse {
	return submissionResponse{
		ID: s.ID, TenantID: s.TenantID, AssignmentID: s.AssignmentID,
		SubmittedByUserID: s.SubmittedByUserID, RecipientType: string(s.RecipientType),
		RecipientClientID: s.RecipientClientID, RecipientUserID: s.RecipientUserID,
		RecipientName: s.RecipientName, RecipientEmail: s.RecipientEmail,
		SubmissionContext: s.SubmissionContext, RecruiterNotes: s.RecruiterNotes,
		CandidateSnapshot: s.CandidateSnapshot, RequirementSnapshot: s.RequirementSnapshot,
		SubmittedAt: s.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func AddAssignmentSubmission(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid assignment ID")
		return
	}

	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context is required")
		return
	}
	actorUserID, ok := r.Context().Value("userID").(int)
	if !ok || actorUserID == 0 {
		respondWithError(w, http.StatusUnauthorized, "Authenticated user is required")
		return
	}

	var req submissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	record, err := submissionService().Submit(tenantID, submission.CreateInput{
		AssignmentID: id, SubmittedByUserID: actorUserID,
		RecipientType: submission.RecipientType(req.RecipientType),
		RecipientClientID: req.RecipientClientID, RecipientUserID: req.RecipientUserID,
		RecipientName: req.RecipientName, RecipientEmail: req.RecipientEmail,
		SubmissionContext: req.SubmissionContext, RecruiterNotes: req.RecruiterNotes,
	})
	if err != nil {
		switch {
		case errors.Is(err, submission.ErrAssignmentNotFound):
			respondWithError(w, http.StatusNotFound, "Assignment not found")
		case errors.Is(err, submission.ErrClientNotFound), errors.Is(err, submission.ErrRecipientNotFound):
			respondWithError(w, http.StatusNotFound, "Submission recipient not found")
		case errors.Is(err, submission.ErrAlreadySubmitted):
			respondWithError(w, http.StatusConflict, "Assignment has already been submitted")
		case errors.Is(err, submission.ErrInvalidRecipient):
			respondWithError(w, http.StatusBadRequest, "A valid submission recipient is required")
		default:
			respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true, Message: "Candidate submitted successfully", Data: toSubmissionResponse(record),
	})
}

func GetAssignmentSubmissions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid assignment ID")
		return
	}
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context is required")
		return
	}

	records, err := submissionService().ListByAssignment(tenantID, id)
	if err != nil {
		if errors.Is(err, submission.ErrAssignmentNotFound) {
			respondWithError(w, http.StatusNotFound, "Assignment not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving submission history"); return
	}

	responses := make([]submissionResponse, 0, len(records))
	for _, record := range records {
		responses = append(responses, toSubmissionResponse(record))
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true, Message: "Submission history retrieved successfully", Data: responses,
	})
}
