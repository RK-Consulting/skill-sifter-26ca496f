package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetBusinessDevs handles fetching all business development records for the
// authenticated tenant. Scoped by tenant_id (ADR 0001), not company_name.
func GetBusinessDevs(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Tenant not found in context")
		return
	}
	companyName, _ := r.Context().Value("companyName").(string)

	query := "SELECT id, client_name, partner_name, contact_person, contact_number, contact_email, created_at, last_modified FROM business_dev WHERE tenant_id = $1 ORDER BY created_at DESC"

	var tableExists bool
	tableCheckErr := db.DB.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'business_dev')").Scan(&tableExists)
	if tableCheckErr != nil {
		fmt.Printf("Error checking if business_dev table exists: %v\n", tableCheckErr)
		respondWithError(w, http.StatusInternalServerError, "Error checking database schema")
		return
	}

	if !tableExists {
		fmt.Println("business_dev table does not exist!")
		respondWithJSON(w, http.StatusOK, models.ApiResponse{
			Success: true,
			Message: "Business development records fetched (table doesn't exist yet)",
			Data:    []models.BusinessDev{},
		})
		return
	}

	rows, err := db.DB.Query(query, tenantID)
	if err != nil {
		fmt.Printf("Database error in GetBusinessDevs: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Error querying business development records")
		return
	}
	defer rows.Close()

	var businessDevs []models.BusinessDev
	for rows.Next() {
		var b models.BusinessDev
		if err := rows.Scan(&b.ID, &b.ClientName, &b.PartnerName, &b.ContactPerson, &b.ContactNumber, &b.ContactEmail, &b.CreatedAt, &b.LastModified); err != nil {
			fmt.Printf("Row scan error in GetBusinessDevs: %v\n", err)
			respondWithError(w, http.StatusInternalServerError, "Error scanning business dev record")
			return
		}
		b.TenantID = tenantID
		b.CompanyName = companyName
		businessDevs = append(businessDevs, b)
	}

	if err = rows.Err(); err != nil {
		fmt.Printf("Rows iteration error in GetBusinessDevs: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Error iterating business dev records")
		return
	}

	if businessDevs == nil {
		businessDevs = []models.BusinessDev{}
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Business development records fetched successfully",
		Data:    businessDevs,
	})
}

// GetBusinessDevByID handles fetching a specific business dev record,
// scoped to the authenticated tenant. A record belonging to another tenant
// returns 404, identically to a nonexistent ID (ADR 0001).
func GetBusinessDevByID(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Tenant not found in context")
		return
	}
	companyName, _ := r.Context().Value("companyName").(string)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid business dev ID")
		return
	}

	var tableExists bool
	tableCheckErr := db.DB.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'business_dev')").Scan(&tableExists)
	if tableCheckErr != nil {
		fmt.Printf("Error checking if business_dev table exists: %v\n", tableCheckErr)
		respondWithError(w, http.StatusInternalServerError, "Error checking database schema")
		return
	}

	if !tableExists {
		respondWithError(w, http.StatusNotFound, "Business dev record not found (table doesn't exist)")
		return
	}

	var b models.BusinessDev
	err = db.DB.QueryRow("SELECT id, client_name, partner_name, contact_person, contact_number, contact_email, created_at, last_modified FROM business_dev WHERE id = $1 AND tenant_id = $2", id, tenantID).
		Scan(&b.ID, &b.ClientName, &b.PartnerName, &b.ContactPerson, &b.ContactNumber, &b.ContactEmail, &b.CreatedAt, &b.LastModified)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Business dev record not found")
		return
	}

	b.TenantID = tenantID
	b.CompanyName = companyName

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Business dev record fetched successfully",
		Data:    b,
	})
}

// AddBusinessDev handles creating a new business development record under
// the authenticated tenant. tenant_id is always derived from context, never
// from the request payload.
func AddBusinessDev(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Tenant not found in context")
		return
	}
	companyName, _ := r.Context().Value("companyName").(string)

	var b models.BusinessDev
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&b); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	var tableExists bool
	tableCheckErr := db.DB.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'business_dev')").Scan(&tableExists)
	if tableCheckErr != nil {
		fmt.Printf("Error checking if business_dev table exists: %v\n", tableCheckErr)
		respondWithError(w, http.StatusInternalServerError, "Error checking database schema")
		return
	}

	if !tableExists {
		// Pre-existing lazy-create fallback (not introduced by Issue #33).
		// tenant_id is included so this fallback path — if it is ever
		// actually exercised — does not recreate a table missing the
		// isolation column added by migration 006. In practice this table
		// already exists via migrations/001_baseline.sql, so this branch
		// should not run in a correctly migrated environment.
		_, err := db.DB.Exec(`
			CREATE TABLE IF NOT EXISTS business_dev (
				id SERIAL PRIMARY KEY,
				client_name VARCHAR(255) NOT NULL,
				partner_name VARCHAR(255),
				contact_person VARCHAR(255) NOT NULL,
				contact_number VARCHAR(50),
				contact_email VARCHAR(255) NOT NULL,
				created_at TIMESTAMP DEFAULT NOW(),
				last_modified TIMESTAMP DEFAULT NOW(),
				tenant_id VARCHAR(255) REFERENCES companies(id),
				company_name VARCHAR(255) NOT NULL
			)
		`)
		if err != nil {
			fmt.Printf("Error creating business_dev table: %v\n", err)
			respondWithError(w, http.StatusInternalServerError, "Error creating business_dev table")
			return
		}

		_, err = db.DB.Exec("CREATE INDEX IF NOT EXISTS idx_business_dev_tenant ON business_dev(tenant_id)")
		if err != nil {
			fmt.Printf("Error creating business_dev index: %v\n", err)
		}
	}

	b.TenantID = tenantID
	b.CompanyName = companyName
	b.CreatedAt = time.Now()
	b.LastModified = time.Now()

	var id int
	err := db.DB.QueryRow(
		"INSERT INTO business_dev (client_name, partner_name, contact_person, contact_number, contact_email, tenant_id, company_name, created_at, last_modified) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		b.ClientName, b.PartnerName, b.ContactPerson, b.ContactNumber, b.ContactEmail, b.TenantID, b.CompanyName, b.CreatedAt, b.LastModified,
	).Scan(&id)
	if err != nil {
		fmt.Printf("Error creating business dev record: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating business dev record")
		return
	}

	b.ID = id

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Business dev record created successfully",
		Data:    b,
	})
}

// UpdateBusinessDev handles updating an existing business development
// record, scoped to the authenticated tenant. A record belonging to another
// tenant affects zero rows.
func UpdateBusinessDev(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Tenant not found in context")
		return
	}
	companyName, _ := r.Context().Value("companyName").(string)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid business dev ID")
		return
	}

	var tableExists bool
	tableCheckErr := db.DB.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'business_dev')").Scan(&tableExists)
	if tableCheckErr != nil {
		fmt.Printf("Error checking if business_dev table exists: %v\n", tableCheckErr)
		respondWithError(w, http.StatusInternalServerError, "Error checking database schema")
		return
	}

	if !tableExists {
		respondWithError(w, http.StatusNotFound, "Business dev record not found (table doesn't exist)")
		return
	}

	var b models.BusinessDev
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&b); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	b.TenantID = tenantID
	b.CompanyName = companyName
	b.ID = id
	b.LastModified = time.Now()

	result, err := db.DB.Exec(
		"UPDATE business_dev SET client_name = $1, partner_name = $2, contact_person = $3, contact_number = $4, contact_email = $5, last_modified = $6 WHERE id = $7 AND tenant_id = $8",
		b.ClientName, b.PartnerName, b.ContactPerson, b.ContactNumber, b.ContactEmail, b.LastModified, b.ID, tenantID,
	)
	if err != nil {
		fmt.Printf("Error updating business dev record: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Error updating business dev record")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error checking update result")
		return
	}
	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Business dev record not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Business dev record updated successfully",
		Data:    b,
	})
}

// DeleteBusinessDev handles deleting a business development record, scoped
// to the authenticated tenant.
func DeleteBusinessDev(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Tenant not found in context")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid business dev ID")
		return
	}

	var tableExists bool
	tableCheckErr := db.DB.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'business_dev')").Scan(&tableExists)
	if tableCheckErr != nil {
		fmt.Printf("Error checking if business_dev table exists: %v\n", tableCheckErr)
		respondWithError(w, http.StatusInternalServerError, "Error checking database schema")
		return
	}

	if !tableExists {
		respondWithError(w, http.StatusNotFound, "Business dev record not found (table doesn't exist)")
		return
	}

	result, err := db.DB.Exec("DELETE FROM business_dev WHERE id = $1 AND tenant_id = $2", id, tenantID)
	if err != nil {
		fmt.Printf("Error deleting business dev record: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Error deleting business dev record")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error checking delete result")
		return
	}
	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Business dev record not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Business dev record deleted successfully",
	})
}
