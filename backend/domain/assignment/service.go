package assignment

import (
	"database/sql"
	"errors"
	"fmt"
)

// Errors returned by Service, distinct from Repository's errors, so
// callers (eventually HTTP handlers) can map each to the right response
// without string-matching.
var (
	ErrCandidateNotFound    = errors.New("candidate not found")
	ErrCandidateNotEligible = errors.New("candidate is not eligible for a new assignment")
	ErrRequirementNotFound  = errors.New("requirement not found")
	ErrUserNotFound         = errors.New("user not found")
)

// eligibleCandidateStatus is the only candidate status a new assignment
// may be created against, per ADR 0004 section 2 ("A candidate in
// inactive, blacklisted, or archived status is not eligible for a new
// recruitment submission by default").
const eligibleCandidateStatus = "active"

// Service enforces the business rules around Assignment creation and
// retrieval: tenant consistency across candidate/requirement/owner/creator,
// candidate eligibility, and the no-duplicate-assignment rule (delegated to
// the repository's database constraint). It is the only place these rules
// live — Repository knows nothing about them, and future handlers should
// not reimplement them.
type Service struct {
	repo Repository
	db   *sql.DB // for tenant-membership lookups against candidates/requirements/users, which are outside recruitment_assignments and therefore outside Repository's responsibility
}

// NewService constructs a Service. dbConn is used only for the
// cross-entity tenant/eligibility checks below (candidates, requirements,
// users) — all recruitment_assignments persistence goes through repo.
func NewService(repo Repository, dbConn *sql.DB) *Service {
	return &Service{repo: repo, db: dbConn}
}

// CreateInput is the validated set of inputs needed to create an
// assignment. It deliberately does not include TenantID or
// CreatedByUserID — those come from authenticated context, passed
// separately to CreateAssignment, never from caller-supplied input (same
// convention as #33/#34: tenant and actor identity are never accepted as
// request data).
type CreateInput struct {
	CandidateID   int
	RequirementID int
	// OwnerUserID may be zero, in which case the assignment's owner
	// defaults to the actor (the user creating the assignment).
	OwnerUserID int
}

// CreateAssignment validates tenant consistency and candidate eligibility,
// then creates a new draft-status assignment. tenantID and actorUserID
// must come from authenticated request context.
func (s *Service) CreateAssignment(tenantID string, actorUserID int, input CreateInput) (*Assignment, error) {
	ownerUserID := input.OwnerUserID
	if ownerUserID == 0 {
		ownerUserID = actorUserID
	}

	candidateStatus, err := s.candidateStatus(input.CandidateID, tenantID)
	if err != nil {
		return nil, err
	}
	if candidateStatus != eligibleCandidateStatus {
		return nil, fmt.Errorf("%w: candidate status is %q", ErrCandidateNotEligible, candidateStatus)
	}

	if err := s.requireRequirementInTenant(input.RequirementID, tenantID); err != nil {
		return nil, err
	}

	// Both the owner and the creator must belong to the assignment's
	// tenant (ADR 0003 section 3: "Both users must belong to the
	// assignment tenant"). actorUserID is re-checked defensively even
	// though it comes from an authenticated JWT, matching the same
	// paranoid-verification pattern used for client ownership in #34.
	if err := s.requireUserInTenant(ownerUserID, tenantID); err != nil {
		return nil, err
	}
	if err := s.requireUserInTenant(actorUserID, tenantID); err != nil {
		return nil, err
	}

	a := &Assignment{
		TenantID:        tenantID,
		CandidateID:     input.CandidateID,
		RequirementID:   input.RequirementID,
		Status:          StatusDraft,
		CreatedByUserID: actorUserID,
		OwnerUserID:     ownerUserID,
	}

	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return a, nil
}

// GetAssignment retrieves an assignment scoped to tenantID. Returns
// ErrNotFound if it doesn't exist in that tenant (including if it exists
// in a different tenant — see Repository.GetByID's documentation).
func (s *Service) GetAssignment(tenantID string, id int) (*Assignment, error) {
	return s.repo.GetByID(tenantID, id)
}

// ListAssignments returns every assignment belonging to tenantID.
func (s *Service) ListAssignments(tenantID string) ([]*Assignment, error) {
	return s.repo.ListByTenant(tenantID)
}

func (s *Service) candidateStatus(candidateID int, tenantID string) (string, error) {
	var status string
	err := s.db.QueryRow(`SELECT status FROM candidates WHERE id = $1 AND tenant_id = $2`, candidateID, tenantID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrCandidateNotFound
		}
		return "", err
	}
	return status, nil
}

func (s *Service) requireRequirementInTenant(requirementID int, tenantID string) error {
	var exists bool
	if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM requirements WHERE id = $1 AND tenant_id = $2)`, requirementID, tenantID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrRequirementNotFound
	}
	return nil
}

func (s *Service) requireUserInTenant(userID int, tenantID string) error {
	var exists bool
	if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND tenant_id = $2)`, userID, tenantID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}
	return nil
}
