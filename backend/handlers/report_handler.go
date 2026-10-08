package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
)

type PeriodReportRow struct {
	Period         string `json:"period"`
	Activities     int    `json:"activities"`
	Candidates     int    `json:"candidates"`
	Resumes        int    `json:"resumes"`
	ResumeSearches int    `json:"resumeSearches"`
	Requirements   int    `json:"requirements"`
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
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}
	var trunc, since, label string
	switch period {
	case "daily":
		trunc, since, label = "day", "30 days", "2006-01-02"
	case "weekly":
		trunc, since, label = "week", "26 weeks", "2006-01-02"
	case "monthly":
		trunc, since, label = "month", "12 months", "2006-01"
	case "quarterly":
		trunc, since, label = "quarter", "8 quarters", "2006-01"
	case "yearly":
		trunc, since, label = "year", "5 years", "2006"
	default:
		respondWithError(w, http.StatusBadRequest, "period must be daily, weekly, monthly, quarterly or yearly")
		return
	}

	query := fmt.Sprintf(`
		WITH activity AS (
			SELECT created_at, 'candidate' AS kind FROM candidates WHERE tenant_id=$1
			UNION ALL SELECT uploaded_at, 'resume' FROM resumes WHERE tenant_id=$1
			UNION ALL SELECT created_at, 'resume_search' FROM resume_search_logs WHERE tenant_id=$1
			UNION ALL SELECT created_at, 'requirement' FROM requirements WHERE tenant_id=$1
			UNION ALL SELECT interview_date, 'interview' FROM interviews WHERE tenant_id=$1
			UNION ALL SELECT decided_at, 'selection' FROM recruitment_selections WHERE tenant_id=$1
			UNION ALL SELECT created_at, 'offer' FROM recruitment_offers WHERE tenant_id=$1
			UNION ALL SELECT created_at, 'joining' FROM recruitment_joinings WHERE tenant_id=$1
			UNION ALL SELECT created_at, 'billing' FROM recruitment_billings WHERE tenant_id=$1
		)
		SELECT date_trunc('%s', created_at) period,
		       COUNT(*) activities,
		       COUNT(*) FILTER (WHERE kind='candidate') candidates,
		       COUNT(*) FILTER (WHERE kind='resume') resumes,
		       COUNT(*) FILTER (WHERE kind='resume_search') resume_searches,
		       COUNT(*) FILTER (WHERE kind='requirement') requirements,
		       COUNT(*) FILTER (WHERE kind='interview') interviews,
		       COUNT(*) FILTER (WHERE kind IN ('joining','billing')) hires,
		       0::bigint business_dev
		FROM activity
		WHERE created_at >= NOW() - INTERVAL '%s'
		GROUP BY period
		ORDER BY period DESC
	`, trunc, since)

	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}
	rows, err := db.RequestDB(r).Query(query, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to build report")
		return
	}
	defer rows.Close()

	var out []PeriodReportRow
	for rows.Next() {
		var p time.Time
		var x PeriodReportRow
		if err := rows.Scan(&p, &x.Activities, &x.Candidates, &x.Resumes, &x.ResumeSearches, &x.Requirements, &x.Interviews, &x.Hires, &x.BusinessDev); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to read report")
			return
		}
		x.Period = p.Format(label)
		if period == "quarterly" {
			x.Period = fmt.Sprintf("Q%d %d", (int(p.Month())-1)/3+1, p.Year())
		}
		out = append(out, x)
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Periodic report fetched", Data: out})
}

func GetActivityLog(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, e := strconv.Atoi(raw); e == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	action := r.URL.Query().Get("action")
	query := "SELECT id,action,entity_type,COALESCE(entity_id::text,''),COALESCE(metadata->>'description',''),actor_user_id,occurred_at FROM audit_events WHERE tenant_id=$1"
	args := []interface{}{r.Context().Value("tenantID")}
	if action != "" {
		query += " AND action=$2"
		args = append(args, action)
	}
	query += fmt.Sprintf(" ORDER BY occurred_at DESC LIMIT %d", limit)
	rows, err := db.RequestDB(r).Query(query, args...)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch activity log")
		return
	}
	defer rows.Close()
	out := []ActivityLogRow{}
	for rows.Next() {
		var x ActivityLogRow
		if err := rows.Scan(&x.ID, &x.Action, &x.EntityType, &x.EntityID, &x.Description, &x.ActorUserID, &x.CreatedAt); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to read activity log")
			return
		}
		out = append(out, x)
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Activity log fetched", Data: out})
}

func GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	rows, err := db.RequestDB(r).Query(`SELECT action,COALESCE(metadata->>'description',action),occurred_at FROM audit_events WHERE tenant_id=$1 ORDER BY occurred_at DESC LIMIT 10`, r.Context().Value("tenantID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching recent activity")
		return
	}
	defer rows.Close()
	activity := []models.ActivityEntry{}
	for rows.Next() {
		var a models.ActivityEntry
		if err := rows.Scan(&a.Type, &a.Description, &a.Timestamp); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning activity row")
			return
		}
		a.Title = a.Type
		activity = append(activity, a)
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Recent activity retrieved successfully", Data: activity})
}

func GetPipelineReport(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	rows, err := db.RequestDB(r).Query(`
		SELECT stage, COUNT(c.id)
		FROM (
			SELECT unnest(ARRAY['screening', 'interview', 'rejected']) AS stage
		) stages
		LEFT JOIN candidates c
			ON c.pipeline_stage = stages.stage AND c.tenant_id = $1
		GROUP BY stage
		ORDER BY CASE stage
			WHEN 'screening' THEN 1
			WHEN 'interview' THEN 2
			WHEN 'rejected' THEN 3
		END`, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch pipeline report")
		return
	}
	defer rows.Close()

	out := []models.PipelineReportEntry{}
	for rows.Next() {
		var entry models.PipelineReportEntry
		if err := rows.Scan(&entry.Stage, &entry.Count); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to read pipeline report")
			return
		}
		out = append(out, entry)
	}
	if err := rows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read pipeline report")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Pipeline report fetched",
		Data:    out,
	})
}

func GetHiringReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)
	rows, err := db.RequestDB(r).Query(`SELECT TO_CHAR(DATE_TRUNC('month',interview_date),'YYYY-MM'),COUNT(*) FROM interviews WHERE tenant_id=$1 GROUP BY 1 ORDER BY 1`, tenantID)
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
