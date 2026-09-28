package offer

import "time"

type Offer struct {
	ID            int
	TenantID      string
	CandidateID   int
	RequirementID int
	SelectionID   int
	Accepted      bool
	CreatedAt     time.Time
	LastModified  time.Time
}

type CreateInput struct {
	CandidateID   int
	RequirementID int
}

type UpdateInput struct {
	Accepted bool
}
