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
	companyName := r.Context().Value("companyName").(string)

	query := `SELECT id, name, email, phone, position, status, date_applied, resume_url, cover_letter, last_modified, company_name, source FROM candidates WHERE company_name = $1`
	rows, err := db.DB.Query(query, companyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching candidates")
		return
	}
	defer rows.Close()

	candidates := []models.Candidate{}
	for rows.Next() {
		var c models.Candidate
		err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, &c.Status,
			&c.DateApplied, &c.ResumeURL, &c.CoverLetter, &c.LastModified, &c.CompanyName, &c.Source)
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

	companyName := r.Context().Value("companyName").(string)

	var c models.Candidate
	err = db.DB.QueryRow(`
		SELECT id, name, email, phone, position, status, date_applied, resume_url, cover_letter, last_modified, company_name, source 
		FROM candidates WHERE id = $1 AND company_name = $2`, id, companyName).Scan(
		&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, &c.Status, &c.DateApplied,
		&c.ResumeURL, &c.CoverLetter, &c.LastModified, &c.CompanyName, &c.Source)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Candidate not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidate retrieved successfully",
		Data:    c,
	})
}

// AddCandidate creates a new candidate
func AddCandidate(w http.ResponseWriter, r *http.Request) {
	var c models.Candidate
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	c.CompanyName = r.Context().Value("companyName").(string)

	err = db.DB.QueryRow(`
		INSERT INTO candidates (name, email, phone, position, status, resume_url, cover_letter, company_name, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		c.Name, c.Email, c.Phone, c.Position, c.Status, c.ResumeURL, c.CoverLetter, c.CompanyName, c.Source).Scan(&c.ID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating candidate")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Candidate created successfully",
		Data:    c,
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

	var c models.Candidate
	err = json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	c.CompanyName = r.Context().Value("companyName").(string)
	c.ID = id

	_, err = db.DB.Exec(`
		UPDATE candidates SET name=$1, email=$2, phone=$3, position=$4, status=$5,
		resume_url=$6, cover_letter=$7, last_modified=NOW(), source=$8
		WHERE id=$9 AND company_name=$10`,
		c.Name, c.Email, c.Phone, c.Position, c.Status,
		c.ResumeURL, c.CoverLetter, c.Source, c.ID, c.CompanyName)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating candidate")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidate updated successfully",
		Data:    c,
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

	companyName := r.Context().Value("companyName").(string)

	result, err := db.DB.Exec(`DELETE FROM candidates WHERE id = $1 AND company_name = $2`, id, companyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting candidate")
		return
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		respondWithError(w, http.StatusNotFound, "Candidate not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidate deleted successfully",
	})
}
