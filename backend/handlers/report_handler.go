package handlers

import (
	"fmt"
	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"net/http"
	"strconv"
	"time"
)

type PeriodReportRow struct {
	Period         string `json:"period"`
	Activities     int    `json:"activities"`
	Candidates     int    `json:"candidates"`
	Resumes        int    `json:"resumes"`
	ResumeSearches int    `json:"resumeSearches"`
	Jobs           int    `json:"jobs"`
	Interviews     int    `json:"interviews"`
	Hires          int    `json:"hires"`
	BusinessDev    int    `json:"businessDev"`
}
type ActivityLogRow struct {
	ID          int64     `json:"id"`
	Action      string    `json:"action"`
	EntityType  string    `json:"entityType"`
	EntityID    string    `json:"entityId,omitempty"`
	Description string    `json:"description"`
	ActorUserID *int      `json:"actorUserId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

func GetPeriodicReport(w http.ResponseWriter, r *http.Request) {
	company := r.Context().Value("companyName").(string)
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}
	var trunc, since, label string
	switch period {
	case "daily":
		trunc, since, label = "day", "30 days", "2006-01-02"
	case "monthly":
		trunc, since, label = "month", "12 months", "2006-01"
	case "quarterly":
		trunc, since, label = "quarter", "8 quarters", "2006-01"
	case "yearly":
		trunc, since, label = "year", "5 years", "2006"
	default:
		respondWithError(w, 400, "period must be daily, monthly, quarterly or yearly")
		return
	}
	query := fmt.Sprintf(`SELECT date_trunc('%s',created_at) period,COUNT(*) activities,COUNT(*) FILTER(WHERE action='CANDIDATES_INSERT') candidates,COUNT(*) FILTER(WHERE action='RESUMES_INSERT') resumes,COUNT(*) FILTER(WHERE action='RESUME_SEARCHED') resume_searches,COUNT(*) FILTER(WHERE action='JOBS_INSERT') jobs,COUNT(*) FILTER(WHERE action='INTERVIEWS_INSERT') interviews,COUNT(*) FILTER(WHERE action='INTERVIEWS_UPDATE' AND metadata->'after'->>'status' IN('hired','selected','offer_accepted')) hires,COUNT(*) FILTER(WHERE action='BUSINESS_DEV_INSERT') business_dev FROM activity_logs WHERE company_name=$1 AND created_at>=NOW()-INTERVAL '%s' GROUP BY period ORDER BY period DESC`, trunc, since)
	rows, err := db.DB.Query(query, company)
	if err != nil {
		respondWithError(w, 500, "Failed to build report")
		return
	}
	defer rows.Close()
	out := []PeriodReportRow{}
	for rows.Next() {
		var p time.Time
		var x PeriodReportRow
		if err := rows.Scan(&p, &x.Activities, &x.Candidates, &x.Resumes, &x.ResumeSearches, &x.Jobs, &x.Interviews, &x.Hires, &x.BusinessDev); err != nil {
			respondWithError(w, 500, "Failed to read report")
			return
		}
		x.Period = p.Format(label)
		if period == "quarterly" {
			q := (int(p.Month())-1)/3 + 1
			x.Period = fmt.Sprintf("Q%d %d", q, p.Year())
		}
		out = append(out, x)
	}
	respondWithJSON(w, 200, models.ApiResponse{Success: true, Message: "Periodic report fetched", Data: out})
}

func GetActivityLog(w http.ResponseWriter, r *http.Request) {
	company := r.Context().Value("companyName").(string)
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, e := strconv.Atoi(raw); e == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	action := r.URL.Query().Get("action")
	query := `SELECT id,action,entity_type,COALESCE(entity_id,''),COALESCE(description,''),actor_user_id,created_at FROM activity_logs WHERE company_name=$1`
	args := []interface{}{company}
	if action != "" {
		query += ` AND action=$2`
		args = append(args, action)
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT %d`, limit)
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch activity log")
		return
	}
	defer rows.Close()
	out := []ActivityLogRow{}
	for rows.Next() {
		var x ActivityLogRow
		if err := rows.Scan(&x.ID, &x.Action, &x.EntityType, &x.EntityID, &x.Description, &x.ActorUserID, &x.CreatedAt); err != nil {
			respondWithError(w, 500, "Failed to read activity log")
			return
		}
		out = append(out, x)
	}
	respondWithJSON(w, 200, models.ApiResponse{Success: true, Message: "Activity log fetched", Data: out})
}

func GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	company := r.Context().Value("companyName").(string)
	rows, err := db.DB.Query(`SELECT action,description,created_at FROM activity_logs WHERE company_name=$1 ORDER BY created_at DESC LIMIT 10`, company)
	if err != nil {
		respondWithError(w, 500, "Error fetching recent activity")
		return
	}
	defer rows.Close()
	activity := []models.ActivityEntry{}
	for rows.Next() {
		var a models.ActivityEntry
		if err := rows.Scan(&a.Type, &a.Description, &a.Timestamp); err != nil {
			respondWithError(w, 500, "Error scanning activity row")
			return
		}
		a.Title = a.Type
		activity = append(activity, a)
	}
	respondWithJSON(w, 200, models.ApiResponse{Success: true, Message: "Recent activity retrieved successfully", Data: activity})
}

func GetHiringReport(w http.ResponseWriter, r *http.Request) {
	company := r.Context().Value("companyName").(string)
	rows, err := db.DB.Query(`SELECT TO_CHAR(DATE_TRUNC('month',interview_date),'YYYY-MM'),COUNT(*) FROM interviews WHERE company_name=$1 GROUP BY 1 ORDER BY 1`, company)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch hiring report")
		return
	}
	defer rows.Close()
	report := []models.HiringReportEntry{}
	for rows.Next() {
		var e models.HiringReportEntry
		if err := rows.Scan(&e.Date, &e.TotalInterviews); err != nil {
			respondWithError(w, 500, "Error scanning hiring report")
			return
		}
		report = append(report, e)
	}
	respondWithJSON(w, 200, models.ApiResponse{Success: true, Message: "Hiring report fetched", Data: report})
}

func GetSourceReport(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, 200, models.ApiResponse{Success: true, Message: "Source report is available through candidate data; legacy source field is not part of the current schema", Data: []models.SourceReportEntry{}})
}
