package submission

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("recruitment submission not found")

type Repository interface {
	Create(*Submission) error
	GetByID(tenantID string, id int) (*Submission, error)
	ListByAssignment(tenantID string, assignmentID int) ([]*Submission, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dbConn *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: dbConn}
}

const submissionSelect = `id, tenant_id, assignment_id, submitted_by_user_id,
	recipient_type, recipient_client_id, recipient_user_id, recipient_name,
	recipient_email, submission_context, recruiter_notes,
	candidate_snapshot, requirement_snapshot, submitted_at, created_at`

func scanSubmission(row *sql.Row) (*Submission, error) {
	s := &Submission{}
	var clientID, userID sql.NullInt64
	if err := row.Scan(
		&s.ID, &s.TenantID, &s.AssignmentID, &s.SubmittedByUserID,
		&s.RecipientType, &clientID, &userID, &s.RecipientName,
		&s.RecipientEmail, &s.SubmissionContext, &s.RecruiterNotes,
		&s.CandidateSnapshot, &s.RequirementSnapshot, &s.SubmittedAt, &s.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }
		return nil, err
	}
	if clientID.Valid {
		v := int(clientID.Int64)
		s.RecipientClientID = &v
	}
	if userID.Valid {
		v := int(userID.Int64)
		s.RecipientUserID = &v
	}
	return s, nil
}

func (r *PostgresRepository) Create(s *Submission) error {
	return r.db.QueryRow(`
		INSERT INTO recruitment_submissions (
			tenant_id, assignment_id, submitted_by_user_id,
			recipient_type, recipient_client_id, recipient_user_id,
			recipient_name, recipient_email, submission_context, recruiter_notes,
			candidate_snapshot, requirement_snapshot
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, submitted_at, created_at`,
		s.TenantID, s.AssignmentID, s.SubmittedByUserID, s.RecipientType,
		nullableInt(s.RecipientClientID), nullableInt(s.RecipientUserID),
		nullableString(s.RecipientName), nullableString(s.RecipientEmail),
		nullableString(s.SubmissionContext), nullableString(s.RecruiterNotes),
		s.CandidateSnapshot, s.RequirementSnapshot,
	).Scan(&s.ID, &s.SubmittedAt, &s.CreatedAt)
}

func (r *PostgresRepository) GetByID(tenantID string, id int) (*Submission, error) {
	return scanSubmission(r.db.QueryRow(
		`SELECT `+submissionSelect+` FROM recruitment_submissions
		 WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	))
}

func (r *PostgresRepository) ListByAssignment(tenantID string, assignmentID int) ([]*Submission, error) {
	rows, err := r.db.Query(
		`SELECT `+submissionSelect+` FROM recruitment_submissions
		 WHERE tenant_id = $1 AND assignment_id = $2
		 ORDER BY submitted_at DESC, id DESC`, tenantID, assignmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*Submission
	for rows.Next() {
		s := &Submission{}
		var clientID, userID sql.NullInt64
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.AssignmentID, &s.SubmittedByUserID,
			&s.RecipientType, &clientID, &userID, &s.RecipientName,
			&s.RecipientEmail, &s.SubmissionContext, &s.RecruiterNotes,
			&s.CandidateSnapshot, &s.RequirementSnapshot, &s.SubmittedAt, &s.CreatedAt,
		); err != nil { return nil, err }
		if clientID.Valid {
		v := int(clientID.Int64)
		s.RecipientClientID = &v
	}
		if userID.Valid {
		v := int(userID.Int64)
		s.RecipientUserID = &v
	}
		results = append(results, s)
	}
	if results == nil {
		results = []*Submission{}
	}
	return results, rows.Err()
}

func nullableString(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}

func nullableInt(v *int) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
