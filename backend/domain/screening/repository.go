package screening

import (
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("recruitment screening not found")

type Repository interface {
	Create(*Screening) error
	GetByID(tenantID string, id int) (*Screening, error)
	ListByCandidateRequirement(tenantID string, candidateID, requirementID int) ([]*Screening, error)
}

type PostgresRepository struct { db *sql.DB }

func NewPostgresRepository(dbConn *sql.DB) *PostgresRepository { return &PostgresRepository{db: dbConn} }

const screeningSelect = `id, tenant_id, candidate_id, requirement_id, recruiter_user_id,
	current_ctc, expected_ctc, notice_period, last_working_day,
	current_location, willing_to_relocate, preferred_location,
	reason_for_change, offers_in_hand, candidate_interest, availability_date,
	relevant_experience, recruiter_assessment, notes, screened_at, created_at`

func scanScreening(row *sql.Row) (*Screening, error) {
	s := &Screening{}
	var lastWorkingDay, availabilityDate sql.NullTime
	var willingToRelocate sql.NullBool
	err := row.Scan(
		&s.ID, &s.TenantID, &s.CandidateID, &s.RequirementID, &s.RecruiterUserID,
		&s.CurrentCTC, &s.ExpectedCTC, &s.NoticePeriod, &lastWorkingDay,
		&s.CurrentLocation, &willingToRelocate, &s.PreferredLocation,
		&s.ReasonForChange, &s.OffersInHand, &s.CandidateInterest, &availabilityDate,
		&s.RelevantExperience, &s.RecruiterAssessment, &s.Notes, &s.ScreenedAt, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }
		return nil, err
	}
	if lastWorkingDay.Valid { t := lastWorkingDay.Time; s.LastWorkingDay = &t }
	if willingToRelocate.Valid { v := willingToRelocate.Bool; s.WillingToRelocate = &v }
	if availabilityDate.Valid { t := availabilityDate.Time; s.AvailabilityDate = &t }
	return s, nil
}

func (r *PostgresRepository) Create(s *Screening) error {
	return r.db.QueryRow(`
		INSERT INTO recruitment_screenings (
			tenant_id, candidate_id, requirement_id, recruiter_user_id,
			current_ctc, expected_ctc, notice_period, last_working_day,
			current_location, willing_to_relocate, preferred_location,
			reason_for_change, offers_in_hand, candidate_interest, availability_date,
			relevant_experience, recruiter_assessment, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		RETURNING id, screened_at, created_at`,
		s.TenantID, s.CandidateID, s.RequirementID, s.RecruiterUserID,
		nullableString(s.CurrentCTC), nullableString(s.ExpectedCTC), nullableString(s.NoticePeriod),
		nullableTime(s.LastWorkingDay), nullableString(s.CurrentLocation),
		nullableBool(s.WillingToRelocate), nullableString(s.PreferredLocation),
		nullableString(s.ReasonForChange), nullableString(s.OffersInHand),
		nullableString(s.CandidateInterest), nullableTime(s.AvailabilityDate),
		nullableString(s.RelevantExperience), nullableString(s.RecruiterAssessment),
		nullableString(s.Notes),
	).Scan(&s.ID, &s.ScreenedAt, &s.CreatedAt)
}

func (r *PostgresRepository) GetByID(tenantID string, id int) (*Screening, error) {
	return scanScreening(r.db.QueryRow(`SELECT `+screeningSelect+` FROM recruitment_screenings WHERE id = $1 AND tenant_id = $2`, id, tenantID))
}

func (r *PostgresRepository) ListByCandidateRequirement(tenantID string, candidateID, requirementID int) ([]*Screening, error) {
	rows, err := r.db.Query(`SELECT `+screeningSelect+` FROM recruitment_screenings
		WHERE tenant_id = $1 AND candidate_id = $2 AND requirement_id = $3
		ORDER BY screened_at DESC, id DESC`, tenantID, candidateID, requirementID)
	if err != nil { return nil, err }
	defer rows.Close()
	results := []*Screening{}
	for rows.Next() {
		s := &Screening{}
		var lastWorkingDay, availabilityDate sql.NullTime
		var willingToRelocate sql.NullBool
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.CandidateID, &s.RequirementID, &s.RecruiterUserID,
			&s.CurrentCTC, &s.ExpectedCTC, &s.NoticePeriod, &lastWorkingDay,
			&s.CurrentLocation, &willingToRelocate, &s.PreferredLocation,
			&s.ReasonForChange, &s.OffersInHand, &s.CandidateInterest, &availabilityDate,
			&s.RelevantExperience, &s.RecruiterAssessment, &s.Notes, &s.ScreenedAt, &s.CreatedAt,
		); err != nil { return nil, err }
		if lastWorkingDay.Valid { t := lastWorkingDay.Time; s.LastWorkingDay = &t }
		if willingToRelocate.Valid { v := willingToRelocate.Bool; s.WillingToRelocate = &v }
		if availabilityDate.Valid { t := availabilityDate.Time; s.AvailabilityDate = &t }
		results = append(results, s)
	}
	return results, rows.Err()
}

func nullableString(v string) interface{} { if v == "" { return nil }; return v }
func nullableTime(v *time.Time) interface{} { if v == nil { return nil }; return *v }
func nullableBool(v *bool) interface{} { if v == nil { return nil }; return *v }
