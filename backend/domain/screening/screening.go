package screening

import "time"

// Screening is a recruiter interaction record attached to one Recruitment
// Assignment. It is evidence gathered during recruitment and is separate
// from the Candidate master record.
type Screening struct {
	ID                  int
	TenantID            string
	AssignmentID        int
	RecruiterUserID     int
	CurrentCTC          string
	ExpectedCTC         string
	NoticePeriod        string
	LastWorkingDay      *time.Time
	CurrentLocation     string
	WillingToRelocate   *bool
	PreferredLocation   string
	ReasonForChange     string
	OffersInHand        string
	CandidateInterest   string
	AvailabilityDate    *time.Time
	RelevantExperience  string
	RecruiterAssessment string
	Notes               string
	ScreenedAt          time.Time
	CreatedAt           time.Time
}

type CreateInput struct {
	AssignmentID        int
	RecruiterUserID     int
	CurrentCTC          string
	ExpectedCTC         string
	NoticePeriod        string
	LastWorkingDay      *time.Time
	CurrentLocation     string
	WillingToRelocate   *bool
	PreferredLocation   string
	ReasonForChange     string
	OffersInHand        string
	CandidateInterest   string
	AvailabilityDate    *time.Time
	RelevantExperience  string
	RecruiterAssessment string
	Notes               string
}
