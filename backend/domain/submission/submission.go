package submission

import "time"

type RecipientType string

const (
	RecipientClient RecipientType = "client"
)

func (t RecipientType) Valid() bool { return t == RecipientClient }

type Submission struct {
	ID                  int
	TenantID            string
	CandidateID         int
	RequirementID       int
	SubmittedByUserID   int
	RecipientType       RecipientType
	RecipientClientID   *int
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
	CandidateID       int
	RequirementID     int
	SubmittedByUserID int
	RecipientType     RecipientType
	RecipientClientID *int
	RecipientUserID   *int
	RecipientName     string
	RecipientEmail    string
	SubmissionContext string
	RecruiterNotes    string
}
