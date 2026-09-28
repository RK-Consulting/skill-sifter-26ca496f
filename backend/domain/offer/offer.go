package offer

import "time"

const (
StatusDraft = "draft"
StatusOffered = "offered"
StatusAccepted = "accepted"
StatusDeclined = "declined"
StatusExpired = "expired"
StatusWithdrawn = "withdrawn"
)

type Offer struct {
ID int
TenantID string
CandidateID int
RequirementID int
SelectionID int
OfferReference string
Status string
OfferedDate time.Time
ExpectedJoiningDate *time.Time
Compensation string
Terms string
Notes string
DecisionAt *time.Time
CreatedAt time.Time
LastModified time.Time
}

type CreateInput struct {
CandidateID int
RequirementID int
OfferReference string
ExpectedJoiningDate *time.Time
Compensation string
Terms string
Notes string
}

type UpdateInput struct {
Status string
ExpectedJoiningDate *time.Time
Compensation string
Terms string
Notes string
}
