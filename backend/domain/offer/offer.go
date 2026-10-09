package offer

import "time"

type Offer struct {
	ID            int       `json:"id"`
	TenantID      string    `json:"tenantId"`
	CandidateID   int       `json:"candidateId"`
	RequirementID int       `json:"requirementId"`
	SelectionID   int       `json:"selectionId"`
	Accepted      bool      `json:"accepted"`
	CreatedAt     time.Time `json:"createdAt"`
	LastModified  time.Time `json:"lastModified"`
}

type CreateInput struct {
	CandidateID   int
	RequirementID int
}

type UpdateInput struct {
	Accepted bool
}
