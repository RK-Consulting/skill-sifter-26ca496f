package feedback

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("submission feedback not found")

type Repository interface {
	Create(*Feedback) error
	GetByID(tenantID string, id int) (*Feedback, error)
	ListBySubmission(tenantID string, submissionID int) ([]*Feedback, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dbConn *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: dbConn}
}

const feedbackSelect = `id, tenant_id, submission_id, feedback_by_user_id,
	outcome, reason_code, comments, next_action, feedback_at, created_at`

func scanFeedback(row *sql.Row) (*Feedback, error) {
	f := &Feedback{}
	var reasonCode, comments, nextAction sql.NullString
	if err := row.Scan(
		&f.ID, &f.TenantID, &f.SubmissionID, &f.FeedbackByUserID,
		&f.Outcome, &reasonCode, &comments, &nextAction,
		&f.FeedbackAt, &f.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if reasonCode.Valid {
		f.ReasonCode = reasonCode.String
	}
	if comments.Valid {
		f.Comments = comments.String
	}
	if nextAction.Valid {
		f.NextAction = nextAction.String
	}
	return f, nil
}

func (r *PostgresRepository) Create(f *Feedback) error { return createFeedback(r.db, f) }
func (r *PostgresRepository) CreateTx(tx *sql.Tx, f *Feedback) error { return createFeedback(tx, f) }

type feedbackInserter interface {
	QueryRow(string, ...interface{}) *sql.Row
}

func createFeedback(q feedbackInserter, f *Feedback) error {
	return q.QueryRow(`
		INSERT INTO recruitment_submission_feedback (
			tenant_id, submission_id, feedback_by_user_id, outcome,
			reason_code, comments, next_action
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, feedback_at, created_at
	`,
		f.TenantID, f.SubmissionID, f.FeedbackByUserID, f.Outcome,
		nullableString(f.ReasonCode), nullableString(f.Comments), nullableString(f.NextAction),
	).Scan(&f.ID, &f.FeedbackAt, &f.CreatedAt)
}

func (r *PostgresRepository) GetByID(tenantID string, id int) (*Feedback, error) {
	return scanFeedback(r.db.QueryRow(
		`SELECT `+feedbackSelect+` FROM recruitment_submission_feedback
		 WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	))
}

func (r *PostgresRepository) ListBySubmission(tenantID string, submissionID int) ([]*Feedback, error) {
	rows, err := r.db.Query(
		`SELECT `+feedbackSelect+` FROM recruitment_submission_feedback
		 WHERE tenant_id = $1 AND submission_id = $2
		 ORDER BY feedback_at DESC, id DESC`, tenantID, submissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*Feedback
	for rows.Next() {
		f := &Feedback{}
		var reasonCode, comments, nextAction sql.NullString
		if err := rows.Scan(
			&f.ID, &f.TenantID, &f.SubmissionID, &f.FeedbackByUserID,
			&f.Outcome, &reasonCode, &comments, &nextAction,
			&f.FeedbackAt, &f.CreatedAt,
		); err != nil {
			return nil, err
		}
		if reasonCode.Valid {
			f.ReasonCode = reasonCode.String
		}
		if comments.Valid {
			f.Comments = comments.String
		}
		if nextAction.Valid {
			f.NextAction = nextAction.String
		}
		results = append(results, f)
	}
	if results == nil {
		results = []*Feedback{}
	}
	return results, rows.Err()
}

func nullableString(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}
