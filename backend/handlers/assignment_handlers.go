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

func assignmentService() *assignment.Service {
	return assignment.NewService(assignment.NewPostgresRepository(db.DB), db.DB)
}

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

func respondWithAssignmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, assignment.ErrNotFound):
		respondWithError(w, http.StatusNotFound, "Assignment not found")
	case errors.Is(err, assignment.ErrCandidateNotFound):
		respondWithError(w, http.StatusNotFound, "Candidate not found")
	case errors.Is(err, assignment.ErrRequirementNotFound):
		respondWithError(w, http.StatusNotFound, "Requirement not found")
	case errors.Is(err, assignment.ErrUserNotFound):
		respondWithError(w, http.StatusNotFound, "User not found")
	case errors.Is(err, assignment.ErrCandidateAlreadyEngaged):
		respondWithError(w, http.StatusConflict, err.Error())
	case errors.Is(err, assignment.ErrCandidateNotEligible):
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, assignment.ErrDuplicateAssignment):
		respondWithError(w, http.StatusConflict, err.Error())
	default:
		var transitionErr *assignment.TransitionError
		if errors.As(err, &transitionErr) {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error processing assignment request")
	}
}

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

type addAssignmentRequest struct {
	CandidateID   int `json:"candidateId"`
	RequirementID int `json:"requirementId"`
	OwnerUserID   int `json:"ownerUserId,omitempty"`
}

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

type updateAssignmentRequest struct {
	OwnerUserID int `json:"ownerUserId"`
}

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

type transitionAssignmentRequest struct {
	Status string `json:"status"`
}

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
