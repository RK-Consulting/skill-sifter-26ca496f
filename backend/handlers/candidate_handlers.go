
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetCandidates retrieves all candidates for a company
func GetCandidates(w http.ResponseWriter, r *http.Request) {
	// Get company ID from context
	companyID := r.Context().Value("companyID").(string)
	
	candidates := []models.Candidate{}
	rows, err := db.DB.Query("SELECT * FROM candidates WHERE company_id = $1", companyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching candidates")
		return
	}
	defer rows.Close()
	
	// Scan rows into candidates slice
	for rows.Next() {
		var c models.Candidate
		// Scan all fields from the row into the candidate struct
		err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, 
			&c.Status, &c.DateApplied, &c.ResumeURL, &c.CoverLetter, 
			&c.LastModified, &c.CompanyID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning candidate row")
			return
		}
		candidates = append(candidates, c)
	}
	
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidates retrieved successfully",
		Data:    candidates,
	})
}

// GetCandidateByID retrieves a single candidate by ID
func GetCandidateByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(string)
	
	var candidate models.Candidate
	err = db.DB.QueryRow(
		"SELECT * FROM candidates WHERE id = $1 AND company_id = $2", 
		id, companyID,
	).Scan(&candidate.ID, &candidate.Name, &candidate.Email, &candidate.Phone, 
		&candidate.Position, &candidate.Status, &candidate.DateApplied, 
		&candidate.ResumeURL, &candidate.CoverLetter, &candidate.LastModified, 
		&candidate.CompanyID)
	
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Candidate not found")
		return
	}
	
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidate retrieved successfully",
		Data:    candidate,
	})
}

// AddCandidate creates a new candidate
func AddCandidate(w http.ResponseWriter, r *http.Request) {
	var candidate models.Candidate
	err := json.NewDecoder(r.Body).Decode(&candidate)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	
	// Set company ID from the authenticated user
	candidate.CompanyID = r.Context().Value("companyID").(string)
	
	// Insert candidate into database
	var id int
	err = db.DB.QueryRow(
		`INSERT INTO candidates (name, email, phone, position, status, 
			resume_url, cover_letter, company_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id`,
		candidate.Name, candidate.Email, candidate.Phone, candidate.Position, 
		candidate.Status, candidate.ResumeURL, candidate.CoverLetter, 
		candidate.CompanyID,
	).Scan(&id)
	
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating candidate")
		return
	}
	
	candidate.ID = id
	
	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Candidate created successfully",
		Data:    candidate,
	})
}

// UpdateCandidate updates an existing candidate
func UpdateCandidate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}
	
	var candidate models.Candidate
	err = json.NewDecoder(r.Body).Decode(&candidate)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	
	// Ensure company ID matches authenticated user's company
	candidate.CompanyID = r.Context().Value("companyID").(string)
	candidate.ID = id
	
	// Update candidate in database
	_, err = db.DB.Exec(
		`UPDATE candidates 
		SET name = $1, email = $2, phone = $3, position = $4, status = $5, 
			resume_url = $6, cover_letter = $7, last_modified = NOW() 
		WHERE id = $8 AND company_id = $9`,
		candidate.Name, candidate.Email, candidate.Phone, candidate.Position, 
		candidate.Status, candidate.ResumeURL, candidate.CoverLetter, 
		candidate.ID, candidate.CompanyID,
	)
	
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating candidate")
		return
	}
	
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidate updated successfully",
		Data:    candidate,
	})
}

// DeleteCandidate deletes a candidate
func DeleteCandidate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}
	
	// Get company ID from context
	companyID := r.Context().Value("companyID").(string)
	
	// Delete candidate from database
	result, err := db.DB.Exec(
		"DELETE FROM candidates WHERE id = $1 AND company_id = $2", 
		id, companyID,
	)
	
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting candidate")
		return
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Candidate not found")
		return
	}
	
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidate deleted successfully",
	})
}
