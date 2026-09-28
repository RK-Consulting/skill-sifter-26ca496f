package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/billing"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type billingRequest struct {
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	InvoiceReference string `json:"invoiceReference,omitempty"`
}

func billingService() *billing.Service {
	return billing.NewService(billing.NewPostgresRepository(db.DB), db.DB)
}

func GetBillingWorklist(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	items, err := billingService().ListWorklist(tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching billing worklist")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Billing worklist retrieved successfully",
		Data:    items,
	})
}

func parseBillingPair(r *http.Request) (int, int, error) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil {
		return 0, 0, errors.New("invalid candidate ID")
	}
	requirementID, err := strconv.Atoi(mux.Vars(r)["requirementId"])
	if err != nil {
		return 0, 0, errors.New("invalid requirement ID")
	}
	return candidateID, requirementID, nil
}

func GetCandidateRequirementBilling(w http.ResponseWriter, r *http.Request) {
	candidateID, requirementID, err := parseBillingPair(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	tenantID := r.Context().Value("tenantID").(string)
	b, err := billingService().Get(tenantID, candidateID, requirementID)
	if errors.Is(err, billing.ErrNotFound) {
		respondWithError(w, http.StatusNotFound, "Billing record not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching billing record")
		return
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Billing record retrieved successfully",
		Data:    b,
	})
}

func CreateCandidateRequirementBilling(w http.ResponseWriter, r *http.Request) {
	candidateID, requirementID, err := parseBillingPair(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	var req billingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	tenantID := r.Context().Value("tenantID").(string)
	b, err := billingService().Create(tenantID, billing.CreateInput{
		CandidateID:      candidateID,
		RequirementID:    requirementID,
		Amount:           req.Amount,
		Currency:         req.Currency,
		InvoiceReference: req.InvoiceReference,
	})
	switch {
	case errors.Is(err, billing.ErrCandidateRequirementNotFound):
		respondWithError(w, http.StatusNotFound, "Candidate or requirement not found")
	case errors.Is(err, billing.ErrJoiningNotFound):
		respondWithError(w, http.StatusUnprocessableEntity, "A joined recruitment record is required before billing")
	case errors.Is(err, billing.ErrBillingExists):
		respondWithError(w, http.StatusConflict, "Billing record already exists for this candidate and requirement")
	case errors.Is(err, billing.ErrInvalidAmount):
		respondWithError(w, http.StatusBadRequest, "Billing amount is required")
	case errors.Is(err, billing.ErrInvalidCurrency):
		respondWithError(w, http.StatusBadRequest, "Billing currency must be a three-letter code")
	case err != nil:
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			respondWithError(w, http.StatusConflict, "Billing record already exists for this candidate and requirement")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error creating billing record")
		}
	default:
		respondWithJSON(w, http.StatusCreated, models.ApiResponse{
			Success: true,
			Message: "Billing record created successfully",
			Data:    b,
		})
	}
}
