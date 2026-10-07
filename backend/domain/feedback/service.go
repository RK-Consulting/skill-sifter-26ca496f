package feedback

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/RK-Consulting/skill-sifter/domain/audit"
	"strings"
)

var (
	ErrSubmissionNotFound    = errors.New("recruitment submission not found")
	ErrFeedbackActorNotFound = errors.New("feedback actor not found")
	ErrInvalidOutcome        = errors.New("invalid feedback outcome")
	ErrInvalidReasonCode     = errors.New("a valid reason code is required for hold or reject feedback")
)

type Service struct {
	repo Repository
	db   *sql.DB
}

func NewService(repo Repository, dbConn *sql.DB) *Service {
	return &Service{repo: repo, db: dbConn}
}

func (s *Service) CreateFeedback(tenantID string, input CreateInput) (*Feedback, error) {
	if input.SubmissionID == 0 {
		return nil, fmt.Errorf("submissionId is required")
	}
	if input.FeedbackByUserID == 0 {
		return nil, fmt.Errorf("feedbackByUserId is required")
	}
	if !input.Outcome.Valid() {
		return nil, ErrInvalidOutcome
	}
	if input.ReasonCode != "" && !ReasonCode(strings.TrimSpace(input.ReasonCode)).Valid() {
		return nil, ErrInvalidReasonCode
	}
	if (input.Outcome == OutcomeHold || input.Outcome == OutcomeReject) &&
		!ReasonCode(strings.TrimSpace(input.ReasonCode)).Valid() {
		return nil, ErrInvalidReasonCode
	}

	var submissionExists bool
	if err := s.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM recruitment_submissions
			WHERE id = $1 AND tenant_id = $2
		)`,
		input.SubmissionID, tenantID,
	).Scan(&submissionExists); err != nil {
		return nil, err
	}
	if !submissionExists {
		return nil, ErrSubmissionNotFound
	}

	var actorExists bool
	if err := s.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM users
			WHERE id = $1 AND tenant_id = $2
		)`,
		input.FeedbackByUserID, tenantID,
	).Scan(&actorExists); err != nil {
		return nil, err
	}
	if !actorExists {
		return nil, ErrFeedbackActorNotFound
	}

	f := &Feedback{
		TenantID:         tenantID,
		SubmissionID:     input.SubmissionID,
		FeedbackByUserID: input.FeedbackByUserID,
		Outcome:          input.Outcome,
		ReasonCode:       strings.TrimSpace(input.ReasonCode),
		Comments:         strings.TrimSpace(input.Comments),
		NextAction:       strings.TrimSpace(input.NextAction),
	}

	if err := s.repo.Create(f); err != nil {
		return nil, err
	}
	if err := audit.Write(s.db, tenantID, input.FeedbackByUserID, "feedback", f.ID, "created", map[string]interface{}{"submissionId": input.SubmissionID, "outcome": input.Outcome}); err != nil {
		return nil, fmt.Errorf("write feedback audit event: %w", err)
	}
	return f, nil
}

func (s *Service) GetFeedback(tenantID string, id int) (*Feedback, error) {
	return s.repo.GetByID(tenantID, id)
}

func (s *Service) ListBySubmission(tenantID string, submissionID int) ([]*Feedback, error) {
	var exists bool
	if err := s.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM recruitment_submissions
			WHERE id = $1 AND tenant_id = $2
		)`,
		submissionID, tenantID,
	).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrSubmissionNotFound
	}
	return s.repo.ListBySubmission(tenantID, submissionID)
}
