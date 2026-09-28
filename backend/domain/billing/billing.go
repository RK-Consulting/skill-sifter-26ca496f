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

type WorklistItem struct {
	CandidateID      int        `json:"candidateId"`
	CandidateName    string     `json:"candidateName"`
	RequirementID    int        `json:"requirementId"`
	RequirementJobID string     `json:"requirementJobId,omitempty"`
	RequirementTitle string     `json:"requirementTitle"`
	ClientID         int        `json:"clientId"`
	ClientName       string     `json:"clientName"`
	JoiningID        int        `json:"joiningId"`
	JoiningDate      time.Time  `json:"joiningDate"`
	BillingID        *int       `json:"billingId,omitempty"`
	BillingDate      *time.Time `json:"billingDate,omitempty"`
	Amount           *string    `json:"amount,omitempty"`
	Currency         *string    `json:"currency,omitempty"`
	InvoiceReference *string    `json:"invoiceReference,omitempty"`
	Billed           bool       `json:"billed"`
}

type CreateInput struct {
	CandidateID      int
	RequirementID    int
	Amount           string
	Currency         string
	InvoiceReference string
}
