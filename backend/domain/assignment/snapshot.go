package assignment

import (
	"database/sql"
	"errors"
)

// CandidateSnapshotData is the immutable candidate-facing content captured
// at formal submission (ADR 0003 section 6): "the relevant candidate
// identity/profile, skills/language information". This mirrors the
// relevant subset of the candidates table at the moment of capture — it is
// NOT a live reference to the candidate row, and nothing here changes if
// the underlying candidate record is later edited.
type CandidateSnapshotData struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone,omitempty"`
	Position     string `json:"position,omitempty"`
	Location     string `json:"location,omitempty"`
	Experience   string `json:"experience,omitempty"`
	CurrentCTC   string `json:"currentCtc,omitempty"`
	ExpectedCTC  string `json:"expectedCtc,omitempty"`
	NoticePeriod string `json:"noticePeriod,omitempty"`
	JLPTLanguage string `json:"jlptLanguage,omitempty"`
	Skills       string `json:"skills,omitempty"`
	Status       string `json:"status"`
}

// RequirementSnapshotData is the immutable requirement-facing content
// captured at formal submission (ADR 0003 section 6): "title, location,
// work arrangement, description, required skills, experience,
// compensation, headcount, and language requirement" — the exact field
// list the ADR specifies as the minimum.
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
// assignment cannot be found at the moment of snapshot capture. This
// should not be reachable in practice (candidate_id/requirement_id are
// immutable on an assignment after creation, and creation already
// validates both exist in-tenant — see Service.CreateAssignment), but is
// handled explicitly rather than assumed, consistent with this codebase's
// existing defensive-verification convention.
var (
	ErrCandidateSnapshotSourceNotFound   = errors.New("candidate record for snapshot not found")
	ErrRequirementSnapshotSourceNotFound = errors.New("requirement record for snapshot not found")
)

// fetchCandidateSnapshotTx reads the current candidate row within tx, for
// embedding into an immutable submission snapshot. COALESCEs nullable
// columns to ” to avoid the nil-scan class of bug found in earlier
// checkpoints (#33/#34's nullable-scan issue).
func fetchCandidateSnapshotTx(tx *sql.Tx, candidateID int, tenantID string) (*CandidateSnapshotData, error) {
	c := &CandidateSnapshotData{}
	err := tx.QueryRow(`
		SELECT id, name, email, COALESCE(phone, ''), COALESCE(position, ''), COALESCE(location, ''),
			COALESCE(experience, ''), COALESCE(currentctc, ''), COALESCE(expectedctc, ''),
			COALESCE(noticeperiod, ''), COALESCE(jlptlanguage, ''), COALESCE(skills, ''), status
		FROM candidates WHERE id = $1 AND tenant_id = $2`,
		candidateID, tenantID,
	).Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, &c.Location, &c.Experience,
		&c.CurrentCTC, &c.ExpectedCTC, &c.NoticePeriod, &c.JLPTLanguage, &c.Skills, &c.Status)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCandidateSnapshotSourceNotFound
		}
		return nil, err
	}
	return c, nil
}

// fetchRequirementSnapshotTx reads the current requirement row within tx,
// for embedding into an immutable submission snapshot.
func fetchRequirementSnapshotTx(tx *sql.Tx, requirementID int, tenantID string) (*RequirementSnapshotData, error) {
	r := &RequirementSnapshotData{}
	err := tx.QueryRow(`
		SELECT id, title, COALESCE(location, ''), COALESCE(work_arrangement, ''), COALESCE(description, ''),
			COALESCE(required_skills, ''), COALESCE(experience_required, ''), COALESCE(compensation, ''),
			headcount, COALESCE(language_requirement, '')
		FROM requirements WHERE id = $1 AND tenant_id = $2`,
		requirementID, tenantID,
	).Scan(&r.ID, &r.Title, &r.Location, &r.WorkArrangement, &r.Description,
		&r.RequiredSkills, &r.ExperienceRequired, &r.Compensation, &r.Headcount, &r.LanguageRequirement)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRequirementSnapshotSourceNotFound
		}
		return nil, err
	}
	return r, nil
}
