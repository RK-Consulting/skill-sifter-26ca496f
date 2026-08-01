package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
)

// GetRecentActivity returns a real, timestamp-sorted feed of recent events
// across candidates, jobs, business_dev, daily_jobs, and interviews.
// Replaces the hardcoded mock data previously in Dashboard.tsx's
// activityData array. Not sourced from a dedicated activity-log table (none
// exists) — built as a UNION ALL over each table's own real timestamp
// column, which is sufficient for a simple recent-events feed without
// needing new schema.
func GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	companyName := r.Context().Value("companyName").(string)

	query := `
		(SELECT 'candidate' AS type, 'New candidate application' AS title,
			name || ' added as a candidate' AS description, created_at AS ts
		 FROM candidates WHERE company_name = $1)
		UNION ALL
		(SELECT 'job' AS type, 'New job posted' AS title,
			title || ' posted' AS description, date_posted AS ts
		 FROM jobs WHERE company_name = $1)
		UNION ALL
		(SELECT 'business_dev' AS type, 'New business contact added' AS title,
			client_name || ' added as a new client' AS description, created_at AS ts
		 FROM business_dev WHERE company_name = $1)
		UNION ALL
		(SELECT 'daily_job' AS type, 'Daily task assigned' AS title,
			instructions AS description, assigned_date AS ts
		 FROM daily_jobs WHERE company_name = $1)
		UNION ALL
		(SELECT 'interview' AS type, 'Interview scheduled' AS title,
			candidate_name || ' — ' || position AS description, last_modified AS ts
		 FROM interviews WHERE company_name = $1)
		ORDER BY ts DESC
		LIMIT 10`

	rows, err := db.DB.Query(query, companyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching recent activity")
		return
	}
	defer rows.Close()

	activity := []models.ActivityEntry{}
	for rows.Next() {
		var a models.ActivityEntry
		if err := rows.Scan(&a.Type, &a.Title, &a.Description, &a.Timestamp); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning activity row")
			return
		}
		activity = append(activity, a)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Recent activity retrieved successfully",
		Data:    activity,
	})
}
func GetHiringReport(w http.ResponseWriter, r *http.Request) {
	companyName := r.Context().Value("companyName").(string)

	query := `
		SELECT DATE_TRUNC('month', interview_date) as month, COUNT(*) as total
		FROM interviews
		WHERE company_name = $1
		GROUP BY month
		ORDER BY month
	`

	rows, err := db.DB.Query(query, companyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch hiring report")
		return
	}
	defer rows.Close()

	report := []models.HiringReportEntry{}
	for rows.Next() {
		var entry models.HiringReportEntry
		var month time.Time
		if err := rows.Scan(&month, &entry.TotalInterviews); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning hiring report")
			return
		}
		entry.Date = month.Format("2006-01")
		report = append(report, entry)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Hiring report fetched",
		Data:    report,
	})
}

// GetSourceReport returns candidate source distribution
func GetSourceReport(w http.ResponseWriter, r *http.Request) {
	companyName := r.Context().Value("companyName").(string)

	query := `
		SELECT COALESCE(source, 'unknown') AS source, COUNT(*)
		FROM candidates
		WHERE company_name = $1
		GROUP BY source
	`

	rows, err := db.DB.Query(query, companyName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch source report")
		return
	}
	defer rows.Close()

	report := []models.SourceReportEntry{}
	for rows.Next() {
		var entry models.SourceReportEntry
		if err := rows.Scan(&entry.Source, &entry.Count); err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			respondWithError(w, http.StatusInternalServerError, "Error scanning source report")
			return
		}
		report = append(report, entry)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Source report fetched",
		Data:    report,
	})
}
