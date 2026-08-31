package assignment

import (
	"database/sql"
	"errors"
)

// CandidateSnapshotData is the immutable candidate-facing content captured
// at formal submission (ADR 0003 section 6): relevant candidate
// identity/profile, language expertise, and technical expertise.
//
// This is a point-in-time snapshot. It is NOT a live reference to the
// candidate record and does not change when the candidate is subsequently
// edited.
type CandidateSnapshotData struct {
	ID                 int                          `json:"id"`
	Name               string                       `json:"name"`
	Email              string                       `json:"email"`
	Phone              string                       `json:"phone,omitempty"`
	Position           string                       `json:"position,omitempty"`
	Location           string                       `json:"location,omitempty"`
	Experience         string                       `json:"experience,omitempty"`
	CurrentCTC         string                       `json:"currentCtc,omitempty"`
	ExpectedCTC        string                       `json:"expectedCtc,omitempty"`
	NoticePeriod       string                       `json:"noticePeriod,omitempty"`
	LanguageExpertise  []LanguageExpertiseSnapshot  `json:"languageExpertise,omitempty"`
	TechnicalExpertise []TechnicalExpertiseSnapshot `json:"technicalExpertise,omitempty"`
	Status             string                       `json:"status"`
}

// LanguageExpertiseSnapshot represents language expertise captured as part
// of the immutable candidate submission snapshot.
type LanguageExpertiseSnapshot struct {
	Language             string `json:"language"`
	ProficiencyFramework string `json:"proficiencyFramework"`
	ProficiencyLevel     string `json:"proficiencyLevel"`
}

// TechnicalExpertiseSnapshot represents technical expertise captured as
// part of the immutable candidate submission snapshot.
type TechnicalExpertiseSnapshot struct {
	Skill            string `json:"skill"`
	Category         string `json:"category"`
	ProficiencyLevel string `json:"proficiencyLevel"`
}

// RequirementSnapshotData is the immutable requirement-facing content
// captured at formal submission (ADR 0003 section 6): title, location,
// work arrangement, description, required skills, experience,
// compensation, headcount, and language requirement.
type RequirementSnapshotData struct {
	ID                  int    `json:"id"`
	Title               string `json:"title"`
	Location            string `json:"location,omitempty"`
	WorkArrangement     string `json:"workArrangement,omitempty"`
	Description         string `json:"description,omitempty"`
	RequiredSkills      string `json:"requiredSkills,omitempty"`
	ExperienceRequired  string `json:"experienceRequired,omitempty"`
	Compensation        string `json:"compensation,omitempty"`
	Headcount           int    `json:"headcount"`
	LanguageRequirement string `json:"languageRequirement,omitempty"`
}

// ErrCandidateSnapshotSourceNotFound / ErrRequirementSnapshotSourceNotFound
// are returned if the candidate/requirement row referenced by the
// assignment cannot be found at the moment of snapshot capture.
var (
	ErrCandidateSnapshotSourceNotFound   = errors.New("candidate record for snapshot not found")
	ErrRequirementSnapshotSourceNotFound = errors.New("requirement record for snapshot not found")
)

// fetchCandidateSnapshotTx reads the current candidate row and its current
// expertise within the same transaction, for embedding into an immutable
// submission snapshot.
//
// Candidate identity/profile data comes from candidates.
// Language expertise comes from candidate_language_expertise.
// Technical expertise comes from candidate_expertise.
//
// All reads are tenant-scoped.
func fetchCandidateSnapshotTx(
	tx *sql.Tx,
	candidateID int,
	tenantID string,
) (*CandidateSnapshotData, error) {
	c := &CandidateSnapshotData{}

	err := tx.QueryRow(`
		SELECT
			id,
			name,
			email,
			COALESCE(phone, ''),
			COALESCE(position, ''),
			COALESCE(location, ''),
			COALESCE(experience, ''),
			COALESCE(currentctc, ''),
			COALESCE(expectedctc, ''),
			COALESCE(noticeperiod, ''),
			status
		FROM candidates
		WHERE id = $1
		  AND tenant_id = $2`,
		candidateID,
		tenantID,
	).Scan(
		&c.ID,
		&c.Name,
		&c.Email,
		&c.Phone,
		&c.Position,
		&c.Location,
		&c.Experience,
		&c.CurrentCTC,
		&c.ExpectedCTC,
		&c.NoticePeriod,
		&c.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCandidateSnapshotSourceNotFound
		}

		return nil, err
	}

	languageRows, err := tx.Query(`
		SELECT
			language,
			proficiency_framework,
			proficiency_level
		FROM candidate_language_expertise
		WHERE candidate_id = $1
		  AND tenant_id = $2
		ORDER BY id`,
		candidateID,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer languageRows.Close()

	c.LanguageExpertise = make([]LanguageExpertiseSnapshot, 0)

	for languageRows.Next() {
		var item LanguageExpertiseSnapshot

		if err := languageRows.Scan(
			&item.Language,
			&item.ProficiencyFramework,
			&item.ProficiencyLevel,
		); err != nil {
			return nil, err
		}

		c.LanguageExpertise = append(c.LanguageExpertise, item)
	}

	if err := languageRows.Err(); err != nil {
		return nil, err
	}

	expertiseRows, err := tx.Query(`
		SELECT
			skill,
			category,
			proficiency_level
		FROM candidate_expertise
		WHERE candidate_id = $1
		  AND tenant_id = $2
		ORDER BY id`,
		candidateID,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer expertiseRows.Close()

	c.TechnicalExpertise = make([]TechnicalExpertiseSnapshot, 0)

	for expertiseRows.Next() {
		var item TechnicalExpertiseSnapshot

		if err := expertiseRows.Scan(
			&item.Skill,
			&item.Category,
			&item.ProficiencyLevel,
		); err != nil {
			return nil, err
		}

		c.TechnicalExpertise = append(c.TechnicalExpertise, item)
	}

	if err := expertiseRows.Err(); err != nil {
		return nil, err
	}

	return c, nil
}

// fetchRequirementSnapshotTx reads the current requirement row within tx,
// for embedding into an immutable submission snapshot.
func fetchRequirementSnapshotTx(
	tx *sql.Tx,
	requirementID int,
	tenantID string,
) (*RequirementSnapshotData, error) {
	r := &RequirementSnapshotData{}

	err := tx.QueryRow(`
		SELECT
			id,
			title,
			COALESCE(location, ''),
			COALESCE(work_arrangement, ''),
			COALESCE(description, ''),
			COALESCE(required_skills, ''),
			COALESCE(experience_required, ''),
			COALESCE(compensation, ''),
			headcount,
			COALESCE(language_requirement, '')
		FROM requirements
		WHERE id = $1
		  AND tenant_id = $2`,
		requirementID,
		tenantID,
	).Scan(
		&r.ID,
		&r.Title,
		&r.Location,
		&r.WorkArrangement,
		&r.Description,
		&r.RequiredSkills,
		&r.ExperienceRequired,
		&r.Compensation,
		&r.Headcount,
		&r.LanguageRequirement,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRequirementSnapshotSourceNotFound
		}

		return nil, err
	}

	return r, nil
}
