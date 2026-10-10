package submission

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/RK-Consulting/skill-sifter/domain/audit"
	"github.com/lib/pq"
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

	var candidateName, requirementTitle string
	err := s.db.QueryRow(`SELECT c.name,r.title FROM candidates c JOIN requirements r ON r.tenant_id=c.tenant_id
		WHERE c.id=$1 AND r.id=$2 AND c.tenant_id=$3 AND r.tenant_id=$3`, input.CandidateID, input.RequirementID, tenantID).
		Scan(&candidateName, &requirementTitle)
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
		RecipientName: input.RecipientName, RecipientEmail: input.RecipientEmail,
		SubmissionContext: input.SubmissionContext, RecruiterNotes: input.RecruiterNotes,
		CandidateSnapshot: candidateSnapshot, RequirementSnapshot: requirementSnapshot}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var completedScreeningID int
	if err := tx.QueryRow(`
		SELECT id
		FROM recruitment_screenings
		WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3 AND status='completed'
		ORDER BY screened_at DESC, id DESC
		LIMIT 1
		FOR UPDATE
	`, tenantID, input.CandidateID, input.RequirementID).Scan(&completedScreeningID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("a completed screening is required before submission")
		}
		return nil, err
	}

	if err := s.repo.CreateTx(tx, record); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrAlreadySubmitted
		}
		return nil, err
	}
	if err := audit.WriteTx(tx, tenantID, input.SubmittedByUserID, "submission", record.ID, "created", map[string]interface{}{"candidateId": input.CandidateID, "requirementId": input.RequirementID}); err != nil {
		return nil, fmt.Errorf("write submission audit event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	_ = candidateName
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
