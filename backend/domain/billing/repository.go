package billing

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("billing record not found")

type Repository interface {
	Create(*Billing) error
	CreateTx(*sql.Tx, *Billing) error
	GetByPair(tenantID string, candidateID, requirementID int) (*Billing, error)
	ListWorklist(tenantID string) ([]WorklistItem, error)
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
	var invoiceReference sql.NullString
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
		&invoiceReference,
		&b.CreatedAt,
		&b.LastModified,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	b.InvoiceReference = invoiceReference.String
	return b, nil
}

func (r *PostgresRepository) Create(b *Billing) error { return createBilling(r.db, b) }
func (r *PostgresRepository) CreateTx(tx *sql.Tx, b *Billing) error { return createBilling(tx, b) }

type billingInserter interface { QueryRow(string, ...interface{}) *sql.Row }

func createBilling(q billingInserter, b *Billing) error {
	return q.QueryRow(`
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

func (r *PostgresRepository) ListWorklist(tenantID string) ([]WorklistItem, error) {
	rows, err := r.db.Query(`
		SELECT
			c.id,
			c.name,
			r.id,

			r.title,
			cl.id,
			cl.name,
			j.id,
			j.joining_date,
			b.id,
			b.billing_date,
			b.amount,
			b.currency,
			b.invoice_reference
		FROM recruitment_joinings j
		JOIN candidates c
		  ON c.id = j.candidate_id AND c.tenant_id = j.tenant_id
		JOIN requirements r
		  ON r.id = j.requirement_id AND r.tenant_id = j.tenant_id
		JOIN clients cl
		  ON cl.id = r.client_id AND cl.tenant_id = j.tenant_id
		LEFT JOIN recruitment_billings b
		  ON b.tenant_id = j.tenant_id
		 AND b.candidate_id = j.candidate_id
		 AND b.requirement_id = j.requirement_id
		WHERE j.tenant_id = $1
		  AND j.joined = TRUE
		ORDER BY j.joining_date DESC, c.name, r.title`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]WorklistItem, 0)
	for rows.Next() {
		var item WorklistItem
		var billingID sql.NullInt64
		var billingDate sql.NullTime
		var amount, currency, invoiceReference sql.NullString

		if err := rows.Scan(
			&item.CandidateID,
			&item.CandidateName,
			&item.RequirementID,
			&item.RequirementTitle,
			&item.ClientID,
			&item.ClientName,
			&item.JoiningID,
			&item.JoiningDate,
			&billingID,
			&billingDate,
			&amount,
			&currency,
			&invoiceReference,
		); err != nil {
			return nil, err
		}

		item.Billed = billingID.Valid
		if billingID.Valid {
			id := int(billingID.Int64)
			item.BillingID = &id
		}
		if billingDate.Valid {
			value := billingDate.Time
			item.BillingDate = &value
		}
		if amount.Valid {
			value := amount.String
			item.Amount = &value
		}
		if currency.Valid {
			value := currency.String
			item.Currency = &value
		}
		if invoiceReference.Valid {
			value := invoiceReference.String
			item.InvoiceReference = &value
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func nullString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
