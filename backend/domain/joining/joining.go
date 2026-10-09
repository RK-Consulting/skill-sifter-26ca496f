package joining

import "time"

type Joining struct {
	ID            int        `json:"id"`
	TenantID      string     `json:"tenantId"`
	CandidateID   int        `json:"candidateId"`
	RequirementID int        `json:"requirementId"`
	OfferID       int        `json:"offerId"`
	JoiningDate   *time.Time `json:"joiningDate,omitempty"`
	Joined        bool       `json:"joined"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastModified  time.Time  `json:"lastModified"`
}

type CreateInput struct {
	CandidateID   int
	RequirementID int
	JoiningDate   *time.Time
	Joined        bool
	ActorUserID   int
}

type UpdateInput struct {
	JoiningDate *time.Time
	Joined      bool
	ActorUserID int
}
