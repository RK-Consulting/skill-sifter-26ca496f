package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

// validClientStatuses is the ADR 0002 Client lifecycle: Prospect -> Active -> Inactive.
var validClientStatuses = map[string]bool{
	"prospect": true,
	"active":   true,
	"inactive": true,
}

// GetClients retrieves all clients for the authenticated tenant. Scoped by
// tenant_id (ADR 0001), not company_name.
func GetClients(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)

	rows, err := db.DB.Query(`
		SELECT id, name, status, COALESCE(contact_email, ''), COALESCE(contact_phone, ''), created_at, updated_at, tenant_id
		FROM clients WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching clients")
		return
	}
	defer rows.Close()

	clients := []models.Client{}
	for rows.Next() {
		var c models.Client
		if err := rows.Scan(&c.ID, &c.Name, &c.Status, &c.ContactEmail, &c.ContactPhone, &c.CreatedAt, &c.UpdatedAt, &c.TenantID); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning client row")
			return
		}
		clients = append(clients, c)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Clients retrieved successfully",
		Data:    clients,
	})
}

// GetClientByID retrieves a single client by ID, scoped to the
// authenticated tenant. A client ID belonging to another tenant returns
// 404, identically to a nonexistent ID (ADR 0001/0002).
func GetClientByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	var c models.Client
	err = db.DB.QueryRow(`
		SELECT id, name, status, COALESCE(contact_email, ''), COALESCE(contact_phone, ''), created_at, updated_at, tenant_id
		FROM clients WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	).Scan(&c.ID, &c.Name, &c.Status, &c.ContactEmail, &c.ContactPhone, &c.CreatedAt, &c.UpdatedAt, &c.TenantID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Client not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Client retrieved successfully",
		Data:    c,
	})
}

// AddClient creates a new client under the authenticated tenant. tenant_id
// is always derived from context, never from the request payload. Status
// defaults to "prospect" (ADR 0002 lifecycle start) if not provided or
// invalid.
func AddClient(w http.ResponseWriter, r *http.Request) {
	var c models.Client
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if c.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Client name is required")
		return
	}

	if c.Status == "" {
		c.Status = "prospect"
	} else if !validClientStatuses[c.Status] {
		respondWithError(w, http.StatusUnprocessableEntity, "Invalid client status: must be one of prospect, active, inactive")
		return
	}

	c.TenantID = r.Context().Value("tenantID").(string)

	err := db.DB.QueryRow(`
		INSERT INTO clients (name, status, contact_email, contact_phone, tenant_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`,
		c.Name, c.Status, c.ContactEmail, c.ContactPhone, c.TenantID,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating client")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Client created successfully",
		Data:    c,
	})
}

// UpdateClient updates an existing client, scoped to the authenticated
// tenant. A client ID belonging to another tenant affects zero rows.
func UpdateClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	var c models.Client
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if c.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Client name is required")
		return
	}
	if !validClientStatuses[c.Status] {
		respondWithError(w, http.StatusUnprocessableEntity, "Invalid client status: must be one of prospect, active, inactive")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)
	c.ID = id
	c.TenantID = tenantID

	result, err := db.DB.Exec(`
		UPDATE clients SET name = $1, status = $2, contact_email = $3, contact_phone = $4, updated_at = NOW()
		WHERE id = $5 AND tenant_id = $6`,
		c.Name, c.Status, c.ContactEmail, c.ContactPhone, c.ID, tenantID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating client")
		return
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		respondWithError(w, http.StatusNotFound, "Client not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Client updated successfully",
		Data:    c,
	})
}

// DeleteClient deletes a client, scoped to the authenticated tenant. A
// client with existing requirements cannot be deleted (the requirements
// foreign key has no ON DELETE CASCADE, so Postgres will reject this with
// a foreign-key-violation error, which is intentional: silently cascading
// deletes into a client's requirements would destroy recruitment demand
// history without an explicit decision to do so).
func DeleteClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	result, err := db.DB.Exec(`DELETE FROM clients WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" { // foreign_key_violation
			respondWithError(w, http.StatusConflict, "Cannot delete client: it has existing requirements. Remove or reassign them first.")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error deleting client")
		return
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		respondWithError(w, http.StatusNotFound, "Client not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Client deleted successfully",
	})
}
