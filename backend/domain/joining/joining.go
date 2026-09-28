package joining

import "time"

type Joining struct {
	ID int
	TenantID string
	CandidateID int
	RequirementID int
	OfferID int
	JoiningDate *time.Time
	Joined bool
	CreatedAt time.Time
	LastModified time.Time
}

type CreateInput struct {
	CandidateID int
	RequirementID int
	JoiningDate *time.Time
	Joined bool
}

type UpdateInput struct {
	JoiningDate *time.Time
	Joined bool
}
