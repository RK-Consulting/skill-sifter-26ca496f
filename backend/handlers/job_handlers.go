package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetJobs retrieves all jobs for a company
func GetJobs(w http.ResponseWriter, r *http.Request) {
	// Get company name from context
	companyName := r.Context().Value("companyName").(string)

	jobs := []models.Job{}
	rows, err := db.DB.Query("SELECT * FROM jobs WHERE company_name = $1", companyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching jobs")
		return
	}
	defer rows.Close()

	// Scan rows into jobs slice
	for rows.Next() {
		var j models.Job
		err := rows.Scan(&j.ID, &j.Title, &j.Department, &j.Location,
			&j.Status, &j.DatePosted, &j.Description, &j.Requirements,
			&j.LastModified, &j.CompanyName)
		if err != nil {
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

// GetJobByID retrieves a single job by ID
func GetJobByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Get company name from context
	companyName := r.Context().Value("companyName").(string)

	var job models.Job
	err = db.DB.QueryRow(
		"SELECT * FROM jobs WHERE id = $1 AND company_name = $2",
		id, companyName,
	).Scan(&job.ID, &job.Title, &job.Department, &job.Location,
		&job.Status, &job.DatePosted, &job.Description, &job.Requirements,
		&job.LastModified, &job.CompanyName)

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

// AddJob creates a new job
func AddJob(w http.ResponseWriter, r *http.Request) {
	var job models.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Set company name from the authenticated user
	job.CompanyName = r.Context().Value("companyName").(string)

	// Record which manager/admin created this job (docs/architecture.md section 13.4)
	job.CreatedByUserID = r.Context().Value("userID").(int)

	// Insert job into database
	var id int
	err = db.DB.QueryRow(
		`INSERT INTO jobs (title, department, location, status, 
			description, requirements, company_name, created_by_user_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id`,
		job.Title, job.Department, job.Location, job.Status,
		job.Description, job.Requirements, job.CompanyName, job.CreatedByUserID,
	).Scan(&id)

	if err != nil {
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

// UpdateJob updates an existing job
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

	// Ensure company name matches authenticated user's company
	job.CompanyName = r.Context().Value("companyName").(string)
	job.ID = id

	// Update job in database
	_, err = db.DB.Exec(
		`UPDATE jobs 
		SET title = $1, department = $2, location = $3, status = $4, 
			description = $5, requirements = $6, last_modified = NOW() 
		WHERE id = $7 AND company_name = $8`,
		job.Title, job.Department, job.Location, job.Status,
		job.Description, job.Requirements, job.ID, job.CompanyName,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating job")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Job updated successfully",
		Data:    job,
	})
}

// DeleteJob deletes a job
func DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Get company name from context
	companyName := r.Context().Value("companyName").(string)

	// Delete job from database
	result, err := db.DB.Exec(
		"DELETE FROM jobs WHERE id = $1 AND company_name = $2",
		id, companyName,
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
