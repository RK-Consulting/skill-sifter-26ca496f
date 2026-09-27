package submission

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/RK-Consulting/skill-sifter/domain/assignment"
)

var (
	ErrAssignmentNotFound = errors.New("recruitment assignment not found")
	ErrClientNotFound     = errors.New("client not found")
	ErrRecipientNotFound  = errors.New("hiring manager not found")
	ErrInvalidRecipient   = errors.New("invalid submission recipient")
	ErrAlreadySubmitted   = errors.New("assignment is already submitted")
)

type Service struct {
	repo Repository
	db   *sql.DB
}

func NewService(repo Repository, dbConn *sql.DB) *Service {
	return &Service{repo: repo, db: dbConn}
}

func (s *Service) Submit(tenantID string, input CreateInput) (*Submission, error) {
	if input.AssignmentID == 0 { return nil, fmt.Errorf("assignmentId is required") }
	if input.SubmittedByUserID == 0 { return nil, fmt.Errorf("submittedByUserId is required") }
	if !input.RecipientType.Valid() { return nil, ErrInvalidRecipient }

	var status string
	var candidateID, requirementID int
	err := s.db.QueryRow(`
		SELECT status, candidate_id, requirement_id
		FROM recruitment_assignments WHERE id = $1 AND tenant_id = $2`,
		input.AssignmentID, tenantID,
	).Scan(&status, &candidateID, &requirementID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, ErrAssignmentNotFound }
		return nil, err
	}
	if status == string(assignment.StatusSubmitted) || status == string(assignment.StatusInterviewing) ||
		status == string(assignment.StatusOffered) || status == string(assignment.StatusJoined) {
		return nil, ErrAlreadySubmitted
	}
	if status == string(assignment.StatusRejected) || status == string(assignment.StatusWithdrawn) {
		return nil, fmt.Errorf("cannot submit a terminal assignment")
	}

	var actorExists bool
	if err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND tenant_id = $2)`,
		input.SubmittedByUserID, tenantID,
	).Scan(&actorExists); err != nil { return nil, err }
	if !actorExists { return nil, ErrRecipientNotFound }

	switch input.RecipientType {
	case RecipientClient:
		if input.RecipientClientID == nil { return nil, ErrInvalidRecipient }
		var exists bool
		if err := s.db.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1 AND tenant_id = $2)`,
			*input.RecipientClientID, tenantID,
		).Scan(&exists); err != nil { return nil, err }
		if !exists { return nil, ErrClientNotFound }
	case RecipientHiringManager:
		if input.RecipientUserID == nil { return nil, ErrInvalidRecipient }
		var exists bool
		if err := s.db.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND tenant_id = $2)`,
			*input.RecipientUserID, tenantID,
		).Scan(&exists); err != nil { return nil, err }
		if !exists { return nil, ErrRecipientNotFound }
	}

	var candidateSnapshot, requirementSnapshot []byte
	if err := s.db.QueryRow(
		`SELECT to_jsonb(c) FROM candidates c WHERE c.id = $1 AND c.tenant_id = $2`,
		candidateID, tenantID,
	).Scan(&candidateSnapshot); err != nil { return nil, err }
	if err := s.db.QueryRow(
		`SELECT to_jsonb(r) FROM requirements r WHERE r.id = $1 AND r.tenant_id = $2`,
		requirementID, tenantID,
	).Scan(&requirementSnapshot); err != nil { return nil, err }

	record := &Submission{
		TenantID: tenantID, AssignmentID: input.AssignmentID,
		SubmittedByUserID: input.SubmittedByUserID, RecipientType: input.RecipientType,
		RecipientClientID: input.RecipientClientID, RecipientUserID: input.RecipientUserID,
		RecipientName: input.RecipientName, RecipientEmail: input.RecipientEmail,
		SubmissionContext: input.SubmissionContext, RecruiterNotes: input.RecruiterNotes,
		CandidateSnapshot: candidateSnapshot, RequirementSnapshot: requirementSnapshot,
	}
	if err := s.repo.Create(record); err != nil { return nil, err }

	// Use the existing Assignment domain transition so lifecycle validation
	// and the assignment's own submission snapshot/audit behavior remain
	// centralized.
	_, err = assignment.NewService(assignment.NewPostgresRepository(s.db), s.db).
		TransitionAssignment(tenantID, input.SubmittedByUserID, input.AssignmentID, assignment.StatusSubmitted)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s *Service) GetSubmission(tenantID string, id int) (*Submission, error) {
	return s.repo.GetByID(tenantID, id)
}

func (s *Service) ListByAssignment(tenantID string, assignmentID int) ([]*Submission, error) {
	var exists bool
	if err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM recruitment_assignments WHERE id = $1 AND tenant_id = $2)`,
		assignmentID, tenantID,
	).Scan(&exists); err != nil { return nil, err }
	if !exists { return nil, ErrAssignmentNotFound }
	return s.repo.ListByAssignment(tenantID, assignmentID)
}
