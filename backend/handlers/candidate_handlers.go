package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetCandidates retrieves all candidates for the authenticated tenant.
func GetCandidates(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, name, email, phone, position, location, experience,
		       currentctc, expectedctc, noticeperiod, jobdescription,
		       status, created_at, tenant_id, company_name
		FROM candidates
		WHERE tenant_id = $1
		ORDER BY id`, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching candidates")
		return
	}
	defer rows.Close()

	candidates := make([]models.Candidate, 0)

	for rows.Next() {
		var c models.Candidate

		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Email,
			&c.Phone,
			&c.Position,
			&c.Location,
			&c.Experience,
			&c.CurrentCTC,
			&c.ExpectedCTC,
			&c.NoticePeriod,
			&c.JobDescription,
			&c.Status,
			&c.CreatedAt,
			&c.TenantID,
			&c.CompanyName,
		); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning candidate row")
			return
		}

		if err := loadCandidateExpertise(&c, tenantID); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error loading candidate expertise")
			return
		}

		candidates = append(candidates, c)
	}

	if err := rows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading candidates")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidates retrieved successfully",
		Data:    candidates,
	})
}

// GetCandidateByID retrieves a single candidate scoped to the authenticated
// tenant.
func GetCandidateByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	var c models.Candidate

	err = db.DB.QueryRow(`
		SELECT id, name, email, phone, position, location, experience,
		       currentctc, expectedctc, noticeperiod, jobdescription,
		       status, created_at, tenant_id, company_name
		FROM candidates
		WHERE id = $1 AND tenant_id = $2`,
		id,
		tenantID,
	).Scan(
		&c.ID,
		&c.Name,
		&c.Email,
		&c.Phone,
		&c.Position,
		&c.Location,
		&c.Experience,
		&c.CurrentCTC,
		&c.ExpectedCTC,
		&c.NoticePeriod,
		&c.JobDescription,
		&c.Status,
		&c.CreatedAt,
		&c.TenantID,
		&c.CompanyName,
	)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Candidate not found")
		return
	}

	if err := loadCandidateExpertise(&c, tenantID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error loading candidate expertise")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidate retrieved successfully",
		Data:    c,
	})
}

// AddCandidate creates a candidate under the authenticated tenant.
func AddCandidate(w http.ResponseWriter, r *http.Request) {
	var c models.Candidate

	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	companyName, _ := r.Context().Value("companyName").(string)

	if c.Status == "" {
		c.Status = "active"
	}

	if !isValidCandidateStatus(c.Status) {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate status")
		return
	}

	c.TenantID = tenantID
	c.CompanyName = companyName

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error starting transaction")
		return
	}
	defer tx.Rollback()

	err = tx.QueryRow(`
		INSERT INTO candidates (
			name, email, phone, position, location, experience,
			currentctc, expectedctc, noticeperiod, jobdescription,
			status, tenant_id, company_name
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13
		)
		RETURNING id, created_at`,
		c.Name,
		c.Email,
		c.Phone,
		c.Position,
		c.Location,
		c.Experience,
		c.CurrentCTC,
		c.ExpectedCTC,
		c.NoticePeriod,
		c.JobDescription,
		c.Status,
		c.TenantID,
		c.CompanyName,
	).Scan(&c.ID, &c.CreatedAt)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating candidate")
		return
	}

	if err := insertCandidateExpertise(tx, &c); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error committing candidate")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Candidate created successfully",
		Data:    c,
	})
}

// UpdateCandidate updates an existing candidate scoped to the authenticated
// tenant.
func UpdateCandidate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	var c models.Candidate

	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	var existingStatus string

	err = db.DB.QueryRow(`
		SELECT status
		FROM candidates
		WHERE id = $1 AND tenant_id = $2`,
		id,
		tenantID,
	).Scan(&existingStatus)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Candidate not found")
		return
	}

	c.ID = id
	c.TenantID = tenantID

	if c.Status == "" {
		c.Status = existingStatus
	}

	if !isValidCandidateStatus(c.Status) {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate status")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error starting transaction")
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		UPDATE candidates
		SET name = $1,
		    email = $2,
		    phone = $3,
		    position = $4,
		    location = $5,
		    experience = $6,
		    currentctc = $7,
		    expectedctc = $8,
		    noticeperiod = $9,
		    jobdescription = $10,
		    status = $11
		WHERE id = $12 AND tenant_id = $13`,
		c.Name,
		c.Email,
		c.Phone,
		c.Position,
		c.Location,
		c.Experience,
		c.CurrentCTC,
		c.ExpectedCTC,
		c.NoticePeriod,
		c.JobDescription,
		c.Status,
		id,
		tenantID,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating candidate")
		return
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		respondWithError(w, http.StatusNotFound, "Candidate not found")
		return
	}

	// Expertise collections are replaced only when supplied by the request.
	// This allows ordinary candidate updates to leave expertise untouched.
	if c.LanguageExpertise != nil || c.TechnicalExpertise != nil {
		if err := replaceCandidateExpertise(tx, &c); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error committing candidate")
		return
	}

	if err := loadCandidateExpertise(&c, tenantID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error loading candidate expertise")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Candidate updated successfully",
		Data:    c,
	})
}

// DeleteCandidate deletes a candidate scoped to the authenticated tenant.
func DeleteCandidate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	result, err := db.DB.Exec(`
		DELETE FROM candidates
		WHERE id = $1 AND tenant_id = $2`,
		id,
		tenantID,
	)

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

func loadCandidateExpertise(c *models.Candidate, tenantID string) error {
	c.LanguageExpertise = make([]models.CandidateLanguageExpertise, 0)
	c.TechnicalExpertise = make([]models.CandidateExpertise, 0)

	languageRows, err := db.DB.Query(`
		SELECT id, tenant_id, candidate_id, language,
		       proficiency_framework, proficiency_level,
		       created_at, updated_at
		FROM candidate_language_expertise
		WHERE candidate_id = $1 AND tenant_id = $2
		ORDER BY id`,
		c.ID,
		tenantID,
	)
	if err != nil {
		return err
	}
	defer languageRows.Close()

	for languageRows.Next() {
		var item models.CandidateLanguageExpertise

		if err := languageRows.Scan(
			&item.ID,
			&item.TenantID,
			&item.CandidateID,
			&item.Language,
			&item.ProficiencyFramework,
			&item.ProficiencyLevel,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return err
		}

		c.LanguageExpertise = append(c.LanguageExpertise, item)
	}

	if err := languageRows.Err(); err != nil {
		return err
	}

	expertiseRows, err := db.DB.Query(`
		SELECT id, tenant_id, candidate_id, skill, category,
		       proficiency_level, created_at, updated_at
		FROM candidate_expertise
		WHERE candidate_id = $1 AND tenant_id = $2
		ORDER BY id`,
		c.ID,
		tenantID,
	)
	if err != nil {
		return err
	}
	defer expertiseRows.Close()

	for expertiseRows.Next() {
		var item models.CandidateExpertise

		if err := expertiseRows.Scan(
			&item.ID,
			&item.TenantID,
			&item.CandidateID,
			&item.Skill,
			&item.Category,
			&item.ProficiencyLevel,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return err
		}

		c.TechnicalExpertise = append(c.TechnicalExpertise, item)
	}

	return expertiseRows.Err()
}

func insertCandidateExpertise(tx *sql.Tx, c *models.Candidate) error {
	for _, item := range c.LanguageExpertise {
		if strings.TrimSpace(item.Language) == "" ||
			strings.TrimSpace(item.ProficiencyFramework) == "" ||
			strings.TrimSpace(item.ProficiencyLevel) == "" {
			return errors.New("language expertise requires language, proficiencyFramework, and proficiencyLevel")
		}

		_, err := tx.Exec(`
			INSERT INTO candidate_language_expertise (
				tenant_id,
				candidate_id,
				language,
				proficiency_framework,
				proficiency_level
			)
			VALUES ($1, $2, $3, $4, $5)`,
			c.TenantID,
			c.ID,
			strings.TrimSpace(item.Language),
			strings.TrimSpace(item.ProficiencyFramework),
			strings.TrimSpace(item.ProficiencyLevel),
		)
		if err != nil {
			return err
		}
	}

	for _, item := range c.TechnicalExpertise {
		if strings.TrimSpace(item.Skill) == "" ||
			strings.TrimSpace(item.Category) == "" ||
			strings.TrimSpace(item.ProficiencyLevel) == "" {
			return errors.New("technical expertise requires skill, category, and proficiencyLevel")
		}

		_, err := tx.Exec(`
			INSERT INTO candidate_expertise (
				tenant_id,
				candidate_id,
				skill,
				category,
				proficiency_level
			)
			VALUES ($1, $2, $3, $4, $5)`,
			c.TenantID,
			c.ID,
			strings.TrimSpace(item.Skill),
			strings.TrimSpace(item.Category),
			strings.TrimSpace(item.ProficiencyLevel),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func replaceCandidateExpertise(tx *sql.Tx, c *models.Candidate) error {
	if c.LanguageExpertise != nil {
		if _, err := tx.Exec(`
			DELETE FROM candidate_language_expertise
			WHERE candidate_id = $1 AND tenant_id = $2`,
			c.ID,
			c.TenantID,
		); err != nil {
			return err
		}

		for _, item := range c.LanguageExpertise {
			if strings.TrimSpace(item.Language) == "" ||
				strings.TrimSpace(item.ProficiencyFramework) == "" ||
				strings.TrimSpace(item.ProficiencyLevel) == "" {
				return errors.New("language expertise requires language, proficiencyFramework, and proficiencyLevel")
			}

			if _, err := tx.Exec(`
				INSERT INTO candidate_language_expertise (
					tenant_id,
					candidate_id,
					language,
					proficiency_framework,
					proficiency_level
				)
				VALUES ($1, $2, $3, $4, $5)`,
				c.TenantID,
				c.ID,
				strings.TrimSpace(item.Language),
				strings.TrimSpace(item.ProficiencyFramework),
				strings.TrimSpace(item.ProficiencyLevel),
			); err != nil {
				return err
			}
		}
	}

	if c.TechnicalExpertise != nil {
		if _, err := tx.Exec(`
			DELETE FROM candidate_expertise
			WHERE candidate_id = $1 AND tenant_id = $2`,
			c.ID,
			c.TenantID,
		); err != nil {
			return err
		}

		for _, item := range c.TechnicalExpertise {
			if strings.TrimSpace(item.Skill) == "" ||
				strings.TrimSpace(item.Category) == "" ||
				strings.TrimSpace(item.ProficiencyLevel) == "" {
				return errors.New("technical expertise requires skill, category, and proficiencyLevel")
			}

			if _, err := tx.Exec(`
				INSERT INTO candidate_expertise (
					tenant_id,
					candidate_id,
					skill,
					category,
					proficiency_level
				)
				VALUES ($1, $2, $3, $4, $5)`,
				c.TenantID,
				c.ID,
				strings.TrimSpace(item.Skill),
				strings.TrimSpace(item.Category),
				strings.TrimSpace(item.ProficiencyLevel),
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func isValidCandidateStatus(status string) bool {
	switch status {
	case "active", "inactive", "blacklisted", "archived":
		return true
	default:
		return false
	}
}
