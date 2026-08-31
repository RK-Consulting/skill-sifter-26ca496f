package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetCompanyUsers retrieves all users for the authenticated tenant, for
// dropdown selection. Scoped by tenant_id (ADR 0001), not company_name.
func GetCompanyUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)

	users := []models.User{}
	rows, err := db.DB.Query("SELECT id, username FROM users WHERE tenant_id = $1", tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching company users")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
		}
		err := rows.Scan(&user.ID, &user.Username)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning user row")
			return
		}
		users = append(users, models.User{ID: user.ID, Username: user.Username})
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Company users retrieved successfully",
		Data:    users,
	})
}

// GetDailyJobs retrieves all daily jobs for the authenticated tenant.
// Scoped by tenant_id (ADR 0001), not company_name.
func GetDailyJobs(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)

	rows, err := db.DB.Query(`
		SELECT dj.id, dj.jd_no, dj.instructions, dj.assigned_user, 
			u.username as assigned_username, dj.assigned_date, dj.last_modified, dj.tenant_id, dj.company_name
		FROM daily_jobs dj
		LEFT JOIN users u ON dj.assigned_user = u.id
		WHERE dj.tenant_id = $1
	`, tenantID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching daily jobs")
		return
	}
	defer rows.Close()

	dailyJobs := []models.DailyJob{}
	for rows.Next() {
		var dj models.DailyJob
		var username *string // Use pointer to handle NULL values

		err := rows.Scan(&dj.ID, &dj.JdNo, &dj.Instructions, &dj.AssignedUser,
			&username, &dj.AssignedDate, &dj.LastModified, &dj.TenantID, &dj.CompanyName)

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning daily job row")
			return
		}

		if username != nil {
			dj.AssignedUsername = *username
		}

		dailyJobs = append(dailyJobs, dj)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Daily jobs retrieved successfully",
		Data:    dailyJobs,
	})
}

// GetDailyJobByID retrieves a single daily job by ID, scoped to the
// authenticated tenant. A daily job ID belonging to another tenant returns
// 404, identically to a nonexistent ID (ADR 0001).
func GetDailyJobByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid daily job ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	var dailyJob models.DailyJob
	var username *string

	err = db.DB.QueryRow(`
		SELECT dj.id, dj.jd_no, dj.instructions, dj.assigned_user, 
			u.username as assigned_username, dj.assigned_date, dj.last_modified, dj.tenant_id, dj.company_name
		FROM daily_jobs dj
		LEFT JOIN users u ON dj.assigned_user = u.id
		WHERE dj.id = $1 AND dj.tenant_id = $2
	`, id, tenantID).Scan(&dailyJob.ID, &dailyJob.JdNo, &dailyJob.Instructions,
		&dailyJob.AssignedUser, &username, &dailyJob.AssignedDate,
		&dailyJob.LastModified, &dailyJob.TenantID, &dailyJob.CompanyName)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Daily job not found")
		return
	}

	if username != nil {
		dailyJob.AssignedUsername = *username
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Daily job retrieved successfully",
		Data:    dailyJob,
	})
}

// AddDailyJob creates a new daily job under the authenticated tenant.
// tenant_id is always derived from context, never from the request payload.
func AddDailyJob(w http.ResponseWriter, r *http.Request) {
	var dailyJob models.DailyJob
	err := json.NewDecoder(r.Body).Decode(&dailyJob)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	dailyJob.TenantID = r.Context().Value("tenantID").(string)
	dailyJob.CompanyName = r.Context().Value("companyName").(string)

	var id int
	err = db.DB.QueryRow(
		`INSERT INTO daily_jobs (jd_no, instructions, assigned_user, tenant_id, company_name) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`,
		dailyJob.JdNo, dailyJob.Instructions, dailyJob.AssignedUser, dailyJob.TenantID, dailyJob.CompanyName,
	).Scan(&id)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating daily job")
		return
	}

	dailyJob.ID = id

	// Get the assigned username, scoped to the same tenant so a
	// cross-tenant assigned_user id can never leak another tenant's
	// username into this response.
	var username string
	err = db.DB.QueryRow(
		"SELECT username FROM users WHERE id = $1 AND tenant_id = $2",
		dailyJob.AssignedUser, dailyJob.TenantID,
	).Scan(&username)

	if err == nil {
		dailyJob.AssignedUsername = username
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Daily job created successfully",
		Data:    dailyJob,
	})
}

// UpdateDailyJob updates an existing daily job, scoped to the authenticated
// tenant. A daily job ID belonging to another tenant affects zero rows.
func UpdateDailyJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid daily job ID")
		return
	}

	var dailyJob models.DailyJob
	err = json.NewDecoder(r.Body).Decode(&dailyJob)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	tenantID := r.Context().Value("tenantID").(string)
	dailyJob.TenantID = tenantID
	dailyJob.ID = id

	result, err := db.DB.Exec(
		`UPDATE daily_jobs 
		SET jd_no = $1, instructions = $2, assigned_user = $3, last_modified = NOW() 
		WHERE id = $4 AND tenant_id = $5`,
		dailyJob.JdNo, dailyJob.Instructions, dailyJob.AssignedUser,
		dailyJob.ID, tenantID,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating daily job")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Daily job not found")
		return
	}

	var username string
	err = db.DB.QueryRow(
		"SELECT username FROM users WHERE id = $1 AND tenant_id = $2",
		dailyJob.AssignedUser, tenantID,
	).Scan(&username)

	if err == nil {
		dailyJob.AssignedUsername = username
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Daily job updated successfully",
		Data:    dailyJob,
	})
}

// DeleteDailyJob deletes a daily job, scoped to the authenticated tenant.
func DeleteDailyJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid daily job ID")
		return
	}

	tenantID := r.Context().Value("tenantID").(string)

	result, err := db.DB.Exec(
		"DELETE FROM daily_jobs WHERE id = $1 AND tenant_id = $2",
		id, tenantID,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting daily job")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Daily job not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Daily job deleted successfully",
	})
}
