package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/audit"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type candidateScreeningRequest struct {
	RequirementID       int        `json:"requirementId"`
	RecruiterUserID     int        `json:"recruiterUserId"`
	CurrentCTC          string     `json:"currentCTC"`
	ExpectedCTC         string     `json:"expectedCTC"`
	NoticePeriod        string     `json:"noticePeriod"`
	LastWorkingDay      *time.Time `json:"lastWorkingDay"`
	CurrentLocation     string     `json:"currentLocation"`
	WillingToRelocate   *bool      `json:"willingToRelocate"`
	PreferredLocation   string     `json:"preferredLocation"`
	ReasonForChange     string     `json:"reasonForChange"`
	OffersInHand        string     `json:"offersInHand"`
	CandidateInterest   string     `json:"candidateInterest"`
	AvailabilityDate    *time.Time `json:"availabilityDate"`
	RelevantExperience  string     `json:"relevantExperience"`
	RecruiterAssessment string     `json:"recruiterAssessment"`
	Notes               string     `json:"notes"`
}

type candidateScreeningStatusRequest struct {
	Status string `json:"status"`
}

type candidateScreeningResponse struct {
	ID                  int        `json:"id"`
	TenantID            string     `json:"tenantId"`
	CandidateID         int        `json:"candidateId"`
	RequirementID       int        `json:"requirementId"`
	RecruiterUserID     int        `json:"recruiterUserId"`
	Status              string     `json:"status"`
	CurrentCTC          string     `json:"currentCTC,omitempty"`
	ExpectedCTC         string     `json:"expectedCTC,omitempty"`
	NoticePeriod        string     `json:"noticePeriod,omitempty"`
	LastWorkingDay      *time.Time `json:"lastWorkingDay,omitempty"`
	CurrentLocation     string     `json:"currentLocation,omitempty"`
	WillingToRelocate   *bool      `json:"willingToRelocate,omitempty"`
	PreferredLocation   string     `json:"preferredLocation,omitempty"`
	ReasonForChange     string     `json:"reasonForChange,omitempty"`
	OffersInHand        string     `json:"offersInHand,omitempty"`
	CandidateInterest   string     `json:"candidateInterest,omitempty"`
	AvailabilityDate    *time.Time `json:"availabilityDate,omitempty"`
	RelevantExperience  string     `json:"relevantExperience,omitempty"`
	RecruiterAssessment string     `json:"recruiterAssessment,omitempty"`
	Notes               string     `json:"notes,omitempty"`
	ScreenedAt          time.Time  `json:"screenedAt"`
	CreatedAt           time.Time  `json:"createdAt"`
}

func scanCandidateScreening(row interface{ Scan(...any) error }) (candidateScreeningResponse, error) {
	var s candidateScreeningResponse
	err := row.Scan(
		&s.ID, &s.TenantID, &s.CandidateID, &s.RequirementID, &s.RecruiterUserID,
		&s.Status, &s.CurrentCTC, &s.ExpectedCTC, &s.NoticePeriod,
		&s.LastWorkingDay, &s.CurrentLocation, &s.WillingToRelocate,
		&s.PreferredLocation, &s.ReasonForChange, &s.OffersInHand,
		&s.CandidateInterest, &s.AvailabilityDate, &s.RelevantExperience,
		&s.RecruiterAssessment, &s.Notes, &s.ScreenedAt, &s.CreatedAt,
	)
	return s, err
}

const candidateScreeningColumns = `id, tenant_id, candidate_id, requirement_id, recruiter_user_id,
	status, current_ctc, expected_ctc, notice_period, last_working_day,
	current_location, willing_to_relocate, preferred_location, reason_for_change,
	offers_in_hand, candidate_interest, availability_date, relevant_experience,
	recruiter_assessment, notes, screened_at, created_at`

func CreateCandidateScreening(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	var req candidateScreeningRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.RequirementID <= 0 || req.RecruiterUserID <= 0 {
		respondWithError(w, http.StatusBadRequest, "requirementId and recruiterUserId are required")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)
	tx, err := db.RequestDB(r).Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error starting screening transaction")
		return
	}
	defer tx.Rollback()

	var valid bool
	if err := tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM candidates c
			JOIN requirements r ON r.tenant_id = c.tenant_id
			JOIN users u ON u.tenant_id = c.tenant_id
			WHERE c.id = $1 AND r.id = $2 AND u.id = $3
			  AND c.tenant_id = $4 AND r.tenant_id = $4 AND u.tenant_id = $4
		)`, candidateID, req.RequirementID, req.RecruiterUserID, tenantID).Scan(&valid); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error validating screening references")
		return
	}
	if !valid {
		respondWithError(w, http.StatusNotFound, "Candidate, requirement, or recruiter not found")
		return
	}

	// The candidate row is the screening concurrency gate. This one UPDATE
	// performs the limit check and increment atomically.
	var updatedCandidateID int
	if err := tx.QueryRow(`
		UPDATE candidates
		SET screening_count = screening_count + 1
		WHERE id = $1 AND tenant_id = $2
		  AND screening_count < screening_limit
		RETURNING id`, candidateID, tenantID).Scan(&updatedCandidateID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusConflict, "Candidate has reached the screening limit")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error acquiring screening capacity")
		return
	}

	var screening candidateScreeningResponse
	err = tx.QueryRow(`
		INSERT INTO recruitment_screenings (
			tenant_id, candidate_id, requirement_id, recruiter_user_id, status,
			current_ctc, expected_ctc, notice_period, last_working_day,
			current_location, willing_to_relocate, preferred_location,
			reason_for_change, offers_in_hand, candidate_interest,
			availability_date, relevant_experience, recruiter_assessment, notes
		) VALUES ($1, $2, $3, $4, 'active', $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING `+candidateScreeningColumns,
		tenantID, candidateID, req.RequirementID, req.RecruiterUserID,
		req.CurrentCTC, req.ExpectedCTC, req.NoticePeriod, req.LastWorkingDay,
		req.CurrentLocation, req.WillingToRelocate, req.PreferredLocation,
		req.ReasonForChange, req.OffersInHand, req.CandidateInterest,
		req.AvailabilityDate, req.RelevantExperience, req.RecruiterAssessment, req.Notes,
	).Scan(
		&screening.ID, &screening.TenantID, &screening.CandidateID, &screening.RequirementID,
		&screening.RecruiterUserID, &screening.Status, &screening.CurrentCTC,
		&screening.ExpectedCTC, &screening.NoticePeriod, &screening.LastWorkingDay,
		&screening.CurrentLocation, &screening.WillingToRelocate, &screening.PreferredLocation,
		&screening.ReasonForChange, &screening.OffersInHand, &screening.CandidateInterest,
		&screening.AvailabilityDate, &screening.RelevantExperience, &screening.RecruiterAssessment,
		&screening.Notes, &screening.ScreenedAt, &screening.CreatedAt,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error recording screening")
		return
	}

	actorID, _ := r.Context().Value("userID").(int)
	if err := audit.WriteTx(tx, tenantID, actorID, "screening", screening.ID, "created", map[string]interface{}{"candidateId": candidateID, "requirementId": req.RequirementID}); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error recording screening audit event")
		return
	}
	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error committing screening")
		return
	}
	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true, Message: "Screening recorded successfully", Data: screening,
	})
}

func GetCandidateScreenings(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}
	tenantID := r.Context().Value("tenantID").(string)

	rows, err := db.RequestDB(r).Query(`
		SELECT `+candidateScreeningColumns+`
		FROM recruitment_screenings
		WHERE candidate_id = $1 AND tenant_id = $2
		ORDER BY screened_at DESC, id DESC`, candidateID, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching screening history")
		return
	}
	defer rows.Close()

	result := []candidateScreeningResponse{}
	for rows.Next() {
		s, err := scanCandidateScreening(rows)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error reading screening history")
			return
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading screening history")
		return
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true, Message: "Candidate screening history retrieved successfully", Data: result,
	})
}

func UpdateCandidateScreening(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}
	screeningID, err := strconv.Atoi(mux.Vars(r)["screeningId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid screening ID")
		return
	}

	var req candidateScreeningStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.Status != "completed" && req.Status != "rejected" && req.Status != "withdrawn" {
		respondWithError(w, http.StatusBadRequest, "status must be completed, rejected, or withdrawn")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)
	tx, err := db.RequestDB(r).Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error starting screening update")
		return
	}
	defer tx.Rollback()

	var previousStatus string
	if err := tx.QueryRow(`
		SELECT status FROM recruitment_screenings
		WHERE id = $1 AND candidate_id = $2 AND tenant_id = $3
		FOR UPDATE`, screeningID, candidateID, tenantID).Scan(&previousStatus); err != nil {
		respondWithError(w, http.StatusNotFound, "Screening not found")
		return
	}

	if previousStatus == "active" {
		var releasedCandidateID int
		if err := tx.QueryRow(`
			UPDATE candidates
			SET screening_count = screening_count - 1
			WHERE id = $1 AND tenant_id = $2 AND screening_count > 0
			RETURNING id`, candidateID, tenantID).Scan(&releasedCandidateID); err != nil {
			respondWithError(w, http.StatusConflict, "Candidate screening count is inconsistent")
			return
		}
	}

	if _, err := tx.Exec(`
		UPDATE recruitment_screenings
		SET status = $1
		WHERE id = $2 AND candidate_id = $3 AND tenant_id = $4`,
		req.Status, screeningID, candidateID, tenantID); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			respondWithError(w, http.StatusInternalServerError, "Error updating screening")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error updating screening")
		return
	}

	actorID, _ := r.Context().Value("userID").(int)
	if err := audit.WriteTx(tx, tenantID, actorID, "screening", screeningID, "updated", map[string]interface{}{"candidateId": candidateID, "status": req.Status}); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error recording screening audit event")
		return
	}
	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error committing screening update")
		return
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true, Message: "Screening updated successfully",
	})
}
