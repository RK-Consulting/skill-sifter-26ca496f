package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// GetCompanyUsers retrieves all users for a company for dropdown selection
func GetCompanyUsers(w http.ResponseWriter, r *http.Request) {
	// Get company name from context
	companyName := r.Context().Value("companyName").(string)

	users := []models.User{}
	rows, err := db.DB.Query("SELECT id, username FROM users WHERE company_name = $1", companyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching company users")
		return
	}
	defer rows.Close()

	// Scan rows into users slice
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

// GetDailyJobs retrieves all daily jobs for a company
func GetDailyJobs(w http.ResponseWriter, r *http.Request) {
	// Get company name from context
	companyName := r.Context().Value("companyName").(string)

	// Modify the query to JOIN with users table to get usernames
	rows, err := db.DB.Query(`
		SELECT dj.id, dj.jd_no, dj.instructions, dj.assigned_user, 
			u.username as assigned_username, dj.assigned_date, dj.last_modified, dj.company_name
		FROM daily_jobs dj
		LEFT JOIN users u ON dj.assigned_user = u.id
		WHERE dj.company_name = $1
	`, companyName)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching daily jobs")
		return
	}
	defer rows.Close()

	dailyJobs := []models.DailyJob{}
	// Scan rows into dailyJobs slice with username
	for rows.Next() {
		var dj models.DailyJob
		var username *string // Use pointer to handle NULL values

		err := rows.Scan(&dj.ID, &dj.JdNo, &dj.Instructions, &dj.AssignedUser,
			&username, &dj.AssignedDate, &dj.LastModified, &dj.CompanyName)

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning daily job row")
			return
		}

		// Set the username if available
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

// GetDailyJobByID retrieves a single daily job by ID
func GetDailyJobByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid daily job ID")
		return
	}

	// Get company name from context
	companyName := r.Context().Value("companyName").(string)

	var dailyJob models.DailyJob
	var username *string // Use pointer to handle NULL values

	err = db.DB.QueryRow(`
		SELECT dj.id, dj.jd_no, dj.instructions, dj.assigned_user, 
			u.username as assigned_username, dj.assigned_date, dj.last_modified, dj.company_name
		FROM daily_jobs dj
		LEFT JOIN users u ON dj.assigned_user = u.id
		WHERE dj.id = $1 AND dj.company_name = $2
	`, id, companyName).Scan(&dailyJob.ID, &dailyJob.JdNo, &dailyJob.Instructions,
		&dailyJob.AssignedUser, &username, &dailyJob.AssignedDate,
		&dailyJob.LastModified, &dailyJob.CompanyName)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Daily job not found")
		return
	}

	// Set the username if available
	if username != nil {
		dailyJob.AssignedUsername = *username
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Daily job retrieved successfully",
		Data:    dailyJob,
	})
}

// AddDailyJob creates a new daily job
func AddDailyJob(w http.ResponseWriter, r *http.Request) {
	var dailyJob models.DailyJob
	err := json.NewDecoder(r.Body).Decode(&dailyJob)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Set company name from the authenticated user
	dailyJob.CompanyName = r.Context().Value("companyName").(string)

	// Insert daily job into database
	var id int
	err = db.DB.QueryRow(
		`INSERT INTO daily_jobs (jd_no, instructions, assigned_user, company_name) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`,
		dailyJob.JdNo, dailyJob.Instructions, dailyJob.AssignedUser, dailyJob.CompanyName,
	).Scan(&id)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating daily job")
		return
	}

	dailyJob.ID = id

	// Get the assigned username
	var username string
	err = db.DB.QueryRow(
		"SELECT username FROM users WHERE id = $1 AND company_name = $2",
		dailyJob.AssignedUser, dailyJob.CompanyName,
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

// UpdateDailyJob updates an existing daily job
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

	// Ensure company name matches authenticated user's company
	dailyJob.CompanyName = r.Context().Value("companyName").(string)
	dailyJob.ID = id

	// Update daily job in database
	_, err = db.DB.Exec(
		`UPDATE daily_jobs 
		SET jd_no = $1, instructions = $2, assigned_user = $3, last_modified = NOW() 
		WHERE id = $4 AND company_name = $5`,
		dailyJob.JdNo, dailyJob.Instructions, dailyJob.AssignedUser,
		dailyJob.ID, dailyJob.CompanyName,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating daily job")
		return
	}

	// Get the assigned username
	var username string
	err = db.DB.QueryRow(
		"SELECT username FROM users WHERE id = $1 AND company_name = $2",
		dailyJob.AssignedUser, dailyJob.CompanyName,
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

// DeleteDailyJob deletes a daily job
func DeleteDailyJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid daily job ID")
		return
	}

	// Get company name from context
	companyName := r.Context().Value("companyName").(string)

	// Delete daily job from database
	result, err := db.DB.Exec(
		"DELETE FROM daily_jobs WHERE id = $1 AND company_name = $2",
		id, companyName,
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
