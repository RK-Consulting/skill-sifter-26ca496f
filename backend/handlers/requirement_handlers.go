package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// validRequirementStatuses is the ADR 0002 Requirement lifecycle:
// Draft -> Open -> On Hold -> Filled, with a Cancelled branch from any
// active state. This handler does not enforce transition order (e.g.
// rejecting Draft -> Filled directly) — ADR 0002 requires only that the
// lifecycle "distinguish an active requirement from one that is paused,
// fulfilled, or cancelled", not a strict state machine. A stricter
// transition policy would be Recruitment Assignment / workflow territory
// (Issue #18), out of scope here.
var validRequirementStatuses = map[string]bool{
	"draft":     true,
	"open":      true,
	"on_hold":   true,
	"filled":    true,
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
		SELECT id, client_id, title, COALESCE(department, ''), COALESCE(location, ''),
			COALESCE(work_arrangement, ''), status, COALESCE(opened_date, created_at),
			COALESCE(description, ''), COALESCE(required_skills, ''), COALESCE(experience_required, ''),
			COALESCE(compensation, ''), headcount, COALESCE(language_requirement, ''),
			created_at, last_modified, tenant_id
		FROM requirements WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching requirements")
		return
	}
	defer rows.Close()

	requirements := []models.Requirement{}
	for rows.Next() {
		var req models.Requirement
		err := rows.Scan(&req.ID, &req.ClientID, &req.Title, &req.Department, &req.Location,
			&req.WorkArrangement, &req.Status, &req.OpenedDate, &req.Description, &req.RequiredSkills,
			&req.ExperienceRequired, &req.Compensation, &req.Headcount, &req.LanguageRequirement,
			&req.CreatedAt, &req.LastModified, &req.TenantID)
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
		SELECT id, client_id, title, COALESCE(department, ''), COALESCE(location, ''),
			COALESCE(work_arrangement, ''), status, COALESCE(opened_date, created_at),
			COALESCE(description, ''), COALESCE(required_skills, ''), COALESCE(experience_required, ''),
			COALESCE(compensation, ''), headcount, COALESCE(language_requirement, ''),
			created_at, last_modified, tenant_id
		FROM requirements WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	).Scan(&req.ID, &req.ClientID, &req.Title, &req.Department, &req.Location,
		&req.WorkArrangement, &req.Status, &req.OpenedDate, &req.Description, &req.RequiredSkills,
		&req.ExperienceRequired, &req.Compensation, &req.Headcount, &req.LanguageRequirement,
		&req.CreatedAt, &req.LastModified, &req.TenantID)

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

	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Requirement title is required")
		return
	}
	if req.ClientID == 0 {
		respondWithError(w, http.StatusBadRequest, "clientId is required")
		return
	}
	if req.Status == "" {
		req.Status = "draft"
	} else if !validRequirementStatuses[req.Status] {
		respondWithError(w, http.StatusUnprocessableEntity, "Invalid requirement status: must be one of draft, open, on_hold, filled, cancelled")
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
		INSERT INTO requirements (client_id, title, department, location, work_arrangement, status,
			opened_date, description, required_skills, experience_required, compensation, headcount,
			language_requirement, tenant_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, last_modified`,
		req.ClientID, req.Title, req.Department, req.Location, req.WorkArrangement, req.Status,
		openedDate, req.Description, req.RequiredSkills, req.ExperienceRequired, req.Compensation,
		req.Headcount, req.LanguageRequirement, req.TenantID,
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

	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Requirement title is required")
		return
	}
	if req.ClientID == 0 {
		respondWithError(w, http.StatusBadRequest, "clientId is required")
		return
	}
	if !validRequirementStatuses[req.Status] {
		respondWithError(w, http.StatusUnprocessableEntity, "Invalid requirement status: must be one of draft, open, on_hold, filled, cancelled")
		return
	}
	if req.Headcount <= 0 {
		respondWithError(w, http.StatusUnprocessableEntity, "headcount must be positive")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)
	req.ID = id
	req.TenantID = tenantID

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
		UPDATE requirements SET client_id = $1, title = $2, department = $3, location = $4,
			work_arrangement = $5, status = $6, opened_date = $7, description = $8,
			required_skills = $9, experience_required = $10, compensation = $11, headcount = $12,
			language_requirement = $13, last_modified = NOW()
		WHERE id = $14 AND tenant_id = $15`,
		req.ClientID, req.Title, req.Department, req.Location, req.WorkArrangement, req.Status,
		openedDate, req.Description, req.RequiredSkills, req.ExperienceRequired, req.Compensation,
		req.Headcount, req.LanguageRequirement, req.ID, tenantID,
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
