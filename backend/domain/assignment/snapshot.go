package assignment

import (
	"database/sql"
	"errors"
)

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
	LanguageExpertise  []LanguageExpertiseSnapshot  `json:"languageExpertise"`
	TechnicalExpertise []TechnicalExpertiseSnapshot `json:"technicalExpertise"`
	Status             string                       `json:"status"`
}

type LanguageExpertiseSnapshot struct {
	Language             string `json:"language"`
	ProficiencyFramework string `json:"proficiencyFramework"`
	ProficiencyLevel     string `json:"proficiencyLevel"`
}

type TechnicalExpertiseSnapshot struct {
	Skill            string `json:"skill"`
	Category         string `json:"category"`
	ProficiencyLevel string `json:"proficiencyLevel"`
}

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

var (
	ErrCandidateSnapshotSourceNotFound   = errors.New("candidate record for snapshot not found")
	ErrRequirementSnapshotSourceNotFound = errors.New("requirement record for snapshot not found")
)

func fetchCandidateSnapshotTx(tx *sql.Tx, candidateID int, tenantID string) (*CandidateSnapshotData, error) {
	c := &CandidateSnapshotData{
		LanguageExpertise:  make([]LanguageExpertiseSnapshot, 0),
		TechnicalExpertise: make([]TechnicalExpertiseSnapshot, 0),
	}

	err := tx.QueryRow(`
		SELECT id, name, email,
			COALESCE(phone, ''), COALESCE(position, ''), COALESCE(location, ''),
			COALESCE(experience, ''), COALESCE(currentctc, ''), COALESCE(expectedctc, ''),
			COALESCE(noticeperiod, ''), status
		FROM candidates
		WHERE id = $1 AND tenant_id = $2`, candidateID, tenantID).Scan(
		&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, &c.Location,
		&c.Experience, &c.CurrentCTC, &c.ExpectedCTC, &c.NoticePeriod, &c.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCandidateSnapshotSourceNotFound
		}
		return nil, err
	}

	languageRows, err := tx.Query(`
		SELECT language, proficiency_framework, proficiency_level
		FROM candidate_language_expertise
		WHERE candidate_id = $1 AND tenant_id = $2 ORDER BY id`, candidateID, tenantID)
	if err != nil {
		return nil, err
	}
	defer languageRows.Close()
	for languageRows.Next() {
		var item LanguageExpertiseSnapshot
		if err := languageRows.Scan(&item.Language, &item.ProficiencyFramework, &item.ProficiencyLevel); err != nil {
			return nil, err
		}
		c.LanguageExpertise = append(c.LanguageExpertise, item)
	}
	if err := languageRows.Err(); err != nil {
		return nil, err
	}

	expertiseRows, err := tx.Query(`
		SELECT skill, category, proficiency_level
		FROM candidate_expertise
		WHERE candidate_id = $1 AND tenant_id = $2 ORDER BY id`, candidateID, tenantID)
	if err != nil {
		return nil, err
	}
	defer expertiseRows.Close()
	for expertiseRows.Next() {
		var item TechnicalExpertiseSnapshot
		if err := expertiseRows.Scan(&item.Skill, &item.Category, &item.ProficiencyLevel); err != nil {
			return nil, err
		}
		c.TechnicalExpertise = append(c.TechnicalExpertise, item)
	}
	if err := expertiseRows.Err(); err != nil {
		return nil, err
	}

	return c, nil
}

func fetchRequirementSnapshotTx(tx *sql.Tx, requirementID int, tenantID string) (*RequirementSnapshotData, error) {
	r := &RequirementSnapshotData{}
	err := tx.QueryRow(`
		SELECT id, title, COALESCE(location, ''), COALESCE(work_arrangement, ''),
			COALESCE(description, ''), COALESCE(required_skills, ''),
			COALESCE(experience_required, ''), COALESCE(compensation, ''),
			headcount, COALESCE(language_requirement, '')
		FROM requirements WHERE id = $1 AND tenant_id = $2`, requirementID, tenantID).Scan(
		&r.ID, &r.Title, &r.Location, &r.WorkArrangement, &r.Description,
		&r.RequiredSkills, &r.ExperienceRequired, &r.Compensation, &r.Headcount,
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
