package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetJobs retrieves all jobs for the authenticated tenant. Scoped by
// tenant_id (ADR 0001), not company_name.
func GetJobs(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)

	jobs := []models.Job{}
	rows, err := db.DB.Query(`
		SELECT id, title, department, location, status, date_posted,
			description, requirements, last_modified, tenant_id, company_name,
			COALESCE(created_by_user_id, 0)
		FROM jobs WHERE tenant_id = $1`, tenantID)
	if err != nil {
		log.Println("GetJobs query error:", err)
		respondWithError(w, http.StatusInternalServerError, "Error fetching jobs")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var j models.Job
		err := rows.Scan(&j.ID, &j.Title, &j.Department, &j.Location,
			&j.Status, &j.DatePosted, &j.Description, &j.Requirements,
			&j.LastModified, &j.TenantID, &j.CompanyName, &j.CreatedByUserID)
		if err != nil {
			log.Println("GetJobs scan error:", err)
			respondWithError(w, http.StatusInternalServerError, "Error scanning job row")
			return
		}
		jobs = append(jobs, j)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Jobs retrieved successfully",
		Data:    jobs,
	})
}

// GetJobByID retrieves a single job by ID, scoped to the authenticated
// tenant. A job ID belonging to another tenant returns 404, identically to
// a nonexistent ID (ADR 0001: must not disclose existence cross-tenant).
func GetJobByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	var job models.Job
	err = db.DB.QueryRow(`
		SELECT id, title, department, location, status, date_posted,
			description, requirements, last_modified, tenant_id, company_name,
			COALESCE(created_by_user_id, 0)
		FROM jobs WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	).Scan(&job.ID, &job.Title, &job.Department, &job.Location,
		&job.Status, &job.DatePosted, &job.Description, &job.Requirements,
		&job.LastModified, &job.TenantID, &job.CompanyName, &job.CreatedByUserID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Job not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Job retrieved successfully",
		Data:    job,
	})
}

// AddJob creates a new job under the authenticated tenant. tenant_id is
// always derived from context, never from the request payload.
func AddJob(w http.ResponseWriter, r *http.Request) {
	var job models.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	job.TenantID = r.Context().Value("tenantID").(string)
	job.CompanyName = r.Context().Value("companyName").(string)

	// Record which manager/admin created this job (docs/architecture.md section 13.4)
	job.CreatedByUserID = r.Context().Value("userID").(int)

	var id int
	err = db.DB.QueryRow(
		`INSERT INTO jobs (title, department, location, status, 
			description, requirements, tenant_id, company_name, created_by_user_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id`,
		job.Title, job.Department, job.Location, job.Status,
		job.Description, job.Requirements, job.TenantID, job.CompanyName, job.CreatedByUserID,
	).Scan(&id)

	if err != nil {
		log.Println("AddJob insert error:", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating job")
		return
	}

	job.ID = id

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Job created successfully",
		Data:    job,
	})
}

// UpdateJob updates an existing job, scoped to the authenticated tenant.
// A job ID belonging to another tenant affects zero rows.
func UpdateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	var job models.Job
	err = json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	tenantID := r.Context().Value("tenantID").(string)
	job.TenantID = tenantID
	job.ID = id

	result, err := db.DB.Exec(
		`UPDATE jobs 
		SET title = $1, department = $2, location = $3, status = $4, 
			description = $5, requirements = $6, last_modified = NOW() 
		WHERE id = $7 AND tenant_id = $8`,
		job.Title, job.Department, job.Location, job.Status,
		job.Description, job.Requirements, job.ID, tenantID,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating job")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Job not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Job updated successfully",
		Data:    job,
	})
}

// DeleteJob deletes a job, scoped to the authenticated tenant.
func DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	result, err := db.DB.Exec(
		"DELETE FROM jobs WHERE id = $1 AND tenant_id = $2",
		id, tenantID,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting job")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Job not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Job deleted successfully",
	})
}
