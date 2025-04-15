package models

import (
	"time"
)

// Candidate model
type Candidate struct {
	ID           int       `json:"id" db:"id,primarykey,autoincrement"`
	Name         string    `json:"name" db:"name,notnull"`
	Email        string    `json:"email" db:"email,notnull"`
	Phone        string    `json:"phone" db:"phone"`
	Position     string    `json:"position" db:"position"`
	Status       string    `json:"status" db:"status,default:'applied'"`
	DateApplied  time.Time `json:"dateApplied" db:"date_applied,default:CURRENT_TIMESTAMP"`
	ResumeURL    string    `json:"resumeUrl,omitempty" db:"resume_url"`
	CoverLetter  string    `json:"coverLetter,omitempty" db:"cover_letter"`
	LastModified time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	CompanyID    int       `json:"companyId" db:"company_id,notnull"`
}

// Job model
type Job struct {
	ID           int       `json:"id" db:"id,primarykey,autoincrement"`
	Title        string    `json:"title" db:"title,notnull"`
	Department   string    `json:"department" db:"department"`
	Location     string    `json:"location" db:"location"`
	Status       string    `json:"status" db:"status,default:'open'"`
	DatePosted   time.Time `json:"datePosted" db:"date_posted,default:CURRENT_TIMESTAMP"`
	Description  string    `json:"description,omitempty" db:"description"`
	Requirements string    `json:"requirements,omitempty" db:"requirements"`
	LastModified time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	CompanyID    int       `json:"companyId" db:"company_id,notnull"`
}

// DailyJob model
type DailyJob struct {
	ID           int       `json:"id" db:"id,primarykey,autoincrement"`
	JdNo         int       `json:"jdNo" db:"jd_no,notnull"`
	Instructions string    `json:"instructions" db:"instructions"`
	AssignedUser int       `json:"assignedUser" db:"assigned_user"`
	AssignedDate time.Time `json:"assignedDate" db:"assigned_date,default:CURRENT_TIMESTAMP"`
	LastModified time.Time `json:"lastModified" db:"last_modified,default:CURRENT_TIMESTAMP"`
	CompanyID    int       `json:"companyId" db:"company_id,notnull"`
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
	CompanyID     int       `json:"companyId" db:"company_id,notnull"`
}

// Company model
type Company struct {
	ID        int       `json:"id" db:"id,primarykey,autoincrement"`
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
	ID        int       `json:"id" db:"id,primarykey,autoincrement"`
	Username  string    `json:"username" db:"username,notnull"`
	Email     string    `json:"email" db:"email,notnull,unique"`
	Password  string    `json:"password,omitempty" db:"password,notnull"`
	Role      string    `json:"role" db:"role,notnull"`
	CompanyID int       `json:"companyId" db:"company_id,notnull"`
	CreatedAt time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
}

// SchemaVersion for tracking db schema version
type SchemaVersion struct {
	Version   int       `json:"version" db:"version,primarykey"`
	CreatedAt time.Time `json:"createdAt" db:"created_at,default:CURRENT_TIMESTAMP"`
}

// Credentials for login/register
type Credentials struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Username  string `json:"username,omitempty"`
	CompanyID int    `json:"companyId,omitempty"`
	Company   string `json:"company,omitempty"`
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
