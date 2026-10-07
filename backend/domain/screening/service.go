package screening

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/RK-Consulting/skill-sifter/domain/audit"
)

var (
	ErrCandidateNotFound   = errors.New("candidate not found")
	ErrRequirementNotFound = errors.New("requirement not found")
	ErrRecruiterNotFound   = errors.New("recruiter not found")
)

type Service struct {
	repo Repository
	db   *sql.DB
}

func NewService(repo Repository, dbConn *sql.DB) *Service { return &Service{repo: repo, db: dbConn} }

func (s *Service) CreateScreening(tenantID string, input CreateInput) (*Screening, error) {
	if input.CandidateID == 0 || input.RequirementID == 0 {
		return nil, fmt.Errorf("candidateId and requirementId are required")
	}
	if input.RecruiterUserID == 0 {
		return nil, fmt.Errorf("recruiter user is required")
	}

	var valid bool
	if err := s.db.QueryRow(`SELECT EXISTS(
		SELECT 1 FROM candidates c JOIN requirements r ON r.tenant_id = c.tenant_id
		WHERE c.id=$1 AND r.id=$2 AND c.tenant_id=$3 AND r.tenant_id=$3)`,
		input.CandidateID, input.RequirementID, tenantID).Scan(&valid); err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrCandidateNotFound
	}

	var recruiterExists bool
	if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND tenant_id=$2)`,
		input.RecruiterUserID, tenantID).Scan(&recruiterExists); err != nil {
		return nil, err
	}
	if !recruiterExists {
		return nil, ErrRecruiterNotFound
	}

	record := &Screening{
		TenantID: tenantID, CandidateID: input.CandidateID, RequirementID: input.RequirementID,
		RecruiterUserID: input.RecruiterUserID, CurrentCTC: input.CurrentCTC,
		ExpectedCTC: input.ExpectedCTC, NoticePeriod: input.NoticePeriod,
		LastWorkingDay: input.LastWorkingDay, CurrentLocation: input.CurrentLocation,
		WillingToRelocate: input.WillingToRelocate, PreferredLocation: input.PreferredLocation,
		ReasonForChange: input.ReasonForChange, OffersInHand: input.OffersInHand,
		CandidateInterest: input.CandidateInterest, AvailabilityDate: input.AvailabilityDate,
		RelevantExperience: input.RelevantExperience, RecruiterAssessment: input.RecruiterAssessment,
		Notes: input.Notes,
	}
	if err := s.repo.Create(record); err != nil {
		return nil, err
	}
	if err := audit.Write(s.db, tenantID, input.RecruiterUserID, "screening", record.ID, "created", map[string]interface{}{"candidateId": input.CandidateID, "requirementId": input.RequirementID}); err != nil {
		return nil, fmt.Errorf("write screening audit event: %w", err)
	}
	return record, nil
}

func (s *Service) GetScreening(tenantID string, id int) (*Screening, error) {
	return s.repo.GetByID(tenantID, id)
}

func (s *Service) ListByCandidateRequirement(tenantID string, candidateID, requirementID int) ([]*Screening, error) {
	var valid bool
	if err := s.db.QueryRow(`SELECT EXISTS(
		SELECT 1 FROM candidates c JOIN requirements r ON r.tenant_id=c.tenant_id
		WHERE c.id=$1 AND r.id=$2 AND c.tenant_id=$3 AND r.tenant_id=$3)`,
		candidateID, requirementID, tenantID).Scan(&valid); err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrCandidateNotFound
	}
	return s.repo.ListByCandidateRequirement(tenantID, candidateID, requirementID)
}
