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

var validClientStatuses = map[string]bool{
	"prospect": true,
	"active":   true,
	"inactive": true,
}

// GetClients retrieves a bounded, deterministically ordered client page for
// the authenticated tenant. The optional status filter is also tenant-scoped.
func GetClients(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)
	pagination, err := parsePagination(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := r.URL.Query().Get("status")
	if status != "" && !validClientStatuses[status] {
		respondWithError(w, http.StatusBadRequest, "status must be one of prospect, active, inactive")
		return
	}

	var total int
	if status == "" {
		err = db.DB.QueryRow(`SELECT COUNT(*) FROM clients WHERE tenant_id = $1`, tenantID).Scan(&total)
	} else {
		err = db.DB.QueryRow(`SELECT COUNT(*) FROM clients WHERE tenant_id = $1 AND status = $2`, tenantID, status).Scan(&total)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error counting clients")
		return
	}

	var rowsQuery string
	var rowsArgs []interface{}
	if status == "" {
		rowsQuery = `SELECT id, name, status, COALESCE(contact_email, ''), COALESCE(contact_phone, ''), created_at, updated_at, tenant_id
			FROM clients WHERE tenant_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`
		rowsArgs = []interface{}{tenantID, pagination.Limit, paginationOffset(pagination)}
	} else {
		rowsQuery = `SELECT id, name, status, COALESCE(contact_email, ''), COALESCE(contact_phone, ''), created_at, updated_at, tenant_id
			FROM clients WHERE tenant_id = $1 AND status = $2 ORDER BY created_at DESC, id DESC LIMIT $3 OFFSET $4`
		rowsArgs = []interface{}{tenantID, status, pagination.Limit, paginationOffset(pagination)}
	}

	rows, err := db.DB.Query(rowsQuery, rowsArgs...)
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
	if err := rows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading clients")
		return
	}

	respondWithJSON(w, http.StatusOK, PaginatedResponse{
		Success: true,
		Message: "Clients retrieved successfully",
		Data:    clients,
		Pagination: Pagination{
			Page:  pagination.Page,
			Limit: pagination.Limit,
			Total: total,
		},
	})
}

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
	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Client retrieved successfully", Data: c})
}

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
		RETURNING id, created_at, updated_at`, c.Name, c.Status, c.ContactEmail, c.ContactPhone, c.TenantID).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating client")
		return
	}
	respondWithJSON(w, http.StatusCreated, models.ApiResponse{Success: true, Message: "Client created successfully", Data: c})
}

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
	c.ID, c.TenantID = id, tenantID
	result, err := db.DB.Exec(`
		UPDATE clients SET name = $1, status = $2, contact_email = $3, contact_phone = $4, updated_at = NOW()
		WHERE id = $5 AND tenant_id = $6`, c.Name, c.Status, c.ContactEmail, c.ContactPhone, c.ID, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating client")
		return
	}
	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		respondWithError(w, http.StatusNotFound, "Client not found")
		return
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Client updated successfully", Data: c})
}

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
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
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
	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Client deleted successfully"})
}
