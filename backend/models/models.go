package models

import (
	"time"
)

// Candidate model — fields match the actual deployed candidates table
// (backend/database/schema.sql / migrations/001_baseline.sql) and what
// frontend/src/pages/AddCandidate.tsx actually sends. The previous version
// of this struct referenced status/source/date_applied/resume_url/cover_letter/
// last_modified — none of which exist in the real schema, meaning
// GetCandidates/AddCandidate/UpdateCandidate were broken (every call would
// fail with "column does not exist"). Fixed as part of the Dashboard
// mock-data investigation, docs/architecture.md gap audit.
type Candidate struct {
	ID             int       `json:"id" db:"id,primarykey,autoincrement"`
	Name           string    `json:"name" db:"name,notnull"`
	Email          string    `json:"email" db:"email,notnull"`
	Phone          string    `json:"phone" db:"phone"`
	Position       string    `json:"position" db:"position"`
	Location       string    `json:"location" db:"location"`
	Experience     string    `json:"experience" db:"experience"`
	CurrentCTC     string    `json:"currentCTC" db:"currentctc"`
	ExpectedCTC    string    `json:"expectedCTC" db:"expectedctc"`
	NoticePeriod   string    `json:"noticePeriod" db:"noticeperiod"`
	JLPTLanguage   string    `json:"jlptLanguage" db:"jlptlanguage"`
	Skills         string    `json:"skills" db:"skills"`
	JobDescription string    `json:"newJD" db:"jobdescription"`
	CreatedAt      time.Time `json:"createdAt,omitempty" db:"created_at,default:CURRENT_TIMESTAMP"`
	TenantID       string    `json:"tenantId" db:"tenant_id,notnull,foreignkey:companies(id)"`
	CompanyName    string    `json:"companyName" db:"company_name,notnull"`
}

// Job model
type Job struct {
	ID              int       `json:"id" db:"id,primarykey,autoincrement"`
	Title           string    `json:"title" db:"title,notnull"`
	Department      string    `json:"department" db:"department"`
	Location        string    `json:"location" db:"location"`
	Status          string    `json:"status" db:"status,default:'open'"`
	DatePosted      time.Time `json:"datePosted" db:"date_posted,default:CURRENT_TIMESTAMP"`
	Description     string    `json:"description,omitempty" db:"description"`
	Requirements    string    `json:"requirements,omitempty" db:"requirements"`
	LastModified    time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	TenantID        string    `json:"tenantId" db:"tenant_id,notnull,foreignkey:companies(id)"`
	CompanyName     string    `json:"companyName" db:"company_name,notnull"`
	CreatedByUserID int       `json:"createdByUserId,omitempty" db:"created_by_user_id"`
}

// DailyJob model
type DailyJob struct {
	ID               int       `json:"id" db:"id,primarykey,autoincrement"`
	JdNo             int       `json:"jdNo" db:"jd_no,notnull"`
	Instructions     string    `json:"instructions" db:"instructions"`
	AssignedUser     int       `json:"assignedUser" db:"assigned_user"`
	AssignedUsername string    `json:"assignedUsername,omitempty"` // Not stored in DB, used for display
	AssignedDate     time.Time `json:"assignedDate" db:"assigned_date,default:CURRENT_TIMESTAMP"`
	LastModified     time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	TenantID         string    `json:"tenantId" db:"tenant_id,notnull,foreignkey:companies(id)"`
	CompanyName      string    `json:"companyName" db:"company_name,notnull"`
}

// Interview model
type Interview struct {
	ID            int       `json:"id" db:"id,primarykey,autoincrement"`
	CandidateID   int       `json:"candidateId" db:"candidate_id"`
	CandidateName string    `json:"candidateName" db:"candidate_name,notnull"`
	Position      string    `json:"position" db:"position"`
	InterviewDate time.Time `json:"interviewDate" db:"interview_date,notnull"`
	Status        string    `json:"status" db:"status,default:'scheduled'"`
	Feedback      string    `json:"feedback" db:"feedback"`
	LastModified  time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	TenantID      string    `json:"tenantId" db:"tenant_id,notnull,foreignkey:companies(id)"`
	CompanyName   string    `json:"companyName" db:"company_name,notnull"`
}

// BusinessDev model
type BusinessDev struct {
	ID            int       `json:"id" db:"id,primarykey,autoincrement"`
	ClientName    string    `json:"clientName" db:"client_name,notnull"`
	PartnerName   string    `json:"partnerName" db:"partner_name"`
	ContactPerson string    `json:"contactPerson" db:"contact_person,notnull"`
	ContactNumber string    `json:"contactNumber" db:"contact_number"`
	ContactEmail  string    `json:"contactEmail" db:"contact_email,notnull"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
	LastModified  time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	TenantID      string    `json:"tenantId" db:"tenant_id,notnull,foreignkey:companies(id)"`
	CompanyName   string    `json:"companyName" db:"company_name,notnull"`
}

// Company model
type Company struct {
	ID        string    `json:"id" db:"id,primarykey"`
	Name      string    `json:"name" db:"name,notnull,unique"`
	CreatedAt time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
}

// Role model
type Role struct {
	ID          int       `json:"id" db:"id,primarykey,autoincrement"`
	Name        string    `json:"name" db:"name,notnull,unique"`
	Permissions []string  `json:"permissions" db:"permissions,type:jsonb,default:'[]'::jsonb"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
}

// User model
type User struct {
	ID          int       `json:"id" db:"id,primarykey,autoincrement"`
	Username    string    `json:"username" db:"username,notnull"`
	Email       string    `json:"email" db:"email,notnull,unique"`
	Password    string    `json:"password,omitempty" db:"password,notnull"`
	Role        string    `json:"role" db:"role,notnull"`
	TenantID    string    `json:"tenantId" db:"tenant_id,notnull,foreignkey:companies(id)"`
	CompanyName string    `json:"companyName" db:"company_name,notnull"` // Changed from CompanyID
	CreatedAt   time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
}

// SchemaVersion for tracking db schema version
type SchemaVersion struct {
	Version   int       `json:"version" db:"version,primarykey"`
	CreatedAt time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
}

// Credentials for login/register
type Credentials struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Username    string `json:"username,omitempty"`
	CompanyName string `json:"companyName,omitempty"` // Changed from CompanyID
	Role        string `json:"role,omitempty"`
}

// ApiResponse represents a standard API response
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// TokenResponse for authentication
type TokenResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Reports
type HiringReportEntry struct {
	Date            string `json:"date"`
	TotalInterviews int    `json:"totalInterviews"`
	TotalCandidates int    `json:"totalCandidates"`
	TotalHires      int    `json:"totalHires"`
}

type SourceReportEntry struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

// HiringReportResponse represents aggregated hiring stats for the reports endpoint
// @Description Aggregated report of interviews by status
// @Success 200 {object} HiringReportResponse
// @Router /api/reports/hiring [get]
type HiringReportResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    []HiringReportEntry `json:"data"`
}

// SourceReportResponse represents candidate source distribution for reporting
// @Description Aggregated report of candidates by source
// @Success 200 {object} SourceReportResponse
// @Router /api/reports/sources [get]
type SourceReportResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    []SourceReportEntry `json:"data"`
}

// ActivityEntry represents a single real event for the Dashboard's Recent
// Activity feed. Built from real timestamps across candidates, jobs,
// business_dev, daily_jobs, and interviews — replaces the hardcoded mock
// data that previously lived in Dashboard.tsx.
type ActivityEntry struct {
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}
