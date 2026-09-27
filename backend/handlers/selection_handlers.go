package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/assignment"
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
	AssignmentID  int       `json:"assignmentId"`
	Decision      string    `json:"decision"`
	DecisionNotes string    `json:"decisionNotes,omitempty"`
	NextAction    string    `json:"nextAction,omitempty"`
	DecidedAt     time.Time `json:"decidedAt"`
	LastModified  time.Time `json:"lastModified"`
}

func GetAssignmentSelection(w http.ResponseWriter, r *http.Request) {
	assignmentID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid assignment ID")
		return
	}
	tenantID := r.Context().Value("tenantID").(string)

	var selection selectionResponse
	err = db.DB.QueryRow(
		`SELECT id, tenant_id, assignment_id, decision, decision_notes, next_action,
		       decided_at, last_modified
		FROM recruitment_selections
		WHERE assignment_id = $1 AND tenant_id = $2`,
		assignmentID, tenantID,
	).Scan(
		&selection.ID, &selection.TenantID, &selection.AssignmentID,
		&selection.Decision, &selection.DecisionNotes, &selection.NextAction,
		&selection.DecidedAt, &selection.LastModified,
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
		Success: true,
		Message: "Selection retrieved successfully",
		Data:    selection,
	})
}

func CreateAssignmentSelection(w http.ResponseWriter, r *http.Request) {
	assignmentID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid assignment ID")
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
	actorUserID := r.Context().Value("userID").(int)

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error starting selection transaction")
		return
	}
	defer tx.Rollback()

	var existingDecision string
	err = tx.QueryRow(
		`SELECT decision
		FROM recruitment_selections
		WHERE assignment_id = $1 AND tenant_id = $2
		FOR UPDATE`,
		assignmentID, tenantID,
	).Scan(&existingDecision)
	if err == nil {
		respondWithError(w, http.StatusConflict, "Selection decision already exists for this assignment")
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusInternalServerError, "Error checking selection history")
		return
	}

	targetStatus := assignment.StatusOffered
	if req.Decision == "rejected" {
		targetStatus = assignment.StatusRejected
	}

	if _, err := assignmentService().TransitionAssignmentTx(tx, tenantID, actorUserID, assignmentID, targetStatus); err != nil {
		respondWithAssignmentError(w, err)
		return
	}

	var selection selectionResponse
	err = tx.QueryRow(
		`INSERT INTO recruitment_selections (
			tenant_id, assignment_id, decision, decision_notes, next_action
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, tenant_id, assignment_id, decision, decision_notes, next_action,
		          decided_at, last_modified`,
		tenantID, assignmentID, req.Decision, req.DecisionNotes, req.NextAction,
	).Scan(
		&selection.ID, &selection.TenantID, &selection.AssignmentID,
		&selection.Decision, &selection.DecisionNotes, &selection.NextAction,
		&selection.DecidedAt, &selection.LastModified,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			respondWithError(w, http.StatusConflict, "Selection decision already exists for this assignment")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error recording selection")
		return
	}

	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error committing selection")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Selection recorded successfully",
		Data:    selection,
	})
}
