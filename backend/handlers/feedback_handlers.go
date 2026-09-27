package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/feedback"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

func feedbackService() *feedback.Service {
	return feedback.NewService(feedback.NewPostgresRepository(db.DB), db.DB)
}

type feedbackRequest struct {
	Outcome    string `json:"outcome"`
	ReasonCode string `json:"reasonCode,omitempty"`
	Comments   string `json:"comments,omitempty"`
	NextAction string `json:"nextAction,omitempty"`
}

type feedbackResponse struct {
	ID               int       `json:"id"`
	TenantID         string    `json:"tenantId"`
	SubmissionID     int       `json:"submissionId"`
	FeedbackByUserID int       `json:"feedbackByUserId"`
	Outcome          string    `json:"outcome"`
	ReasonCode       string    `json:"reasonCode,omitempty"`
	Comments         string    `json:"comments,omitempty"`
	NextAction       string    `json:"nextAction,omitempty"`
	FeedbackAt       time.Time `json:"feedbackAt"`
	CreatedAt        time.Time `json:"createdAt"`
}

func toFeedbackResponse(f *feedback.Feedback) feedbackResponse {
	return feedbackResponse{
		ID: f.ID, TenantID: f.TenantID, SubmissionID: f.SubmissionID,
		FeedbackByUserID: f.FeedbackByUserID, Outcome: string(f.Outcome),
		ReasonCode: f.ReasonCode, Comments: f.Comments, NextAction: f.NextAction,
		FeedbackAt: f.FeedbackAt, CreatedAt: f.CreatedAt,
	}
}

func AddSubmissionFeedback(w http.ResponseWriter, r *http.Request) {
	submissionID, err := strconv.Atoi(mux.Vars(r)["submissionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid submission ID")
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

	var req feedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	record, err := feedbackService().CreateFeedback(tenantID, feedback.CreateInput{
		SubmissionID: submissionID,
		FeedbackByUserID: actorUserID,
		Outcome: feedback.Outcome(req.Outcome),
		ReasonCode: req.ReasonCode,
		Comments: req.Comments,
		NextAction: req.NextAction,
	})
	if err != nil {
		switch {
		case errors.Is(err, feedback.ErrSubmissionNotFound):
			respondWithError(w, http.StatusNotFound, "Submission not found")
		case errors.Is(err, feedback.ErrFeedbackActorNotFound):
			respondWithError(w, http.StatusNotFound, "Feedback actor not found")
		case errors.Is(err, feedback.ErrInvalidOutcome), errors.Is(err, feedback.ErrInvalidReasonCode):
			respondWithError(w, http.StatusBadRequest, err.Error())
		default:
			respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Submission feedback recorded successfully",
		Data: toFeedbackResponse(record),
	})
}

func GetSubmissionFeedback(w http.ResponseWriter, r *http.Request) {
	submissionID, err := strconv.Atoi(mux.Vars(r)["submissionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid submission ID")
		return
	}
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context is required")
		return
	}

	records, err := feedbackService().ListBySubmission(tenantID, submissionID)
	if err != nil {
		if errors.Is(err, feedback.ErrSubmissionNotFound) {
			respondWithError(w, http.StatusNotFound, "Submission not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving feedback history")
		return
	}

	responses := make([]feedbackResponse, 0, len(records))
	for _, record := range records {
		responses = append(responses, toFeedbackResponse(record))
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Submission feedback history retrieved successfully",
		Data: responses,
	})
}
