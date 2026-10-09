package billing

import "time"

type Billing struct {
	ID               int       `json:"id"`
	TenantID         string    `json:"tenantId"`
	CandidateID      int       `json:"candidateId"`
	RequirementID    int       `json:"requirementId"`
	ClientID         int       `json:"clientId"`
	JoiningID        int       `json:"joiningId"`
	BillingDate      time.Time `json:"billingDate"`
	Amount           string    `json:"amount"`
	Currency         string    `json:"currency"`
	InvoiceReference string    `json:"invoiceReference,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	LastModified     time.Time `json:"lastModified"`
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
