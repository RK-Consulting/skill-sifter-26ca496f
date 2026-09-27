package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/screening"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

func screeningService() *screening.Service {
	return screening.NewService(screening.NewPostgresRepository(db.DB), db.DB)
}

type screeningResponse struct {
	ID int `json:"id"`
	TenantID string `json:"tenantId"`
	AssignmentID int `json:"assignmentId"`
	RecruiterUserID int `json:"recruiterUserId"`
	CurrentCTC string `json:"currentCTC,omitempty"`
	ExpectedCTC string `json:"expectedCTC,omitempty"`
	NoticePeriod string `json:"noticePeriod,omitempty"`
	LastWorkingDay *time.Time `json:"lastWorkingDay,omitempty"`
	CurrentLocation string `json:"currentLocation,omitempty"`
	WillingToRelocate *bool `json:"willingToRelocate,omitempty"`
	PreferredLocation string `json:"preferredLocation,omitempty"`
	ReasonForChange string `json:"reasonForChange,omitempty"`
	OffersInHand string `json:"offersInHand,omitempty"`
	CandidateInterest string `json:"candidateInterest,omitempty"`
	AvailabilityDate *time.Time `json:"availabilityDate,omitempty"`
	RelevantExperience string `json:"relevantExperience,omitempty"`
	RecruiterAssessment string `json:"recruiterAssessment,omitempty"`
	Notes string `json:"notes,omitempty"`
	ScreenedAt time.Time `json:"screenedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

func toScreeningResponse(s *screening.Screening) screeningResponse {
	return screeningResponse{
		ID:s.ID, TenantID:s.TenantID, AssignmentID:s.AssignmentID, RecruiterUserID:s.RecruiterUserID,
		CurrentCTC:s.CurrentCTC, ExpectedCTC:s.ExpectedCTC, NoticePeriod:s.NoticePeriod,
		LastWorkingDay:s.LastWorkingDay, CurrentLocation:s.CurrentLocation,
		WillingToRelocate:s.WillingToRelocate, PreferredLocation:s.PreferredLocation,
		ReasonForChange:s.ReasonForChange, OffersInHand:s.OffersInHand,
		CandidateInterest:s.CandidateInterest, AvailabilityDate:s.AvailabilityDate,
		RelevantExperience:s.RelevantExperience, RecruiterAssessment:s.RecruiterAssessment,
		Notes:s.Notes, ScreenedAt:s.ScreenedAt, CreatedAt:s.CreatedAt,
	}
}

type screeningRequest struct {
	CurrentCTC string `json:"currentCTC"`
	ExpectedCTC string `json:"expectedCTC"`
	NoticePeriod string `json:"noticePeriod"`
	LastWorkingDay string `json:"lastWorkingDay"`
	CurrentLocation string `json:"currentLocation"`
	WillingToRelocate *bool `json:"willingToRelocate"`
	PreferredLocation string `json:"preferredLocation"`
	ReasonForChange string `json:"reasonForChange"`
	OffersInHand string `json:"offersInHand"`
	CandidateInterest string `json:"candidateInterest"`
	AvailabilityDate string `json:"availabilityDate"`
	RelevantExperience string `json:"relevantExperience"`
	RecruiterAssessment string `json:"recruiterAssessment"`
	Notes string `json:"notes"`
}

func parseScreeningDate(value string) (*time.Time, error) {
	if value == "" { return nil, nil }
	t, err := time.Parse("2006-01-02", value)
	if err != nil { return nil, errors.New("dates must use YYYY-MM-DD") }
	return &t, nil
}

func decodeScreeningRequest(r *http.Request, assignmentID int) (screening.CreateInput, error) {
	var req screeningRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return screening.CreateInput{}, errors.New("invalid request payload")
	}
	defer r.Body.Close()

	lastWorkingDay, err := parseScreeningDate(req.LastWorkingDay)
	if err != nil { return screening.CreateInput{}, err }
	availabilityDate, err := parseScreeningDate(req.AvailabilityDate)
	if err != nil { return screening.CreateInput{}, err }

	recruiterUserID, ok := r.Context().Value("userID").(int)
	if !ok || recruiterUserID == 0 {
		return screening.CreateInput{}, errors.New("authenticated user is required")
	}

	return screening.CreateInput{
		AssignmentID:assignmentID, RecruiterUserID:recruiterUserID,
		CurrentCTC:req.CurrentCTC, ExpectedCTC:req.ExpectedCTC, NoticePeriod:req.NoticePeriod,
		LastWorkingDay:lastWorkingDay, CurrentLocation:req.CurrentLocation,
		WillingToRelocate:req.WillingToRelocate, PreferredLocation:req.PreferredLocation,
		ReasonForChange:req.ReasonForChange, OffersInHand:req.OffersInHand,
		CandidateInterest:req.CandidateInterest, AvailabilityDate:availabilityDate,
		RelevantExperience:req.RelevantExperience, RecruiterAssessment:req.RecruiterAssessment,
		Notes:req.Notes,
	}, nil
}

func AddAssignmentScreening(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil { respondWithError(w, http.StatusBadRequest, "Invalid assignment ID"); return }

	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" { respondWithError(w, http.StatusUnauthorized, "Tenant context is required"); return }

	input, err := decodeScreeningRequest(r, id)
	if err != nil { respondWithError(w, http.StatusBadRequest, err.Error()); return }

	s, err := screeningService().CreateScreening(tenantID, input)
	if err != nil {
		switch {
		case errors.Is(err, screening.ErrAssignmentNotFound):
			respondWithError(w, http.StatusNotFound, "Assignment not found")
		case errors.Is(err, screening.ErrRecruiterNotFound):
			respondWithError(w, http.StatusNotFound, "Recruiter not found")
		default:
			respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success:true, Message:"Recruiter screening recorded successfully", Data:toScreeningResponse(s),
	})
}

func GetAssignmentScreenings(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil { respondWithError(w, http.StatusBadRequest, "Invalid assignment ID"); return }
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" { respondWithError(w, http.StatusUnauthorized, "Tenant context is required"); return }

	screenings, err := screeningService().ListByAssignment(tenantID, id)
	if err != nil {
		if errors.Is(err, screening.ErrAssignmentNotFound) {
			respondWithError(w, http.StatusNotFound, "Assignment not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving screening history")
		return
	}

	responses := make([]screeningResponse, 0, len(screenings))
	for _, s := range screenings { responses = append(responses, toScreeningResponse(s)) }
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success:true, Message:"Screening history retrieved successfully", Data:responses,
	})
}

func GetAssignmentScreeningByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["screeningId"])
	if err != nil { respondWithError(w, http.StatusBadRequest, "Invalid screening ID"); return }
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" { respondWithError(w, http.StatusUnauthorized, "Tenant context is required"); return }

	s, err := screeningService().GetScreening(tenantID, id)
	if err != nil {
		if errors.Is(err, screening.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Screening not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving screening")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success:true, Message:"Screening retrieved successfully", Data:toScreeningResponse(s),
	})
}
