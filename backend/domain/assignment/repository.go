package assignment

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// ErrNotFound is returned by repository lookups when no matching row
// exists for the given (id, tenant_id) pair — deliberately the same error
// whether the row doesn't exist at all or exists in a different tenant, so
// callers cannot distinguish the two (ADR 0003 section 8 / the #33
// tenant-isolation convention: cross-tenant access must fail closed and
// must not disclose existence).
var ErrNotFound = errors.New("recruitment assignment not found")

// ErrDuplicateAssignment is returned when attempting to create an
// assignment for a (candidate_id, requirement_id) pair that already has
// one, per ADR 0003 section 2's uniqueness rule (enforced at the database
// level by the recruitment_assignments_candidate_requirement_unique
// constraint; this error is the repository's translation of that
// constraint violation into a typed error the service layer can act on).
var ErrDuplicateAssignment = errors.New("an assignment already exists for this candidate and requirement")

// Repository persists and retrieves Assignments. It knows nothing about
// business rules (transition validity, tenant consistency across
// candidate/requirement/owner, candidate eligibility) — that is the
// service layer's responsibility. The repository's only job is mapping
// between the domain type and the recruitment_assignments table.
type Repository interface {
	// Create inserts a new assignment and populates a.ID, a.CreatedAt,
	// a.LastModified on success. Returns ErrDuplicateAssignment if a
	// (candidate_id, requirement_id) assignment already exists.
	Create(a *Assignment) error

	// GetByID returns the assignment with the given id, scoped to
	// tenantID. Returns ErrNotFound if no such assignment exists in that
	// tenant (including if it exists in a different tenant).
	GetByID(tenantID string, id int) (*Assignment, error)

	// ListByTenant returns every assignment belonging to tenantID, most
	// recently created first.
	ListByTenant(tenantID string) ([]*Assignment, error)

	// Update persists the assignment's current status, owner, and
	// snapshot fields, scoped to (a.ID, a.TenantID), and refreshes
	// LastModified. Returns ErrNotFound if no matching row exists in that
	// tenant.
	Update(a *Assignment) error
}

// PostgresRepository is the Repository implementation backed by
// recruitment_assignments.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository constructs a PostgresRepository using the given
// database connection (callers pass db.DB from the db package; this
// package takes the connection explicitly rather than importing db
// directly, to keep the domain layer decoupled from the application's
// specific database wiring).
func NewPostgresRepository(dbConn *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: dbConn}
}

func (r *PostgresRepository) Create(a *Assignment) error {
	err := r.db.QueryRow(`
		INSERT INTO recruitment_assignments
			(tenant_id, candidate_id, requirement_id, status, created_by_user_id, owner_user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, last_modified`,
		a.TenantID, a.CandidateID, a.RequirementID, string(a.Status), a.CreatedByUserID, a.OwnerUserID,
	).Scan(&a.ID, &a.CreatedAt, &a.LastModified)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return ErrDuplicateAssignment
		}
		return err
	}
	return nil
}

func (r *PostgresRepository) GetByID(tenantID string, id int) (*Assignment, error) {
	a := &Assignment{}
	var status string
	var candidateSnapshot, requirementSnapshot []byte
	var snapshotCreatedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, tenant_id, candidate_id, requirement_id, status, created_by_user_id, owner_user_id,
			candidate_snapshot, requirement_snapshot, snapshot_created_at, created_at, last_modified
		FROM recruitment_assignments WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	).Scan(&a.ID, &a.TenantID, &a.CandidateID, &a.RequirementID, &status, &a.CreatedByUserID, &a.OwnerUserID,
		&candidateSnapshot, &requirementSnapshot, &snapshotCreatedAt, &a.CreatedAt, &a.LastModified)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	a.Status = Status(status)
	a.CandidateSnapshot = candidateSnapshot
	a.RequirementSnapshot = requirementSnapshot
	if snapshotCreatedAt.Valid {
		t := snapshotCreatedAt.Time
		a.SnapshotCreatedAt = &t
	}

	return a, nil
}

func (r *PostgresRepository) ListByTenant(tenantID string) ([]*Assignment, error) {
	rows, err := r.db.Query(`
		SELECT id, tenant_id, candidate_id, requirement_id, status, created_by_user_id, owner_user_id,
			candidate_snapshot, requirement_snapshot, snapshot_created_at, created_at, last_modified
		FROM recruitment_assignments WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*Assignment
	for rows.Next() {
		a := &Assignment{}
		var status string
		var candidateSnapshot, requirementSnapshot []byte
		var snapshotCreatedAt sql.NullTime

		if err := rows.Scan(&a.ID, &a.TenantID, &a.CandidateID, &a.RequirementID, &status, &a.CreatedByUserID, &a.OwnerUserID,
			&candidateSnapshot, &requirementSnapshot, &snapshotCreatedAt, &a.CreatedAt, &a.LastModified); err != nil {
			return nil, err
		}

		a.Status = Status(status)
		a.CandidateSnapshot = candidateSnapshot
		a.RequirementSnapshot = requirementSnapshot
		if snapshotCreatedAt.Valid {
			t := snapshotCreatedAt.Time
			a.SnapshotCreatedAt = &t
		}

		results = append(results, a)
	}
	if results == nil {
		results = []*Assignment{}
	}
	return results, rows.Err()
}

func (r *PostgresRepository) Update(a *Assignment) error {
	var snapshotCreatedAt interface{}
	if a.SnapshotCreatedAt != nil {
		snapshotCreatedAt = *a.SnapshotCreatedAt
	}

	// A nil []byte passed directly as a query arg does NOT become SQL NULL
	// for a JSONB column via lib/pq (it errors with "invalid input syntax
	// for type json") the way an untyped nil interface{} does. Convert
	// explicitly so an unset snapshot persists as NULL, not a driver error.
	var candidateSnapshot, requirementSnapshot interface{}
	if a.CandidateSnapshot != nil {
		candidateSnapshot = a.CandidateSnapshot
	}
	if a.RequirementSnapshot != nil {
		requirementSnapshot = a.RequirementSnapshot
	}

	result, err := r.db.Exec(`
		UPDATE recruitment_assignments
		SET status = $1, owner_user_id = $2, candidate_snapshot = $3, requirement_snapshot = $4,
			snapshot_created_at = $5, last_modified = NOW()
		WHERE id = $6 AND tenant_id = $7`,
		string(a.Status), a.OwnerUserID, candidateSnapshot, requirementSnapshot, snapshotCreatedAt,
		a.ID, a.TenantID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
