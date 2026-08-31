package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/assignment"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// assignmentService constructs a fresh assignment.Service backed by the
// current db.DB on every call, rather than caching one at package-init
// time. db.DB is reassigned in tests (and could in principle be
// reconnected in production), so caching a Service pointing at a stale
// *sql.DB would silently operate against the wrong connection.
func assignmentService() *assignment.Service {
	return assignment.NewService(assignment.NewPostgresRepository(db.DB), db.DB)
}

// assignmentResponse is the HTTP wire format for an Assignment,
// deliberately separate from the domain type (assignment.Assignment has no
// JSON tags on purpose — see domain/assignment/assignment.go). Submission
// snapshot fields are intentionally omitted from this checkpoint's
// response: snapshot capture is not implemented yet (Issue #35 checkpoint
// 3 scope), so every assignment's snapshots are always nil regardless of
// status, and exposing three always-null fields would be noise.
type assignmentResponse struct {
	ID              int       `json:"id"`
	TenantID        string    `json:"tenantId"`
	CandidateID     int       `json:"candidateId"`
	RequirementID   int       `json:"requirementId"`
	Status          string    `json:"status"`
	CreatedByUserID int       `json:"createdByUserId"`
	OwnerUserID     int       `json:"ownerUserId"`
	CreatedAt       time.Time `json:"createdAt"`
	LastModified    time.Time `json:"lastModified"`
}

func toAssignmentResponse(a *assignment.Assignment) assignmentResponse {
	return assignmentResponse{
		ID:              a.ID,
		TenantID:        a.TenantID,
		CandidateID:     a.CandidateID,
		RequirementID:   a.RequirementID,
		Status:          string(a.Status),
		CreatedByUserID: a.CreatedByUserID,
		OwnerUserID:     a.OwnerUserID,
		CreatedAt:       a.CreatedAt,
		LastModified:    a.LastModified,
	}
}

// respondWithAssignmentError maps the domain/service error vocabulary onto
// HTTP status codes. This is the one place that mapping lives, so it stays
// consistent across all five handlers below rather than being reimplemented
// per handler.
func respondWithAssignmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, assignment.ErrNotFound):
		respondWithError(w, http.StatusNotFound, "Assignment not found")
	case errors.Is(err, assignment.ErrCandidateNotFound):
		// 404, not 403: must not disclose whether that candidate ID exists
		// in another tenant (same convention as #33/#34).
		respondWithError(w, http.StatusNotFound, "Candidate not found")
	case errors.Is(err, assignment.ErrRequirementNotFound):
		respondWithError(w, http.StatusNotFound, "Requirement not found")
	case errors.Is(err, assignment.ErrUserNotFound):
		respondWithError(w, http.StatusNotFound, "User not found")
	case errors.Is(err, assignment.ErrCandidateNotEligible):
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, assignment.ErrDuplicateAssignment):
		respondWithError(w, http.StatusConflict, err.Error())
	default:
		var transitionErr *assignment.TransitionError
		if errors.As(err, &transitionErr) {
			// A genuine state conflict, not a malformed request: the
			// target status is a recognized value (validated earlier in
			// TransitionAssignment, before this error can occur), but not
			// a legal transition from the assignment's current status, or
			// the assignment is already in a terminal state.
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error processing assignment request")
	}
}

// GetAssignments retrieves all recruitment assignments for the
// authenticated tenant. Scoped by tenant_id (ADR 0001), not company_name.
func GetAssignments(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)

	assignments, err := assignmentService().ListAssignments(tenantID)
	if err != nil {
		respondWithAssignmentError(w, err)
		return
	}

	responses := make([]assignmentResponse, 0, len(assignments))
	for _, a := range assignments {
		responses = append(responses, toAssignmentResponse(a))
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Assignments retrieved successfully",
		Data:    responses,
	})
}

// GetAssignmentByID retrieves a single assignment by ID, scoped to the
// authenticated tenant. An assignment ID belonging to another tenant
// returns 404, identically to a nonexistent ID.
func GetAssignmentByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid assignment ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	a, err := assignmentService().GetAssignment(tenantID, id)
	if err != nil {
		respondWithAssignmentError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Assignment retrieved successfully",
		Data:    toAssignmentResponse(a),
	})
}

// addAssignmentRequest is the create request body. TenantID and
// CreatedByUserID are deliberately absent — those always come from
// authenticated context (see AddAssignment), never from the request.
type addAssignmentRequest struct {
	CandidateID   int `json:"candidateId"`
	RequirementID int `json:"requirementId"`
	// OwnerUserID may be omitted, in which case the assignment's owner
	// defaults to the authenticated actor (enforced by
	// assignment.Service.CreateAssignment, not here).
	OwnerUserID int `json:"ownerUserId,omitempty"`
}

// AddAssignment creates a new recruitment assignment under the
// authenticated tenant. All business rules (tenant consistency, candidate
// eligibility, duplicate-pair rejection) live in assignment.Service — this
// handler only translates the HTTP request into a service call and the
// result back into an HTTP response.
func AddAssignment(w http.ResponseWriter, r *http.Request) {
	var req addAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.CandidateID == 0 {
		respondWithError(w, http.StatusBadRequest, "candidateId is required")
		return
	}
	if req.RequirementID == 0 {
		respondWithError(w, http.StatusBadRequest, "requirementId is required")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)
	actorUserID := r.Context().Value("userID").(int)

	a, err := assignmentService().CreateAssignment(tenantID, actorUserID, assignment.CreateInput{
		CandidateID:   req.CandidateID,
		RequirementID: req.RequirementID,
		OwnerUserID:   req.OwnerUserID,
	})
	if err != nil {
		respondWithAssignmentError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Assignment created successfully",
		Data:    toAssignmentResponse(a),
	})
}

// updateAssignmentRequest is the update request body. This checkpoint
// intentionally supports ONLY reassigning the owner — status (lifecycle)
// transitions are explicitly out of scope for this checkpoint (Issue #35),
// and candidate_id/requirement_id are the assignment's identity (enforced
// by the database's UNIQUE(candidate_id, requirement_id) constraint) and
// are not mutable via this endpoint. A candidateId/requirementId/status in
// the request body is accepted for forward JSON compatibility but silently
// ignored, matching how other handlers in this codebase ignore unknown/
// not-yet-supported fields rather than rejecting the whole request.
type updateAssignmentRequest struct {
	OwnerUserID int `json:"ownerUserId"`
}

// UpdateAssignment reassigns an existing assignment's owner, scoped to the
// authenticated tenant. An assignment ID belonging to another tenant
// affects nothing (assignment.Service.ChangeOwner returns ErrNotFound).
func UpdateAssignment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid assignment ID")
		return
	}

	var req updateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.OwnerUserID == 0 {
		respondWithError(w, http.StatusBadRequest, "ownerUserId is required")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)
	actorUserID := r.Context().Value("userID").(int)

	a, err := assignmentService().ChangeOwner(tenantID, actorUserID, id, req.OwnerUserID)
	if err != nil {
		respondWithAssignmentError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Assignment updated successfully",
		Data:    toAssignmentResponse(a),
	})
}

// DeleteAssignment deletes an assignment, scoped to the authenticated
// tenant.
func DeleteAssignment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid assignment ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	if err := assignmentService().DeleteAssignment(tenantID, id); err != nil {
		respondWithAssignmentError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Assignment deleted successfully",
	})
}

// transitionAssignmentRequest is the request body for the dedicated
// lifecycle-transition endpoint. This is deliberately a separate endpoint
// from UpdateAssignment/PUT (which only reassigns owner) rather than
// letting arbitrary status values enter the PUT payload — keeping owner
// mutation and lifecycle transition as two distinct concepts.
type transitionAssignmentRequest struct {
	Status string `json:"status"`
}

// TransitionAssignment moves an assignment to a new lifecycle status,
// scoped to the authenticated tenant. All transition-legality rules (ADR
// 0003 section 4) live in assignment.Assignment.TransitionTo via
// assignment.Service.TransitionAssignment — this handler validates only
// that the request is well-formed (a non-empty, recognized status value)
// before delegating, so "not a recognized status at all" (400, malformed
// request) stays distinct from "a recognized status but not a legal
// transition from here" (409, domain conflict), which is everything
// TransitionAssignment itself can return.
func TransitionAssignment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid assignment ID")
		return
	}

	var req transitionAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	targetStatus := assignment.Status(req.Status)
	if req.Status == "" || !targetStatus.Valid() {
		respondWithError(w, http.StatusBadRequest, "status is required and must be one of: draft, screening, submitted, interviewing, offered, joined, rejected, withdrawn")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)
	actorUserID := r.Context().Value("userID").(int)

	a, err := assignmentService().TransitionAssignment(tenantID, actorUserID, id, targetStatus)
	if err != nil {
		respondWithAssignmentError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Assignment transitioned successfully",
		Data:    toAssignmentResponse(a),
	})
}
