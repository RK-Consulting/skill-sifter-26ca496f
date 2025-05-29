// (Implemented GetHiringReport and GetSourceReport functions)
package handlers

import (
	//"database/sql"
	"net/http"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
)

// GetHiringReport returns aggregated hiring data
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
		SELECT source, COUNT(*)
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
