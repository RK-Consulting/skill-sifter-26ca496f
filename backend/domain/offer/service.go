package offer

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/RK-Consulting/skill-sifter/domain/audit"
	"github.com/lib/pq"
)

var (
	ErrCandidateRequirementNotFound = errors.New("candidate or requirement not found")
	ErrSelectionNotFound            = errors.New("selected recruitment decision not found")
	ErrOfferExists                  = errors.New("offer already exists for candidate and requirement")
)

type Service struct {
	repo Repository
	db   *sql.DB
}

func NewService(repo Repository, dbConn *sql.DB) *Service {
	return &Service{repo: repo, db: dbConn}
}

func (s *Service) Create(tenantID string, input CreateInput) (*Offer, error) {
	if input.CandidateID == 0 || input.RequirementID == 0 {
		return nil, fmt.Errorf("candidateId and requirementId are required")
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
		return nil, ErrOfferExists
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	o := &Offer{
		TenantID:      tenantID,
		CandidateID:   input.CandidateID,
		RequirementID: input.RequirementID,
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var decision string
	if err := tx.QueryRow(`
		SELECT id, decision
		FROM recruitment_selections
		WHERE tenant_id = $1 AND candidate_id = $2 AND requirement_id = $3
		FOR UPDATE
	`, tenantID, input.CandidateID, input.RequirementID).Scan(&o.SelectionID, &decision); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSelectionNotFound
		}
		return nil, err
	}
	if decision != "selected" {
		return nil, ErrSelectionNotFound
	}

	if err := s.repo.CreateTx(tx, o); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrOfferExists
		}
		return nil, err
	}
	if err := audit.WriteTx(tx, tenantID, input.ActorUserID, "offer", o.ID, "created", map[string]interface{}{"candidateId": input.CandidateID, "requirementId": input.RequirementID}); err != nil {
		return nil, fmt.Errorf("write offer audit event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *Service) Get(tenantID string, candidateID, requirementID int) (*Offer, error) {
	return s.repo.GetByPair(tenantID, candidateID, requirementID)
}

func (s *Service) Update(tenantID string, candidateID, requirementID int, input UpdateInput) (*Offer, error) {
	o, err := s.repo.GetByPair(tenantID, candidateID, requirementID)
	if err != nil {
		return nil, err
	}
	o.Accepted = input.Accepted
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := s.repo.UpdateTx(tx, o); err != nil {
		return nil, err
	}
	if err := audit.WriteTx(tx, tenantID, input.ActorUserID, "offer", o.ID, "updated", map[string]interface{}{"candidateId": candidateID, "requirementId": requirementID, "accepted": input.Accepted}); err != nil {
		return nil, fmt.Errorf("write offer audit event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return o, nil
}
