package billing

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("billing record not found")

type Repository interface {
	Create(*Billing) error
	GetByPair(tenantID string, candidateID, requirementID int) (*Billing, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dbConn *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: dbConn}
}

const billingSelect = `id, tenant_id, candidate_id, requirement_id, client_id, joining_id,
	billing_date, amount, currency, invoice_reference, created_at, last_modified`

func scanBilling(row *sql.Row) (*Billing, error) {
	b := &Billing{}
	if err := row.Scan(
		&b.ID,
		&b.TenantID,
		&b.CandidateID,
		&b.RequirementID,
		&b.ClientID,
		&b.JoiningID,
		&b.BillingDate,
		&b.Amount,
		&b.Currency,
		&b.InvoiceReference,
		&b.CreatedAt,
		&b.LastModified,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

func (r *PostgresRepository) Create(b *Billing) error {
	return r.db.QueryRow(`
		INSERT INTO recruitment_billings (
			tenant_id, candidate_id, requirement_id, client_id, joining_id,
			billing_date, amount, currency, invoice_reference
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at, last_modified
	`,
		b.TenantID, b.CandidateID, b.RequirementID, b.ClientID, b.JoiningID,
		b.BillingDate, b.Amount, b.Currency, nullString(b.InvoiceReference),
	).Scan(&b.ID, &b.CreatedAt, &b.LastModified)
}

func (r *PostgresRepository) GetByPair(tenantID string, candidateID, requirementID int) (*Billing, error) {
	return scanBilling(r.db.QueryRow(
		`SELECT `+billingSelect+` FROM recruitment_billings
		 WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3`,
		tenantID, candidateID, requirementID,
	))
}

func nullString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
