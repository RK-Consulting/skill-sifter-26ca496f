package submission

import "time"

type RecipientType string

const (
	RecipientClient        RecipientType = "client"
	RecipientHiringManager RecipientType = "hiring_manager"
)

func (t RecipientType) Valid() bool {
	return t == RecipientClient || t == RecipientHiringManager
}

type Submission struct {
	ID                  int
	TenantID            string
	AssignmentID        int
	SubmittedByUserID   int
	RecipientType       RecipientType
	RecipientClientID   *int
	RecipientUserID     *int
	RecipientName       string
	RecipientEmail      string
	SubmissionContext   string
	RecruiterNotes      string
	CandidateSnapshot   []byte
	RequirementSnapshot []byte
	SubmittedAt         time.Time
	CreatedAt           time.Time
}

type CreateInput struct {
	AssignmentID      int
	SubmittedByUserID int
	RecipientType     RecipientType
	RecipientClientID *int
	RecipientUserID   *int
	RecipientName     string
	RecipientEmail    string
	SubmissionContext string
	RecruiterNotes    string
}
