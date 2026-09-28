package billing

import "time"

type Billing struct {
	ID               int
	TenantID         string
	CandidateID      int
	RequirementID    int
	ClientID         int
	JoiningID        int
	BillingDate      time.Time
	Amount           string
	Currency         string
	InvoiceReference string
	CreatedAt        time.Time
	LastModified     time.Time
}

type CreateInput struct {
	CandidateID      int
	RequirementID    int
	Amount           string
	Currency         string
	InvoiceReference string
}
