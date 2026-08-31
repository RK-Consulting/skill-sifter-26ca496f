package assignment

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// Errors returned by Service, distinct from Repository's errors, so
// callers (eventually HTTP handlers) can map each to the right response
// without string-matching.
var (
	ErrCandidateNotFound       = errors.New("candidate not found")
	ErrCandidateNotEligible    = errors.New("candidate is not eligible for a new assignment")
	ErrCandidateAlreadyEngaged = errors.New("candidate already has an active recruitment engagement")
	ErrRequirementNotFound     = errors.New("requirement not found")
	ErrUserNotFound            = errors.New("user not found")
)

const eligibleCandidateStatus = "active"

type Service struct {
	repo Repository
	db   *sql.DB
}

func NewService(repo Repository, dbConn *sql.DB) *Service {
	return &Service{repo: repo, db: dbConn}
}

type CreateInput struct {
	CandidateID   int
	RequirementID int
	OwnerUserID   int
}

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

	correlationID, err := newCorrelationID()
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.QueryRow(`
		INSERT INTO recruitment_assignments (tenant_id, candidate_id, requirement_id, status, created_by_user_id, owner_user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, last_modified`,
		a.TenantID, a.CandidateID, a.RequirementID, string(a.Status), a.CreatedByUserID, a.OwnerUserID,
	).Scan(&a.ID, &a.CreatedAt, &a.LastModified)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch {
			case pqErr.Code == "23505":
				return nil, ErrDuplicateAssignment
			case pqErr.Code == "23514" && pqErr.Constraint == "candidate_active_recruitment_engagement":
				return nil, ErrCandidateAlreadyEngaged
			}
		}
		return nil, err
	}

	if err := recordAuditEventTx(tx, tenantID, actorUserID, a.ID, AuditAssignmentCreated, correlationID, map[string]int{
		"candidateId":   a.CandidateID,
		"requirementId": a.RequirementID,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Service) GetAssignment(tenantID string, id int) (*Assignment, error) {
	return s.repo.GetByID(tenantID, id)
}

func (s *Service) ListAssignments(tenantID string) ([]*Assignment, error) {
	return s.repo.ListByTenant(tenantID)
}

func (s *Service) DeleteAssignment(tenantID string, id int) error {
	return s.repo.Delete(tenantID, id)
}

func (s *Service) ChangeOwner(tenantID string, actorUserID int, id int, newOwnerUserID int) (*Assignment, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRow(`SELECT `+assignmentSelectColumns+`
		FROM recruitment_assignments WHERE id = $1 AND tenant_id = $2 FOR UPDATE`,
		id, tenantID,
	)
	a, err := scanAssignmentRow(row)
	if err != nil {
		return nil, err
	}

	var newOwnerExists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND tenant_id = $2)`, newOwnerUserID, tenantID).Scan(&newOwnerExists); err != nil {
		return nil, err
	}
	if !newOwnerExists {
		return nil, ErrUserNotFound
	}

	previousOwnerUserID := a.OwnerUserID
	a.OwnerUserID = newOwnerUserID

	result, err := tx.Exec(`UPDATE recruitment_assignments SET owner_user_id = $1, last_modified = NOW() WHERE id = $2 AND tenant_id = $3`,
		a.OwnerUserID, a.ID, a.TenantID)
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

	correlationID, err := newCorrelationID()
	if err != nil {
		return nil, err
	}
	if err := recordAuditEventTx(tx, tenantID, actorUserID, a.ID, AuditAssignmentOwnerChanged, correlationID, map[string]int{
		"previousOwnerUserId": previousOwnerUserID,
		"newOwnerUserId":      newOwnerUserID,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Service) TransitionAssignment(tenantID string, actorUserID int, id int, newStatus Status) (*Assignment, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRow(`SELECT `+assignmentSelectColumns+`
		FROM recruitment_assignments WHERE id = $1 AND tenant_id = $2 FOR UPDATE`,
		id, tenantID,
	)
	a, err := scanAssignmentRow(row)
	if err != nil {
		return nil, err
	}

	fromStatus := a.Status
	if err := a.TransitionTo(newStatus); err != nil {
		return nil, err
	}

	snapshotJustCaptured := false
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
		snapshotJustCaptured = true
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

	correlationID, err := newCorrelationID()
	if err != nil {
		return nil, err
	}
	if action, ok := auditActionForStatus[newStatus]; ok {
		if err := recordAuditEventTx(tx, tenantID, actorUserID, a.ID, action, correlationID, map[string]string{
			"from": string(fromStatus),
			"to":   string(newStatus),
		}); err != nil {
			return nil, err
		}
	}
	if snapshotJustCaptured {
		if err := recordAuditEventTx(tx, tenantID, actorUserID, a.ID, AuditAssignmentSnapshotCreated, correlationID, map[string]string{}); err != nil {
			return nil, err
		}
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
