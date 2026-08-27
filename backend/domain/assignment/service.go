package assignment

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
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

// DeleteAssignment removes the assignment with the given id, scoped to
// tenantID. Returns ErrNotFound if it doesn't exist in that tenant.
func (s *Service) DeleteAssignment(tenantID string, id int) error {
	return s.repo.Delete(tenantID, id)
}

// ChangeOwner reassigns an existing assignment's owner. This is
// deliberately the only mutation exposed via UpdateAssignment/PUT in
// checkpoint 3: lifecycle status transitions go through
// TransitionAssignment instead (checkpoint 4), keeping owner mutation and
// lifecycle transition as two distinct concepts rather than letting
// arbitrary status values enter the PUT payload.
func (s *Service) ChangeOwner(tenantID string, id int, newOwnerUserID int) (*Assignment, error) {
	a, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	if err := s.requireUserInTenant(newOwnerUserID, tenantID); err != nil {
		return nil, err
	}

	a.OwnerUserID = newOwnerUserID
	if err := s.repo.Update(a); err != nil {
		return nil, err
	}
	return a, nil
}

// TransitionAssignment moves an existing assignment to newStatus,
// enforcing ADR 0003's transition rules via Assignment.TransitionTo, and
// persists the result. It is the ONLY way an assignment's status changes
// via the service layer — no other Service method touches Status. Returns
// ErrNotFound if the assignment doesn't exist in tenantID (including if it
// exists in a different tenant), or a *TransitionError if newStatus is not
// a legal transition from the assignment's current status.
//
// At formal submission (transitioning into StatusSubmitted), this captures
// immutable candidate/requirement snapshots per ADR 0003 section 6. The
// fetch-current-assignment, capture-snapshot, and persist-new-status steps
// all happen inside a single database transaction (SELECT ... FOR UPDATE,
// then UPDATE), so the status change and the snapshot write are atomic: if
// anything fails partway, the whole attempt rolls back and the assignment
// is left exactly as it was, never partially transitioned or
// partially snapshotted. If a snapshot already exists (SnapshotCreatedAt
// is already set), it is never recaptured or overwritten — this should
// not be reachable in ordinary use anyway, since screening -> submitted is
// the only transition into StatusSubmitted in the lifecycle, but the guard
// is explicit rather than assumed.
func (s *Service) TransitionAssignment(tenantID string, id int, newStatus Status) (*Assignment, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // no-op once Commit succeeds; rolls back on any early return

	row := tx.QueryRow(`SELECT `+assignmentSelectColumns+`
		FROM recruitment_assignments WHERE id = $1 AND tenant_id = $2 FOR UPDATE`,
		id, tenantID,
	)
	a, err := scanAssignmentRow(row)
	if err != nil {
		return nil, err
	}

	if err := a.TransitionTo(newStatus); err != nil {
		return nil, err
	}

	if newStatus == StatusSubmitted && a.SnapshotCreatedAt == nil {
		candidateData, err := fetchCandidateSnapshotTx(tx, a.CandidateID, tenantID)
		if err != nil {
			return nil, err
		}
		requirementData, err := fetchRequirementSnapshotTx(tx, a.RequirementID, tenantID)
		if err != nil {
			return nil, err
		}

		candidateJSON, err := json.Marshal(candidateData)
		if err != nil {
			return nil, err
		}
		requirementJSON, err := json.Marshal(requirementData)
		if err != nil {
			return nil, err
		}

		now := time.Now()
		a.CandidateSnapshot = candidateJSON
		a.RequirementSnapshot = requirementJSON
		a.SnapshotCreatedAt = &now
	}

	var snapshotCreatedAtParam interface{}
	if a.SnapshotCreatedAt != nil {
		snapshotCreatedAtParam = *a.SnapshotCreatedAt
	}
	var candidateSnapshotParam, requirementSnapshotParam interface{}
	if a.CandidateSnapshot != nil {
		candidateSnapshotParam = a.CandidateSnapshot
	}
	if a.RequirementSnapshot != nil {
		requirementSnapshotParam = a.RequirementSnapshot
	}

	result, err := tx.Exec(`
		UPDATE recruitment_assignments
		SET status = $1, owner_user_id = $2, candidate_snapshot = $3, requirement_snapshot = $4,
			snapshot_created_at = $5, last_modified = NOW()
		WHERE id = $6 AND tenant_id = $7`,
		string(a.Status), a.OwnerUserID, candidateSnapshotParam, requirementSnapshotParam,
		snapshotCreatedAtParam, a.ID, a.TenantID,
	)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return a, nil
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
