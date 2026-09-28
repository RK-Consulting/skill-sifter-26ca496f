package joining

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrCandidateRequirementNotFound = errors.New("candidate or requirement not found")
	ErrOfferNotFound               = errors.New("accepted offer not found")
	ErrJoiningExists               = errors.New("joining record already exists for candidate and requirement")
	ErrInvalidStatus               = errors.New("invalid joining status")
	ErrInvalidTransition           = errors.New("invalid joining status transition")
	ErrActualDateRequired          = errors.New("actual joining date is required when marking joined")
	ErrActualDateNotAllowed        = errors.New("actual joining date is only valid when joined")
)

type Service struct {
	repo Repository
	db   *sql.DB
}

func NewService(repo Repository, dbConn *sql.DB) *Service {
	return &Service{repo: repo, db: dbConn}
}

func (s *Service) Create(tenantID string, input CreateInput) (*Joining, error) {
	if input.CandidateID == 0 || input.RequirementID == 0 {
		return nil, fmt.Errorf("candidateId and requirementId are required")
	}

	var exists bool
	if err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM candidates c
			JOIN requirements r ON r.tenant_id=c.tenant_id
			WHERE c.id=$1 AND r.id=$2
			  AND c.tenant_id=$3 AND r.tenant_id=$3
		)`, input.CandidateID, input.RequirementID, tenantID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCandidateRequirementNotFound
	}

	var offerID int
	if err := s.db.QueryRow(`
		SELECT id FROM recruitment_offers
		WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3
		  AND status='accepted'
	`, tenantID, input.CandidateID, input.RequirementID).Scan(&offerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOfferNotFound
		}
		return nil, err
	}

	if _, err := s.repo.GetByPair(tenantID, input.CandidateID, input.RequirementID); err == nil {
		return nil, ErrJoiningExists
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	j := &Joining{
		TenantID: tenantID,
		CandidateID: input.CandidateID,
		RequirementID: input.RequirementID,
		OfferID: offerID,
		Status: StatusPending,
		ExpectedJoiningDate: input.ExpectedJoiningDate,
		Notes: strings.TrimSpace(input.Notes),
	}
	if err := s.repo.Create(j); err != nil {
		return nil, err
	}
	return j, nil
}

func (s *Service) Get(tenantID string, candidateID, requirementID int) (*Joining, error) {
	return s.repo.GetByPair(tenantID, candidateID, requirementID)
}

func (s *Service) Update(tenantID string, candidateID, requirementID int, input UpdateInput) (*Joining, error) {
	j, err := s.repo.GetByPair(tenantID, candidateID, requirementID)
	if err != nil {
		return nil, err
	}
	if !validStatus(input.Status) {
		return nil, ErrInvalidStatus
	}
	if !validTransition(j.Status, input.Status) {
		return nil, ErrInvalidTransition
	}
	if input.Status == StatusJoined && input.ActualJoiningDate == nil {
		return nil, ErrActualDateRequired
	}
	if input.Status != StatusJoined && input.ActualJoiningDate != nil {
		return nil, ErrActualDateNotAllowed
	}

	j.Status = input.Status
	j.ExpectedJoiningDate = input.ExpectedJoiningDate
	j.ActualJoiningDate = input.ActualJoiningDate
	j.Notes = strings.TrimSpace(input.Notes)
	if err := s.repo.Update(j); err != nil {
		return nil, err
	}
	return j, nil
}

func validStatus(v string) bool {
	switch v {
	case StatusPending, StatusJoined, StatusNotJoined, StatusWithdrawn:
		return true
	default:
		return false
	}
}

func validTransition(from, to string) bool {
	if from == to {
		return false
	}
	switch from {
	case StatusPending:
		return to == StatusJoined || to == StatusNotJoined || to == StatusWithdrawn
	case StatusJoined:
		return false
	default:
		return false
	}
}

var _ = time.Time{}
