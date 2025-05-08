
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetInterviews retrieves all interviews for a company
func GetInterviews(w http.ResponseWriter, r *http.Request) {
	// Get company name from context
	companyName := r.Context().Value("companyName").(string)
	
	interviews := []models.Interview{}
	rows, err := db.DB.Query("SELECT * FROM interviews WHERE company_name = $1", companyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching interviews")
		return
	}
	defer rows.Close()
	
	// Scan rows into interviews slice
	for rows.Next() {
		var i models.Interview
		err := rows.Scan(&i.ID, &i.CandidateID, &i.CandidateName, &i.Position,
			&i.InterviewDate, &i.Status, &i.Feedback, &i.LastModified, &i.CompanyName)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning interview row")
			return
		}
		interviews = append(interviews, i)
	}
	
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Interviews retrieved successfully",
		Data:    interviews,
	})
}

// GetInterviewByID retrieves a single interview by ID
func GetInterviewByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}
	
	// Get company name from context
	companyName := r.Context().Value("companyName").(string)
	
	var interview models.Interview
	err = db.DB.QueryRow(
		"SELECT * FROM interviews WHERE id = $1 AND company_name = $2", 
		id, companyName,
	).Scan(&interview.ID, &interview.CandidateID, &interview.CandidateName, &interview.Position,
		&interview.InterviewDate, &interview.Status, &interview.Feedback, 
		&interview.LastModified, &interview.CompanyName)
	
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Interview not found")
		return
	}
	
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Interview retrieved successfully",
		Data:    interview,
	})
}

// ScheduleInterview creates a new interview
func ScheduleInterview(w http.ResponseWriter, r *http.Request) {
	var interview models.Interview
	err := json.NewDecoder(r.Body).Decode(&interview)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	
	// Set company name from the authenticated user
	interview.CompanyName = r.Context().Value("companyName").(string)
	
	// Insert interview into database
	var id int
	err = db.DB.QueryRow(
		`INSERT INTO interviews (candidate_id, candidate_name, position, interview_date, status, feedback, company_name) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id`,
		interview.CandidateID, interview.CandidateName, interview.Position,
		interview.InterviewDate, interview.Status, interview.Feedback, interview.CompanyName,
	).Scan(&id)
	
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error scheduling interview")
		return
	}
	
	interview.ID = id
	
	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Interview scheduled successfully",
		Data:    interview,
	})
}

// UpdateInterview updates an existing interview
func UpdateInterview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}
	
	var interview models.Interview
	err = json.NewDecoder(r.Body).Decode(&interview)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	
	// Ensure company name matches authenticated user's company
	interview.CompanyName = r.Context().Value("companyName").(string)
	interview.ID = id
	
	// Update interview in database
	_, err = db.DB.Exec(
		`UPDATE interviews 
		SET candidate_id = $1, candidate_name = $2, position = $3, interview_date = $4, 
			status = $5, feedback = $6, last_modified = NOW() 
		WHERE id = $7 AND company_name = $8`,
		interview.CandidateID, interview.CandidateName, interview.Position,
		interview.InterviewDate, interview.Status, interview.Feedback, 
		interview.ID, interview.CompanyName,
	)
	
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating interview")
		return
	}
	
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Interview updated successfully",
		Data:    interview,
	})
}

// DeleteInterview deletes an interview
func DeleteInterview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}
	
	// Get company name from context
	companyName := r.Context().Value("companyName").(string)
	
	// Delete interview from database
	result, err := db.DB.Exec(
		"DELETE FROM interviews WHERE id = $1 AND company_name = $2", 
		id, companyName,
	)
	
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting interview")
		return
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Interview not found")
		return
	}
	
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Interview deleted successfully",
	})
}
