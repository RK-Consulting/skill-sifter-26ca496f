package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

var selectionDecisions = map[string]bool{
	"selected": true,
	"rejected": true,
}

func validateSelectionDecision(decision string) bool {
	return selectionDecisions[decision]
}

type selectionRequest struct {
	Decision      string `json:"decision"`
	DecisionNotes string `json:"decisionNotes,omitempty"`
	NextAction    string `json:"nextAction,omitempty"`
}

type selectionResponse struct {
	ID            int       `json:"id"`
	TenantID      string    `json:"tenantId"`
	CandidateID   int       `json:"candidateId"`
	RequirementID int       `json:"requirementId"`
	Decision      string    `json:"decision"`
	DecisionNotes string    `json:"decisionNotes,omitempty"`
	NextAction    string    `json:"nextAction,omitempty"`
	DecidedAt     time.Time `json:"decidedAt"`
	LastModified  time.Time `json:"lastModified"`
}

func GetCandidateRequirementSelection(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}
	requirementID, err := strconv.Atoi(mux.Vars(r)["requirementId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid requirement ID")
		return
	}
	tenantID := r.Context().Value("tenantID").(string)

	var selection selectionResponse
	err = db.DB.QueryRow(
		`SELECT id, tenant_id, candidate_id, requirement_id, decision,
		        decision_notes, next_action, decided_at, last_modified
		FROM recruitment_selections
		WHERE candidate_id = $1 AND requirement_id = $2 AND tenant_id = $3`,
		candidateID, requirementID, tenantID,
	).Scan(
		&selection.ID, &selection.TenantID, &selection.CandidateID,
		&selection.RequirementID, &selection.Decision, &selection.DecisionNotes,
		&selection.NextAction, &selection.DecidedAt, &selection.LastModified,
	)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "Selection not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching selection")
		return
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true, Message: "Selection retrieved successfully", Data: selection,
	})
}

func CreateCandidateRequirementSelection(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}
	requirementID, err := strconv.Atoi(mux.Vars(r)["requirementId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid requirement ID")
		return
	}

	var req selectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if !validateSelectionDecision(req.Decision) {
		respondWithError(w, http.StatusBadRequest, "decision is required and must be one of: selected, rejected")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)
	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error starting selection transaction")
		return
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM candidates c
			JOIN requirements r ON r.tenant_id = c.tenant_id
			WHERE c.id = $1 AND r.id = $2
			  AND c.tenant_id = $3 AND r.tenant_id = $3
		)`, candidateID, requirementID, tenantID,
	).Scan(&exists); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error validating candidate and requirement")
		return
	}
	if !exists {
		respondWithError(w, http.StatusNotFound, "Candidate or requirement not found")
		return
	}

	// Selection is a decision after an interview, not a separate lifecycle
	// object. Require a completed interview for this Candidate × Requirement.
	var completedInterview bool
	if err := tx.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM interviews
			WHERE candidate_id = $1 AND requirement_id = $2 AND tenant_id = $3
			  AND status = 'completed'
		)`, candidateID, requirementID, tenantID,
	).Scan(&completedInterview); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error validating interview history")
		return
	}
	if !completedInterview {
		respondWithError(w, http.StatusUnprocessableEntity, "A completed interview is required before selection")
		return
	}

	var selection selectionResponse
	err = tx.QueryRow(
		`INSERT INTO recruitment_selections (
			tenant_id, candidate_id, requirement_id, decision,
			decision_notes, next_action
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tenant_id, candidate_id, requirement_id, decision,
		          decision_notes, next_action, decided_at, last_modified`,
		tenantID, candidateID, requirementID, req.Decision, req.DecisionNotes, req.NextAction,
	).Scan(
		&selection.ID, &selection.TenantID, &selection.CandidateID,
		&selection.RequirementID, &selection.Decision, &selection.DecisionNotes,
		&selection.NextAction, &selection.DecidedAt, &selection.LastModified,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			respondWithError(w, http.StatusConflict, "Selection decision already exists for this candidate and requirement")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error recording selection")
		return
	}

	// Rejection does not mutate Candidate master state. Selection is scoped to
	// Candidate × Requirement; the decision is the durable recruitment state.

	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error committing selection")
		return
	}
	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true, Message: "Selection recorded successfully", Data: selection,
	})
}
