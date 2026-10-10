package joining

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("joining record not found")

type Repository interface {
	Create(*Joining) error
	CreateTx(*sql.Tx, *Joining) error
	GetByPair(tenantID string, candidateID, requirementID int) (*Joining, error)
	Update(*Joining) error
	UpdateTx(*sql.Tx, *Joining) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dbConn *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: dbConn}
}

const joiningSelect = `id, tenant_id, candidate_id, requirement_id, offer_id, joining_date, joined, created_at, last_modified`

func scanJoining(row *sql.Row) (*Joining, error) {
	j := &Joining{}
	if err := row.Scan(
		&j.ID,
		&j.TenantID,
		&j.CandidateID,
		&j.RequirementID,
		&j.OfferID,
		&j.JoiningDate,
		&j.Joined,
		&j.CreatedAt,
		&j.LastModified,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return j, nil
}

func (r *PostgresRepository) Create(j *Joining) error { return createJoining(r.db, j) }
func (r *PostgresRepository) CreateTx(tx *sql.Tx, j *Joining) error { return createJoining(tx, j) }

type joiningInserter interface { QueryRow(string, ...interface{}) *sql.Row }

func createJoining(q joiningInserter, j *Joining) error {
	return q.QueryRow(`
		INSERT INTO recruitment_joinings (
			tenant_id, candidate_id, requirement_id, offer_id, joining_date, joined
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, last_modified
	`,
		j.TenantID, j.CandidateID, j.RequirementID, j.OfferID,
		j.JoiningDate, j.Joined,
	).Scan(&j.ID, &j.CreatedAt, &j.LastModified)
}

func (r *PostgresRepository) GetByPair(tenantID string, candidateID, requirementID int) (*Joining, error) {
	return scanJoining(r.db.QueryRow(
		`SELECT `+joiningSelect+` FROM recruitment_joinings
		 WHERE tenant_id = $1 AND candidate_id = $2 AND requirement_id = $3`,
		tenantID, candidateID, requirementID,
	))
}

func (r *PostgresRepository) Update(j *Joining) error { return updateJoining(r.db, j) }
func (r *PostgresRepository) UpdateTx(tx *sql.Tx, j *Joining) error { return updateJoining(tx, j) }

func updateJoining(q joiningInserter, j *Joining) error {
	return q.QueryRow(`
		UPDATE recruitment_joinings
		SET joining_date = $1, joined = $2, last_modified = CURRENT_TIMESTAMP
		WHERE id = $3 AND tenant_id = $4
		RETURNING last_modified
	`, j.JoiningDate, j.Joined, j.ID, j.TenantID).Scan(&j.LastModified)
}
