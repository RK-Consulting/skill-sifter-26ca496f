package joining

import "time"

const (
	StatusScheduled = "scheduled"
	StatusJoined    = "joined"
	StatusNoShow    = "no_show"
	StatusCancelled = "cancelled"
)

type Joining struct {
	ID                   int
	TenantID             string
	CandidateID          int
	RequirementID        int
	OfferID              int
	Status               string
	ExpectedJoiningDate  *time.Time
	ActualJoiningDate    *time.Time
	Notes                string
	CreatedAt            time.Time
	LastModified         time.Time
}

type CreateInput struct {
	CandidateID         int
	RequirementID       int
	ExpectedJoiningDate *time.Time
	Notes               string
}

type UpdateInput struct {
	Status              string
	ExpectedJoiningDate *time.Time
	ActualJoiningDate   *time.Time
	Notes               string
}
