package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetCandidates retrieves all candidates for the authenticated tenant.
// Scoped by tenant_id (ADR 0001), not company_name.
func GetCandidates(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)

	query := `SELECT id, name, email, phone, position, location, experience,
		currentctc, expectedctc, noticeperiod, jlptlanguage, skills, jobdescription,
		created_at, tenant_id, company_name FROM candidates WHERE tenant_id = $1`
	rows, err := db.DB.Query(query, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching candidates")
		return
	}
	defer rows.Close()

	candidates := []models.Candidate{}
	for rows.Next() {
		var c models.Candidate
		err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, &c.Location,
			&c.Experience, &c.CurrentCTC, &c.ExpectedCTC, &c.NoticePeriod, &c.JLPTLanguage,
			&c.Skills, &c.JobDescription, &c.CreatedAt, &c.TenantID, &c.CompanyName)
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

// GetCandidateByID retrieves a single candidate by ID, scoped to the
// authenticated tenant. A candidate ID belonging to another tenant returns
// 404, identically to a nonexistent ID — it must not disclose whether the
// resource exists in another tenant (ADR 0001).
func GetCandidateByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	var c models.Candidate
	err = db.DB.QueryRow(`
		SELECT id, name, email, phone, position, location, experience,
		currentctc, expectedctc, noticeperiod, jlptlanguage, skills, jobdescription,
		created_at, tenant_id, company_name
		FROM candidates WHERE id = $1 AND tenant_id = $2`, id, tenantID).Scan(
		&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, &c.Location, &c.Experience,
		&c.CurrentCTC, &c.ExpectedCTC, &c.NoticePeriod, &c.JLPTLanguage, &c.Skills,
		&c.JobDescription, &c.CreatedAt, &c.TenantID, &c.CompanyName)

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

// AddCandidate creates a new candidate under the authenticated tenant.
// tenant_id is always derived from context, never from the request payload.
func AddCandidate(w http.ResponseWriter, r *http.Request) {
	var c models.Candidate
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	c.TenantID = r.Context().Value("tenantID").(string)
	c.CompanyName = r.Context().Value("companyName").(string)

	err = db.DB.QueryRow(`
		INSERT INTO candidates (name, email, phone, position, location, experience,
			currentctc, expectedctc, noticeperiod, jlptlanguage, skills, jobdescription, tenant_id, company_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at`,
		c.Name, c.Email, c.Phone, c.Position, c.Location, c.Experience,
		c.CurrentCTC, c.ExpectedCTC, c.NoticePeriod, c.JLPTLanguage, c.Skills,
		c.JobDescription, c.TenantID, c.CompanyName).Scan(&c.ID, &c.CreatedAt)

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

// UpdateCandidate updates an existing candidate, scoped to the authenticated
// tenant. A candidate ID belonging to another tenant affects zero rows.
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

	tenantID := r.Context().Value("tenantID").(string)
	c.TenantID = tenantID
	c.ID = id

	result, err := db.DB.Exec(`
		UPDATE candidates SET name=$1, email=$2, phone=$3, position=$4, location=$5,
		experience=$6, currentctc=$7, expectedctc=$8, noticeperiod=$9, jlptlanguage=$10,
		skills=$11, jobdescription=$12
		WHERE id=$13 AND tenant_id=$14`,
		c.Name, c.Email, c.Phone, c.Position, c.Location, c.Experience,
		c.CurrentCTC, c.ExpectedCTC, c.NoticePeriod, c.JLPTLanguage, c.Skills,
		c.JobDescription, c.ID, tenantID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating candidate")
		return
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		respondWithError(w, http.StatusNotFound, "Candidate not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidate updated successfully",
		Data:    c,
	})
}

// DeleteCandidate deletes a candidate, scoped to the authenticated tenant.
func DeleteCandidate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	result, err := db.DB.Exec(`DELETE FROM candidates WHERE id = $1 AND tenant_id = $2`, id, tenantID)
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
