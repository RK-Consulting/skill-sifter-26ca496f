package screening

import "time"

type Screening struct {
	ID                  int
	TenantID            string
	CandidateID         int
	RequirementID       int
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