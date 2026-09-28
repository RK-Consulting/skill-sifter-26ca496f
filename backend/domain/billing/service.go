package billing

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrCandidateRequirementNotFound = errors.New("candidate or requirement not found")
	ErrJoiningNotFound              = errors.New("joined recruitment record not found")
	ErrBillingExists                = errors.New("billing record already exists for candidate and requirement")
	ErrInvalidAmount                = errors.New("billing amount is required")
	ErrInvalidCurrency              = errors.New("billing currency must be a three-letter code")
)

var (
	currencyPattern = regexp.MustCompile(`^[A-Za-z]{3}$`)
	amountPattern   = regexp.MustCompile(`^[0-9]+(\.[0-9]{1,2})?package billing

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrCandidateRequirementNotFound = errors.New("candidate or requirement not found")
	ErrJoiningNotFound              = errors.New("joined recruitment record not found")
	ErrBillingExists                = errors.New("billing record already exists for candidate and requirement")
	ErrInvalidAmount                = errors.New("billing amount is required")
	ErrInvalidCurrency              = errors.New("billing currency must be a three-letter code")
)

var (
	currencyPattern = regexp.MustCompile(`^[A-Za-z]{3}$`)
	)
)

type Service struct {
	repo Repository
	db   *sql.DB
}

func NewService(repo Repository, dbConn *sql.DB) *Service {
	return &Service{repo: repo, db: dbConn}
}

func (s *Service) Create(tenantID string, input CreateInput) (*Billing, error) {
	if input.CandidateID == 0 || input.RequirementID == 0 {
		return nil, fmt.Errorf("candidateId and requirementId are required")
	}

	input.Amount = strings.TrimSpace(input.Amount)
	amountDigits := strings.ReplaceAll(input.Amount, ".", "")
	if !amountPattern.MatchString(input.Amount) || strings.TrimLeft(amountDigits, "0") == "" {
		return nil, ErrInvalidAmount
	}

	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	if !currencyPattern.MatchString(input.Currency) {
		return nil, ErrInvalidCurrency
	}

	var exists bool
	if err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM candidates c
			JOIN requirements r ON r.tenant_id = c.tenant_id
			WHERE c.id=$1 AND r.id=$2
			  AND c.tenant_id=$3 AND r.tenant_id=$3
		)
	`, input.CandidateID, input.RequirementID, tenantID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCandidateRequirementNotFound
	}

	var joiningID int
	var joiningDate sql.NullTime
	var joined bool
	if err := s.db.QueryRow(`
		SELECT id, joining_date, joined
		FROM recruitment_joinings
		WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3
	`, tenantID, input.CandidateID, input.RequirementID).Scan(&joiningID, &joiningDate, &joined); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrJoiningNotFound
		}
		return nil, err
	}
	if !joined || !joiningDate.Valid {
		return nil, ErrJoiningNotFound
	}

	var clientID int
	if err := s.db.QueryRow(`
		SELECT client_id
		FROM requirements
		WHERE tenant_id=$1 AND id=$2
	`, tenantID, input.RequirementID).Scan(&clientID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCandidateRequirementNotFound
		}
		return nil, err
	}

	if _, err := s.repo.GetByPair(tenantID, input.CandidateID, input.RequirementID); err == nil {
		return nil, ErrBillingExists
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	b := &Billing{
		TenantID:         tenantID,
		CandidateID:      input.CandidateID,
		RequirementID:    input.RequirementID,
		ClientID:         clientID,
		JoiningID:        joiningID,
		BillingDate:      joiningDate.Time,
		Amount:           input.Amount,
		Currency:         input.Currency,
		InvoiceReference: strings.TrimSpace(input.InvoiceReference),
	}
	if err := s.repo.Create(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) Get(tenantID string, candidateID, requirementID int) (*Billing, error) {
	return s.repo.GetByPair(tenantID, candidateID, requirementID)
}

func (s *Service) ListWorklist(tenantID string) ([]WorklistItem, error) {
	return s.repo.ListWorklist(tenantID)
}
