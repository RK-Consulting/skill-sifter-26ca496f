package submission

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrCandidateNotFound = errors.New("candidate or requirement not found")
	ErrClientNotFound    = errors.New("client not found")
	ErrRecipientNotFound = errors.New("recipient not found")
	ErrInvalidRecipient  = errors.New("invalid submission recipient")
	ErrAlreadySubmitted  = errors.New("candidate has already been submitted for this requirement")
)

type Service struct {
	repo Repository
	db   *sql.DB
}

func NewService(repo Repository, dbConn *sql.DB) *Service { return &Service{repo: repo, db: dbConn} }

func (s *Service) Submit(tenantID string, input CreateInput) (*Submission, error) {
	if input.CandidateID == 0 || input.RequirementID == 0 {
		return nil, fmt.Errorf("candidateId and requirementId are required")
	}
	if input.SubmittedByUserID == 0 {
		return nil, fmt.Errorf("submittedByUserId is required")
	}
	if !input.RecipientType.Valid() {
		return nil, ErrInvalidRecipient
	}

	var candidateName, jobID, requirementTitle string
	err := s.db.QueryRow(`SELECT c.name,r.job_id,r.title FROM candidates c JOIN requirements r ON r.tenant_id=c.tenant_id
		WHERE c.id=$1 AND r.id=$2 AND c.tenant_id=$3 AND r.tenant_id=$3`, input.CandidateID, input.RequirementID, tenantID).
		Scan(&candidateName, &jobID, &requirementTitle)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCandidateNotFound
		}
		return nil, err
	}

	var recruiterExists bool
	if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND tenant_id=$2)`, input.SubmittedByUserID, tenantID).Scan(&recruiterExists); err != nil {
		return nil, err
	}
	if !recruiterExists {
		return nil, ErrRecipientNotFound
	}

	var screened bool
	if err := s.db.QueryRow(`SELECT EXISTS(
		SELECT 1 FROM recruitment_screenings
		WHERE candidate_id=$1 AND requirement_id=$2 AND tenant_id=$3
		  AND status = 'completed'
	)`, input.CandidateID, input.RequirementID, tenantID).Scan(&screened); err != nil {
		return nil, err
	}
	if !screened {
		return nil, fmt.Errorf("a completed screening is required before submission")
	}

	var existing bool
	if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM recruitment_submissions WHERE candidate_id=$1 AND requirement_id=$2 AND tenant_id=$3)`,
		input.CandidateID, input.RequirementID, tenantID).Scan(&existing); err != nil {
		return nil, err
	}
	if existing {
		return nil, ErrAlreadySubmitted
	}

	switch input.RecipientType {
	case RecipientClient:
		if input.RecipientClientID == nil {
			return nil, ErrInvalidRecipient
		}
		var exists bool
		if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM clients WHERE id=$1 AND tenant_id=$2)`, *input.RecipientClientID, tenantID).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrClientNotFound
		}
	case RecipientHiringManager:
		if input.RecipientUserID == nil {
			return nil, ErrInvalidRecipient
		}
		var exists bool
		if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND tenant_id=$2)`, *input.RecipientUserID, tenantID).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrRecipientNotFound
		}
	}

	var candidateSnapshot, requirementSnapshot []byte
	if err := s.db.QueryRow(`SELECT row_to_json(c)::jsonb FROM candidates c WHERE c.id=$1 AND c.tenant_id=$2`, input.CandidateID, tenantID).Scan(&candidateSnapshot); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT row_to_json(r)::jsonb FROM requirements r WHERE r.id=$1 AND r.tenant_id=$2`, input.RequirementID, tenantID).Scan(&requirementSnapshot); err != nil {
		return nil, err
	}

	record := &Submission{TenantID: tenantID, CandidateID: input.CandidateID, RequirementID: input.RequirementID,
		SubmittedByUserID: input.SubmittedByUserID, RecipientType: input.RecipientType, RecipientClientID: input.RecipientClientID,
		RecipientUserID: input.RecipientUserID, RecipientName: input.RecipientName, RecipientEmail: input.RecipientEmail,
		SubmissionContext: input.SubmissionContext, RecruiterNotes: input.RecruiterNotes,
		CandidateSnapshot: candidateSnapshot, RequirementSnapshot: requirementSnapshot}
	if err := s.repo.Create(record); err != nil {
		return nil, err
	}
	_ = candidateName
	_ = jobID
	_ = requirementTitle
	return record, nil
}

func (s *Service) GetSubmission(tenantID string, id int) (*Submission, error) {
	return s.repo.GetByID(tenantID, id)
}
func (s *Service) ListByCandidateRequirement(tenantID string, candidateID, requirementID int) ([]*Submission, error) {
	var valid bool
	if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM candidates c JOIN requirements r ON r.tenant_id=c.tenant_id WHERE c.id=$1 AND r.id=$2 AND c.tenant_id=$3 AND r.tenant_id=$3)`, candidateID, requirementID, tenantID).Scan(&valid); err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrCandidateNotFound
	}
	return s.repo.ListByCandidateRequirement(tenantID, candidateID, requirementID)
}
