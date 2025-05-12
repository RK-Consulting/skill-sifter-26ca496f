
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetBusinessDevs handles fetching all business development records for a company
func GetBusinessDevs(w http.ResponseWriter, r *http.Request) {
	// Get company from context (set by auth middleware)
	companyName, ok := r.Context().Value("companyName").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	// Query database
	rows, err := db.DB.Query("SELECT id, client_name, partner_name, contact_person, contact_number, contact_email, created_at, last_modified FROM business_dev WHERE company_name = $1 ORDER BY created_at DESC", companyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching business development records")
		return
	}
	defer rows.Close()

	var businessDevs []models.BusinessDev
	for rows.Next() {
		var b models.BusinessDev
		if err := rows.Scan(&b.ID, &b.ClientName, &b.PartnerName, &b.ContactPerson, &b.ContactNumber, &b.ContactEmail, &b.CreatedAt, &b.LastModified); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning business dev record")
			return
		}
		b.CompanyName = companyName
		businessDevs = append(businessDevs, b)
	}

	// Return response
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Business development records fetched successfully",
		Data:    businessDevs,
	})
}

// GetBusinessDevByID handles fetching a specific business dev record
func GetBusinessDevByID(w http.ResponseWriter, r *http.Request) {
	// Get company from context (set by auth middleware)
	companyName, ok := r.Context().Value("companyName").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	// Get ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid business dev ID")
		return
	}

	// Query database
	var b models.BusinessDev
	err = db.DB.QueryRow("SELECT id, client_name, partner_name, contact_person, contact_number, contact_email, created_at, last_modified FROM business_dev WHERE id = $1 AND company_name = $2", id, companyName).
		Scan(&b.ID, &b.ClientName, &b.PartnerName, &b.ContactPerson, &b.ContactNumber, &b.ContactEmail, &b.CreatedAt, &b.LastModified)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Business dev record not found")
		return
	}

	b.CompanyName = companyName

	// Return response
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Business dev record fetched successfully",
		Data:    b,
	})
}

// AddBusinessDev handles creating a new business development record
func AddBusinessDev(w http.ResponseWriter, r *http.Request) {
	// Get company from context (set by auth middleware)
	companyName, ok := r.Context().Value("companyName").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	// Decode request body
	var b models.BusinessDev
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&b); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Set company name from context
	b.CompanyName = companyName
	b.CreatedAt = time.Now()
	b.LastModified = time.Now()

	// Insert into database
	var id int
	err := db.DB.QueryRow(
		"INSERT INTO business_dev (client_name, partner_name, contact_person, contact_number, contact_email, company_name, created_at, last_modified) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		b.ClientName, b.PartnerName, b.ContactPerson, b.ContactNumber, b.ContactEmail, b.CompanyName, b.CreatedAt, b.LastModified,
	).Scan(&id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating business dev record")
		return
	}

	b.ID = id

	// Return response
	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Business dev record created successfully",
		Data:    b,
	})
}

// UpdateBusinessDev handles updating an existing business development record
func UpdateBusinessDev(w http.ResponseWriter, r *http.Request) {
	// Get company from context (set by auth middleware)
	companyName, ok := r.Context().Value("companyName").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	// Get ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid business dev ID")
		return
	}

	// Decode request body
	var b models.BusinessDev
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&b); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Set company name and ID from context/URL
	b.CompanyName = companyName
	b.ID = id
	b.LastModified = time.Now()

	// Update database
	result, err := db.DB.Exec(
		"UPDATE business_dev SET client_name = $1, partner_name = $2, contact_person = $3, contact_number = $4, contact_email = $5, last_modified = $6 WHERE id = $7 AND company_name = $8",
		b.ClientName, b.PartnerName, b.ContactPerson, b.ContactNumber, b.ContactEmail, b.LastModified, b.ID, b.CompanyName,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating business dev record")
		return
	}

	// Check if record existed
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error checking update result")
		return
	}
	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Business dev record not found")
		return
	}

	// Return response
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Business dev record updated successfully",
		Data:    b,
	})
}

// DeleteBusinessDev handles deleting a business development record
func DeleteBusinessDev(w http.ResponseWriter, r *http.Request) {
	// Get company from context (set by auth middleware)
	companyName, ok := r.Context().Value("companyName").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Company not found in context")
		return
	}

	// Get ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid business dev ID")
		return
	}

	// Delete from database
	result, err := db.DB.Exec("DELETE FROM business_dev WHERE id = $1 AND company_name = $2", id, companyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting business dev record")
		return
	}

	// Check if record existed
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error checking delete result")
		return
	}
	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Business dev record not found")
		return
	}

	// Return response
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Business dev record deleted successfully",
	})
}
