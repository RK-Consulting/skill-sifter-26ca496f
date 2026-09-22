package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// Requirement lifecycle exposed by the application.
// open -> on_hold -> closed is the normal business flow; cancelled is
// available when the requirement is no longer being pursued.
var validRequirementStatuses = map[string]bool{
	"open":      true,
	"closed":    true,
	"on_hold":   true,
	"cancelled": true,
}

// clientBelongsToTenant checks that a client_id exists and belongs to the
// authenticated tenant, so a requirement can never be attached to another
// tenant's client (which would otherwise let a requirement's tenant_id and
// its client's actual tenant silently diverge).
func clientBelongsToTenant(clientID int, tenantID string) (bool, error) {
	var exists bool
	err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1 AND tenant_id = $2)`, clientID, tenantID).Scan(&exists)
	return exists, err
}

// GetRequirements retrieves all requirements for the authenticated tenant.
// Scoped by tenant_id (ADR 0001), not company_name.
func GetRequirements(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)

	rows, err := db.DB.Query(`
		SELECT id, client_id, COALESCE(job_id, ''), COALESCE(job_type, ''), title, COALESCE(department, ''),
			COALESCE(experience_required, ''), COALESCE(budget, ''), COALESCE(language_requirement, ''),
			COALESCE(certifications_required, ''), COALESCE(notice_period, ''),
			COALESCE(work_arrangement, ''), COALESCE(mandatory_requirements, ''),
			COALESCE(description, ''), status, COALESCE(location, ''),
			headcount, COALESCE(opened_date, created_at), created_at, last_modified, tenant_id
		FROM requirements WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching requirements")
		return
	}
	defer rows.Close()

	requirements := []models.Requirement{}
	for rows.Next() {
		var req models.Requirement
		err := rows.Scan(&req.ID, &req.ClientID, &req.JobID, &req.JobType, &req.Title, &req.Department,
			&req.ExperienceRequired, &req.Budget, &req.LanguageRequirements,
			&req.CertificationsRequired, &req.NoticePeriod, &req.WorkArrangement,
			&req.MandatoryRequirements, &req.Description, &req.Status, &req.Location,
			&req.Headcount, &req.OpenedDate, &req.CreatedAt, &req.LastModified, &req.TenantID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning requirement row")
			return
		}
		requirements = append(requirements, req)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Requirements retrieved successfully",
		Data:    requirements,
	})
}

// GetRequirementByID retrieves a single requirement by ID, scoped to the
// authenticated tenant. A requirement ID belonging to another tenant
// returns 404, identically to a nonexistent ID (ADR 0001/0002).
func GetRequirementByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid requirement ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	var req models.Requirement
	err = db.DB.QueryRow(`
		SELECT id, client_id, COALESCE(job_id, ''), COALESCE(job_type, ''), title, COALESCE(department, ''),
			COALESCE(experience_required, ''), COALESCE(budget, ''), COALESCE(language_requirement, ''),
			COALESCE(certifications_required, ''), COALESCE(notice_period, ''),
			COALESCE(work_arrangement, ''), COALESCE(mandatory_requirements, ''),
			COALESCE(description, ''), status, COALESCE(location, ''),
			headcount, COALESCE(opened_date, created_at), created_at, last_modified, tenant_id
		FROM requirements WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	).Scan(&req.ID, &req.ClientID, &req.JobID, &req.JobType, &req.Title, &req.Department,
		&req.ExperienceRequired, &req.Budget, &req.LanguageRequirements,
		&req.CertificationsRequired, &req.NoticePeriod, &req.WorkArrangement,
		&req.MandatoryRequirements, &req.Description, &req.Status, &req.Location,
		&req.Headcount, &req.OpenedDate, &req.CreatedAt, &req.LastModified, &req.TenantID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Requirement not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Requirement retrieved successfully",
		Data:    req,
	})
}

// nullableRequirementField converts Go's zero-value empty string into SQL NULL.
// This preserves database semantics for optional fields, especially enum
// fields such as job_type and work_arrangement.
func nullableRequirementField(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

// AddRequirement creates a new requirement under the authenticated tenant.
// tenant_id is always derived from context, never from the request
// payload. client_id must reference a client belonging to the same
// authenticated tenant — this is validated explicitly rather than relying
// only on the foreign key, since the FK alone would not stop a
// cross-tenant client_id (the referenced clients row would exist, just in
// another tenant).
func AddRequirement(w http.ResponseWriter, r *http.Request) {
	var req models.Requirement
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.JobID == "" {
		respondWithError(w, http.StatusBadRequest, "jobId is required")
		return
	}
	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Requirement title is required")
		return
	}
	if req.ClientID == 0 {
		respondWithError(w, http.StatusBadRequest, "clientId is required")
		return
	}
	if req.Status == "" {
		req.Status = "open"
	} else if !validRequirementStatuses[req.Status] {
		respondWithError(w, http.StatusUnprocessableEntity, "Invalid requirement status: must be one of open, closed, on_hold, cancelled")
		return
	}
	if req.Headcount == 0 {
		req.Headcount = 1
	} else if req.Headcount < 0 {
		respondWithError(w, http.StatusUnprocessableEntity, "headcount must be positive")
		return
	}

	req.TenantID = r.Context().Value("tenantID").(string)

	belongs, err := clientBelongsToTenant(req.ClientID, req.TenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error validating client")
		return
	}
	if !belongs {
		// 404, not 403: must not disclose whether that client ID exists in
		// another tenant (ADR 0001).
		respondWithError(w, http.StatusNotFound, "Client not found")
		return
	}

	var openedDate interface{}
	if !req.OpenedDate.IsZero() {
		openedDate = req.OpenedDate
	}

	err = db.DB.QueryRow(`
		INSERT INTO requirements (client_id, job_id, job_type, title, department, experience_required, budget,
			language_requirement, certifications_required, notice_period, work_arrangement,
			mandatory_requirements, description, status, location, headcount, opened_date, tenant_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, last_modified`,
		req.ClientID, req.JobID, nullableRequirementField(req.JobType), req.Title, nullableRequirementField(req.Department),
		nullableRequirementField(req.ExperienceRequired), nullableRequirementField(req.Budget),
		nullableRequirementField(req.LanguageRequirements), nullableRequirementField(req.CertificationsRequired),
		nullableRequirementField(req.NoticePeriod), nullableRequirementField(req.WorkArrangement),
		nullableRequirementField(req.MandatoryRequirements), nullableRequirementField(req.Description),
		req.Status, nullableRequirementField(req.Location), req.Headcount, openedDate, req.TenantID,
	).Scan(&req.ID, &req.CreatedAt, &req.LastModified)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating requirement")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Requirement created successfully",
		Data:    req,
	})
}

// UpdateRequirement updates an existing requirement, scoped to the
// authenticated tenant. A requirement ID belonging to another tenant
// affects zero rows. If client_id is being changed, the new client must
// also belong to the authenticated tenant.
func UpdateRequirement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid requirement ID")
		return
	}

	var req models.Requirement
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.JobID == "" {
		respondWithError(w, http.StatusBadRequest, "jobId is required")
		return
	}
	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Requirement title is required")
		return
	}
	if req.ClientID == 0 {
		respondWithError(w, http.StatusBadRequest, "clientId is required")
		return
	}
	if !validRequirementStatuses[req.Status] {
		respondWithError(w, http.StatusUnprocessableEntity, "Invalid requirement status: must be one of open, closed, on_hold, cancelled")
		return
	}
	if req.Headcount <= 0 {
		respondWithError(w, http.StatusUnprocessableEntity, "headcount must be positive")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)
	req.ID = id
	req.TenantID = tenantID

	var existingJobID string
	if err := db.DB.QueryRow(`SELECT COALESCE(job_id, '') FROM requirements WHERE id = $1 AND tenant_id = $2`, id, tenantID).Scan(&existingJobID); err != nil {
		respondWithError(w, http.StatusNotFound, "Requirement not found")
		return
	}
	if existingJobID != "" && existingJobID != req.JobID {
		respondWithError(w, http.StatusConflict, "Job ID cannot be changed once assigned")
		return
	}

	belongs, err := clientBelongsToTenant(req.ClientID, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error validating client")
		return
	}
	if !belongs {
		respondWithError(w, http.StatusNotFound, "Client not found")
		return
	}

	var openedDate interface{}
	if !req.OpenedDate.IsZero() {
		openedDate = req.OpenedDate
	}

	result, err := db.DB.Exec(`
		UPDATE requirements SET client_id = $1, job_id = $2, job_type = $3, title = $4, department = $5,
			experience_required = $6, budget = $7, language_requirement = $8,
			certifications_required = $9, notice_period = $10, work_arrangement = $11,
			mandatory_requirements = $12, description = $13, status = $14, location = $15,
			headcount = $16, opened_date = $17, last_modified = NOW()
		WHERE id = $18 AND tenant_id = $19`,
		req.ClientID, req.JobID, nullableRequirementField(req.JobType), req.Title, nullableRequirementField(req.Department),
		nullableRequirementField(req.ExperienceRequired), nullableRequirementField(req.Budget),
		nullableRequirementField(req.LanguageRequirements), nullableRequirementField(req.CertificationsRequired),
		nullableRequirementField(req.NoticePeriod), nullableRequirementField(req.WorkArrangement),
		nullableRequirementField(req.MandatoryRequirements), nullableRequirementField(req.Description),
		req.Status, nullableRequirementField(req.Location), req.Headcount, openedDate, req.ID, tenantID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating requirement")
		return
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		respondWithError(w, http.StatusNotFound, "Requirement not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Requirement updated successfully",
		Data:    req,
	})
}

// DeleteRequirement deletes a requirement, scoped to the authenticated
// tenant.
func DeleteRequirement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid requirement ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	result, err := db.DB.Exec(`DELETE FROM requirements WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting requirement")
		return
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		respondWithError(w, http.StatusNotFound, "Requirement not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Requirement deleted successfully",
	})
}
