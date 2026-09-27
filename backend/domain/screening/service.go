package screening

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrAssignmentNotFound = errors.New("recruitment assignment not found")
	ErrRecruiterNotFound  = errors.New("recruiter not found")
)

type Service struct {
	repo Repository
	db   *sql.DB
}

func NewService(repo Repository, dbConn *sql.DB) *Service {
	return &Service{repo: repo, db: dbConn}
}

func (s *Service) CreateScreening(tenantID string, input CreateInput) (*Screening, error) {
	if input.AssignmentID == 0 {
		return nil, fmt.Errorf("assignmentId is required")
	}
	if input.RecruiterUserID == 0 {
		return nil, fmt.Errorf("recruiter user is required")
	}

	var assignmentStatus string
	err := s.db.QueryRow(
		`SELECT status FROM recruitment_assignments WHERE id = $1 AND tenant_id = $2`,
		input.AssignmentID, tenantID,
	).Scan(&assignmentStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAssignmentNotFound
		}
		return nil, err
	}
	if assignmentStatus == "joined" || assignmentStatus == "rejected" || assignmentStatus == "withdrawn" {
		return nil, fmt.Errorf("cannot record screening for terminal assignment")
	}

	var recruiterExists bool
	if err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND tenant_id = $2)`,
		input.RecruiterUserID, tenantID,
	).Scan(&recruiterExists); err != nil {
		return nil, err
	}
	if !recruiterExists {
		return nil, ErrRecruiterNotFound
	}

	record := &Screening{
		TenantID:            tenantID,
		AssignmentID:        input.AssignmentID,
		RecruiterUserID:     input.RecruiterUserID,
		CurrentCTC:          input.CurrentCTC,
		ExpectedCTC:         input.ExpectedCTC,
		NoticePeriod:        input.NoticePeriod,
		LastWorkingDay:      input.LastWorkingDay,
		CurrentLocation:     input.CurrentLocation,
		WillingToRelocate:   input.WillingToRelocate,
		PreferredLocation:   input.PreferredLocation,
		ReasonForChange:     input.ReasonForChange,
		OffersInHand:        input.OffersInHand,
		CandidateInterest:   input.CandidateInterest,
		AvailabilityDate:    input.AvailabilityDate,
		RelevantExperience:  input.RelevantExperience,
		RecruiterAssessment: input.RecruiterAssessment,
		Notes:               input.Notes,
	}
	if err := s.repo.Create(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *Service) GetScreening(tenantID string, id int) (*Screening, error) {
	return s.repo.GetByID(tenantID, id)
}

func (s *Service) ListByAssignment(tenantID string, assignmentID int) ([]*Screening, error) {
	var exists bool
	if err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM recruitment_assignments WHERE id = $1 AND tenant_id = $2)`,
		assignmentID, tenantID,
	).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAssignmentNotFound
	}
	return s.repo.ListByAssignment(tenantID, assignmentID)
}
