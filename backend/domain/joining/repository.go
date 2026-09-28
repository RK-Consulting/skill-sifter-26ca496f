package joining

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("joining record not found")

type Repository interface {
	Create(*Joining) error
	GetByPair(tenantID string, candidateID, requirementID int) (*Joining, error)
	Update(*Joining) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dbConn *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: dbConn}
}

const joiningSelect = `id, tenant_id, candidate_id, requirement_id, offer_id,
	status, expected_joining_date, actual_joining_date, notes, created_at, last_modified`

func scanJoining(row *sql.Row) (*Joining, error) {
	j := &Joining{}
	var notes sql.NullString
	if err := row.Scan(
		&j.ID, &j.TenantID, &j.CandidateID, &j.RequirementID, &j.OfferID,
		&j.Status, &j.ExpectedJoiningDate, &j.ActualJoiningDate, &notes,
		&j.CreatedAt, &j.LastModified,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	j.Notes = notes.String
	return j, nil
}

func (r *PostgresRepository) Create(j *Joining) error {
	return r.db.QueryRow(`
		INSERT INTO recruitment_joinings (
			tenant_id, candidate_id, requirement_id, offer_id, status,
			expected_joining_date, actual_joining_date, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, created_at, last_modified
	`,
		j.TenantID, j.CandidateID, j.RequirementID, j.OfferID, j.Status,
		j.ExpectedJoiningDate, j.ActualJoiningDate, nullableString(j.Notes),
	).Scan(&j.ID, &j.CreatedAt, &j.LastModified)
}

func (r *PostgresRepository) GetByPair(tenantID string, candidateID, requirementID int) (*Joining, error) {
	return scanJoining(r.db.QueryRow(
		`SELECT `+joiningSelect+` FROM recruitment_joinings
		 WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3`,
		tenantID, candidateID, requirementID,
	))
}

func (r *PostgresRepository) Update(j *Joining) error {
	return r.db.QueryRow(`
		UPDATE recruitment_joinings
		SET status=$1, expected_joining_date=$2, actual_joining_date=$3,
		    notes=$4, last_modified=CURRENT_TIMESTAMP
		WHERE id=$5 AND tenant_id=$6
		RETURNING last_modified
	`,
		j.Status, j.ExpectedJoiningDate, j.ActualJoiningDate,
		nullableString(j.Notes), j.ID, j.TenantID,
	).Scan(&j.LastModified)
}

func nullableString(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}
