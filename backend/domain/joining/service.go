package joining

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/RK-Consulting/skill-sifter/domain/audit"
	"github.com/lib/pq"
)

var (
	ErrCandidateRequirementNotFound = errors.New("candidate or requirement not found")
	ErrOfferNotFound                = errors.New("accepted offer not found")
	ErrJoiningExists                = errors.New("joining record already exists for candidate and requirement")
	ErrJoiningDateRequired          = errors.New("joining date is required when joined is true")
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
	if input.Joined && input.JoiningDate == nil {
		return nil, ErrJoiningDateRequired
	}

	var exists bool
	if err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM candidates c
			JOIN requirements r ON r.tenant_id = c.tenant_id
			WHERE c.id = $1
			  AND r.id = $2
			  AND c.tenant_id = $3
			  AND r.tenant_id = $3
		)
	`, input.CandidateID, input.RequirementID, tenantID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCandidateRequirementNotFound
	}

	if _, err := s.repo.GetByPair(tenantID, input.CandidateID, input.RequirementID); err == nil {
		return nil, ErrJoiningExists
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	j := &Joining{
		TenantID:      tenantID,
		CandidateID:   input.CandidateID,
		RequirementID: input.RequirementID,
		JoiningDate:   input.JoiningDate,
		Joined:        input.Joined,
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var accepted bool
	if err := tx.QueryRow(`
		SELECT id, accepted
		FROM recruitment_offers
		WHERE tenant_id = $1 AND candidate_id = $2 AND requirement_id = $3
		FOR UPDATE
	`, tenantID, input.CandidateID, input.RequirementID).Scan(&j.OfferID, &accepted); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOfferNotFound
		}
		return nil, err
	}
	if !accepted {
		return nil, ErrOfferNotFound
	}

	if err := s.repo.CreateTx(tx, j); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrJoiningExists
		}
		return nil, err
	}
	if err := audit.WriteTx(tx, tenantID, input.ActorUserID, "joining", j.ID, "created", map[string]interface{}{"candidateId": input.CandidateID, "requirementId": input.RequirementID}); err != nil {
		return nil, fmt.Errorf("write joining audit event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return j, nil
}

func (s *Service) Get(tenantID string, candidateID, requirementID int) (*Joining, error) {
	return s.repo.GetByPair(tenantID, candidateID, requirementID)
}

func (s *Service) Update(tenantID string, candidateID, requirementID int, input UpdateInput) (*Joining, error) {
	if input.Joined && input.JoiningDate == nil {
		return nil, ErrJoiningDateRequired
	}
	j, err := s.repo.GetByPair(tenantID, candidateID, requirementID)
	if err != nil {
		return nil, err
	}
	j.JoiningDate = input.JoiningDate
	j.Joined = input.Joined
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := s.repo.UpdateTx(tx, j); err != nil {
		return nil, err
	}
	if err := audit.WriteTx(tx, tenantID, input.ActorUserID, "joining", j.ID, "updated", map[string]interface{}{"candidateId": candidateID, "requirementId": requirementID, "joined": input.Joined}); err != nil {
		return nil, fmt.Errorf("write joining audit event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return j, nil
}
