package offer

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("offer not found")

type Repository interface {
	Create(*Offer) error
	CreateTx(*sql.Tx, *Offer) error
	GetByPair(tenantID string, candidateID, requirementID int) (*Offer, error)
	Update(*Offer) error
	UpdateTx(*sql.Tx, *Offer) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dbConn *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: dbConn}
}

const offerSelect = `id, tenant_id, candidate_id, requirement_id, selection_id, accepted, created_at, last_modified`

func scanOffer(row *sql.Row) (*Offer, error) {
	o := &Offer{}
	if err := row.Scan(
		&o.ID,
		&o.TenantID,
		&o.CandidateID,
		&o.RequirementID,
		&o.SelectionID,
		&o.Accepted,
		&o.CreatedAt,
		&o.LastModified,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return o, nil
}

type offerQueryer interface {
	QueryRow(string, ...interface{}) *sql.Row
}

func (r *PostgresRepository) Create(o *Offer) error { return createOffer(r.db, o) }
func (r *PostgresRepository) CreateTx(tx *sql.Tx, o *Offer) error { return createOffer(tx, o) }

func createOffer(q offerQueryer, o *Offer) error {
	return q.QueryRow(`
		INSERT INTO recruitment_offers (
			tenant_id, candidate_id, requirement_id, selection_id, accepted
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, last_modified
	`,
		o.TenantID, o.CandidateID, o.RequirementID, o.SelectionID, o.Accepted,
	).Scan(&o.ID, &o.CreatedAt, &o.LastModified)
}

func (r *PostgresRepository) GetByPair(tenantID string, candidateID, requirementID int) (*Offer, error) {
	return scanOffer(r.db.QueryRow(
		`SELECT `+offerSelect+` FROM recruitment_offers
		 WHERE tenant_id = $1 AND candidate_id = $2 AND requirement_id = $3`,
		tenantID, candidateID, requirementID,
	))
}

func (r *PostgresRepository) Update(o *Offer) error { return updateOffer(r.db, o) }
func (r *PostgresRepository) UpdateTx(tx *sql.Tx, o *Offer) error { return updateOffer(tx, o) }

func updateOffer(q offerQueryer, o *Offer) error {
	return q.QueryRow(`
		UPDATE recruitment_offers
		SET accepted = $1, last_modified = CURRENT_TIMESTAMP
		WHERE id = $2 AND tenant_id = $3
		RETURNING last_modified
	`, o.Accepted, o.ID, o.TenantID).Scan(&o.LastModified)
}
