package handlers

import (
	"database/sql"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

type resumeEmployment struct {
	Employer    string `json:"employer"`
	JobTitle    string `json:"jobTitle"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	StartYear   int    `json:"startYear"`
	EndYear     int    `json:"endYear"`
	IsCurrent   bool   `json:"isCurrent"`
	Description string `json:"description"`
}

type resumeEducation struct {
	Institution  string `json:"institution"`
	Degree       string `json:"degree"`
	FieldOfStudy string `json:"fieldOfStudy"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	StartYear    int    `json:"startYear"`
	EndYear      int    `json:"endYear"`
	Description  string `json:"description"`
}

type resumeCertification struct {
	Name                string `json:"name"`
	Issuer              string `json:"issuer"`
	IssueDate           string `json:"issueDate"`
	ExpiryDate          string `json:"expiryDate"`
	IssueYear           int    `json:"issueYear"`
	ExpiryYear          int    `json:"expiryYear"`
	CredentialReference string `json:"credentialReference"`
}

type resumeProject struct {
	ProjectName  string   `json:"projectName"`
	Description  string   `json:"description"`
	Role         string   `json:"role"`
	Technologies []string `json:"technologies"`
	StartDate    string   `json:"startDate"`
	EndDate      string   `json:"endDate"`
	StartYear    int      `json:"startYear"`
	EndYear      int      `json:"endYear"`
}

type resumeLanguage struct {
	Name                 string `json:"name"`
	ProficiencyFramework string `json:"proficiencyFramework"`
	ProficiencyLevel     string `json:"proficiencyLevel"`
}

var technicalSkillAliases = map[string]string{
	"golang": "Go", "go programming": "Go",
	"postgres": "PostgreSQL", "postgresql": "PostgreSQL", "postgres sql": "PostgreSQL",
	"javascript": "JavaScript", "java script": "JavaScript",
	"typescript": "TypeScript", "react.js": "React", "reactjs": "React",
	"node": "Node.js", "nodejs": "Node.js", "node.js": "Node.js",
	"k8s": "Kubernetes", "kubernetes": "Kubernetes",
	"amazon web services": "AWS", "aws": "AWS",
	"google cloud platform": "GCP", "gcp": "GCP",
	"microsoft azure": "Azure", "azure": "Azure",
	"c plus plus": "C++", "c++": "C++", "c sharp": "C#", "c#": "C#",
}

var humanLanguageNames = map[string]bool{
	"english": true, "hindi": true, "kannada": true, "tamil": true,
	"telugu": true, "malayalam": true, "marathi": true, "bengali": true,
	"gujarati": true, "punjabi": true, "urdu": true, "french": true,
	"german": true, "spanish": true, "italian": true, "portuguese": true,
	"japanese": true, "korean": true, "mandarin": true, "chinese": true,
	"arabic": true, "russian": true,
}

func normalizeTechnicalSkills(skills []string) []string {
	out := make([]string, 0, len(skills))
	seen := make(map[string]bool)
	for _, raw := range skills {
		skill := strings.TrimSpace(raw)
		if skill == "" {
			continue
		}
		key := strings.ToLower(strings.Join(strings.Fields(skill), " "))
		if humanLanguageNames[key] {
			continue
		}
		if canonical, ok := technicalSkillAliases[key]; ok {
			skill = canonical
		} else {
			skill = strings.Join(strings.Fields(skill), " ")
			if len(skill) > 0 {
				skill = strings.ToUpper(skill[:1]) + skill[1:]
			}
		}
		key = strings.ToLower(skill)
		if !seen[key] {
			seen[key] = true
			out = append(out, skill)
		}
	}
	return out
}

func normalizeResumeAI(ai *resumeAIResult) error {
	ai.Name = strings.TrimSpace(ai.Name)
	ai.Email = strings.TrimSpace(ai.Email)
	ai.Phone = strings.TrimSpace(ai.Phone)
	ai.Location = strings.TrimSpace(ai.Location)
	ai.CurrentTitle = strings.TrimSpace(ai.CurrentTitle)
	ai.ProfessionalSummary = strings.TrimSpace(ai.ProfessionalSummary)
	ai.TotalExperience = strings.TrimSpace(ai.TotalExperience)
	ai.RelevantExperience = strings.TrimSpace(ai.RelevantExperience)
	ai.Skills = normalizeTechnicalSkills(ai.Skills)

	if ai.Email != "" {
		if _, err := mail.ParseAddress(ai.Email); err != nil {
			return fmt.Errorf("invalid extracted email")
		}
	}
	if err := validateResumeYears(ai.EmploymentHistory); err != nil {
		return err
	}
	if err := validateEducationYears(ai.Education); err != nil {
		return err
	}
	if err := validateCertificationYears(ai.Certifications); err != nil {
		return err
	}
	if err := validateProjectYears(ai.Projects); err != nil {
		return err
	}

	for i := range ai.Languages {
		ai.Languages[i].Name = strings.TrimSpace(ai.Languages[i].Name)
		ai.Languages[i].ProficiencyFramework = strings.TrimSpace(ai.Languages[i].ProficiencyFramework)
		ai.Languages[i].ProficiencyLevel = strings.TrimSpace(ai.Languages[i].ProficiencyLevel)
		if ai.Languages[i].ProficiencyFramework == "" {
			ai.Languages[i].ProficiencyFramework = "unspecified"
		}
		if ai.Languages[i].ProficiencyLevel == "" {
			ai.Languages[i].ProficiencyLevel = "unspecified"
		}
	}
	return nil
}

func validateResumeYears(items []resumeEmployment) error {
	for _, item := range items {
		if item.StartYear != 0 && (item.StartYear < 1900 || item.StartYear > 2200) {
			return fmt.Errorf("invalid employment startYear %d", item.StartYear)
		}
		if item.EndYear != 0 && (item.EndYear < 1900 || item.EndYear > 2200) {
			return fmt.Errorf("invalid employment endYear %d", item.EndYear)
		}
		if item.StartYear != 0 && item.EndYear != 0 && item.EndYear < item.StartYear {
			return fmt.Errorf("employment endYear precedes startYear")
		}
	}
	return nil
}

func validateEducationYears(items []resumeEducation) error {
	for _, item := range items {
		if item.StartYear != 0 && (item.StartYear < 1900 || item.StartYear > 2200) {
			return fmt.Errorf("invalid education startYear %d", item.StartYear)
		}
		if item.EndYear != 0 && (item.EndYear < 1900 || item.EndYear > 2200) {
			return fmt.Errorf("invalid education endYear %d", item.EndYear)
		}
		if item.StartYear != 0 && item.EndYear != 0 && item.EndYear < item.StartYear {
			return fmt.Errorf("education endYear precedes startYear")
		}
	}
	return nil
}

func validateCertificationYears(items []resumeCertification) error {
	for _, item := range items {
		if item.IssueYear != 0 && (item.IssueYear < 1900 || item.IssueYear > 2200) {
			return fmt.Errorf("invalid certification issueYear %d", item.IssueYear)
		}
		if item.ExpiryYear != 0 && (item.ExpiryYear < 1900 || item.ExpiryYear > 2200) {
			return fmt.Errorf("invalid certification expiryYear %d", item.ExpiryYear)
		}
		if item.IssueYear != 0 && item.ExpiryYear != 0 && item.ExpiryYear < item.IssueYear {
			return fmt.Errorf("certification expiryYear precedes issueYear")
		}
	}
	return nil
}

func validateProjectYears(items []resumeProject) error {
	for _, item := range items {
		if item.StartYear != 0 && (item.StartYear < 1900 || item.StartYear > 2200) {
			return fmt.Errorf("invalid project startYear %d", item.StartYear)
		}
		if item.EndYear != 0 && (item.EndYear < 1900 || item.EndYear > 2200) {
			return fmt.Errorf("invalid project endYear %d", item.EndYear)
		}
		if item.StartYear != 0 && item.EndYear != 0 && item.EndYear < item.StartYear {
			return fmt.Errorf("project endYear precedes startYear")
		}
	}
	return nil
}

func resumeDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	// Partial dates are retained through the corresponding year field.
	// Never invent a day such as the first of the month/year.
	if _, err := time.Parse("2006-01", value); err == nil {
		return nil, nil
	}
	if _, err := time.Parse("2006", value); err == nil {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q", value)
	}
	return &parsed, nil
}

func resumeYear(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func resumeYearFromPartialDate(value string, explicitYear int) *int {
	if explicitYear != 0 {
		return resumeYear(explicitYear)
	}

	value = strings.TrimSpace(value)
	if len(value) >= 4 {
		if year, err := strconv.Atoi(value[:4]); err == nil && year >= 1900 && year <= 2200 {
			return &year
		}
	}
	return nil
}

func persistResumeIntelligence(database *sql.DB, resumeID, candidateID int, tenantID string, ai resumeAIResult) error {
	if err := normalizeResumeAI(&ai); err != nil {
		return err
	}
	if candidateID <= 0 || tenantID == "" {
		return fmt.Errorf("candidate and tenant are required for resume intelligence")
	}

	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, query := range []string{
		"DELETE FROM candidate_employment_history WHERE tenant_id = $1 AND candidate_id = $2 AND source_resume_id = $3",
		"DELETE FROM candidate_education WHERE tenant_id = $1 AND candidate_id = $2 AND source_resume_id = $3",
		"DELETE FROM candidate_certifications WHERE tenant_id = $1 AND candidate_id = $2 AND source_resume_id = $3",
		"DELETE FROM candidate_projects WHERE tenant_id = $1 AND candidate_id = $2 AND source_resume_id = $3",
	} {
		if _, err := tx.Exec(query, tenantID, candidateID, resumeID); err != nil {
			return err
		}
	}

	_, err = tx.Exec(`
		INSERT INTO candidate_professional_profiles (
			tenant_id, candidate_id, source_resume_id, current_title,
			professional_summary, location, total_experience, relevant_experience
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tenant_id, candidate_id)
		DO UPDATE SET
			source_resume_id = EXCLUDED.source_resume_id,
			current_title = COALESCE(NULLIF(EXCLUDED.current_title, ''), candidate_professional_profiles.current_title),
			professional_summary = COALESCE(NULLIF(EXCLUDED.professional_summary, ''), candidate_professional_profiles.professional_summary),
			location = COALESCE(NULLIF(EXCLUDED.location, ''), candidate_professional_profiles.location),
			total_experience = COALESCE(NULLIF(EXCLUDED.total_experience, ''), candidate_professional_profiles.total_experience),
			relevant_experience = COALESCE(NULLIF(EXCLUDED.relevant_experience, ''), candidate_professional_profiles.relevant_experience),
			updated_at = NOW()
	`, tenantID, candidateID, resumeID, ai.CurrentTitle, ai.ProfessionalSummary, ai.Location, ai.TotalExperience, ai.RelevantExperience)
	if err != nil {
		return err
	}

	for _, skill := range ai.Skills {
		if _, err := tx.Exec(`
			INSERT INTO candidate_expertise (tenant_id, candidate_id, skill, category, proficiency_level)
			VALUES ($1,$2,$3,'resume_import','unspecified')
			ON CONFLICT (candidate_id, skill, category)
			DO UPDATE SET proficiency_level = EXCLUDED.proficiency_level, updated_at = NOW()
		`, tenantID, candidateID, skill); err != nil {
			return err
		}
	}

	for _, language := range ai.Languages {
		if language.Name == "" {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO candidate_language_expertise (
				tenant_id, candidate_id, language, proficiency_framework, proficiency_level, source_resume_id
			) VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (candidate_id, language, proficiency_framework, proficiency_level)
			DO UPDATE SET source_resume_id = EXCLUDED.source_resume_id, updated_at = NOW()
		`, tenantID, candidateID, language.Name, language.ProficiencyFramework, language.ProficiencyLevel, resumeID); err != nil {
			return err
		}
	}

	for order, item := range ai.EmploymentHistory {
		if strings.TrimSpace(item.Employer) == "" {
			continue
		}
		startDate, err := resumeDate(item.StartDate)
		if err != nil {
			return err
		}
		endDate, err := resumeDate(item.EndDate)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO candidate_employment_history (
				tenant_id,candidate_id,source_resume_id,employer,job_title,start_date,end_date,
				start_year,end_year,is_current,description,sort_order
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, tenantID, candidateID, resumeID, strings.TrimSpace(item.Employer), strings.TrimSpace(item.JobTitle),
			startDate, endDate, resumeYearFromPartialDate(item.StartDate, item.StartYear), resumeYearFromPartialDate(item.EndDate, item.EndYear), item.IsCurrent,
			strings.TrimSpace(item.Description), order); err != nil {
			return err
		}
	}

	for order, item := range ai.Education {
		if strings.TrimSpace(item.Institution) == "" {
			continue
		}
		startDate, err := resumeDate(item.StartDate)
		if err != nil {
			return err
		}
		endDate, err := resumeDate(item.EndDate)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO candidate_education (
				tenant_id,candidate_id,source_resume_id,institution,degree,field_of_study,
				start_date,end_date,start_year,end_year,description,sort_order
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, tenantID, candidateID, resumeID, strings.TrimSpace(item.Institution), strings.TrimSpace(item.Degree),
			strings.TrimSpace(item.FieldOfStudy), startDate, endDate, resumeYearFromPartialDate(item.StartDate, item.StartYear),
			resumeYearFromPartialDate(item.EndDate, item.EndYear), strings.TrimSpace(item.Description), order); err != nil {
			return err
		}
	}

	for _, item := range ai.Certifications {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		issueDate, err := resumeDate(item.IssueDate)
		if err != nil {
			return err
		}
		expiryDate, err := resumeDate(item.ExpiryDate)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO candidate_certifications (
				tenant_id,candidate_id,source_resume_id,name,issuer,issue_date,expiry_date,
				issue_year,expiry_year,credential_reference
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`, tenantID, candidateID, resumeID, strings.TrimSpace(item.Name), strings.TrimSpace(item.Issuer),
			issueDate, expiryDate, resumeYearFromPartialDate(item.IssueDate, item.IssueYear), resumeYearFromPartialDate(item.ExpiryDate, item.ExpiryYear),
			strings.TrimSpace(item.CredentialReference)); err != nil {
			return err
		}
	}

	for order, item := range ai.Projects {
		if strings.TrimSpace(item.ProjectName) == "" {
			continue
		}
		startDate, err := resumeDate(item.StartDate)
		if err != nil {
			return err
		}
		endDate, err := resumeDate(item.EndDate)
		if err != nil {
			return err
		}
		technologies := normalizeTechnicalSkills(item.Technologies)
		if _, err := tx.Exec(`
			INSERT INTO candidate_projects (
				tenant_id,candidate_id,source_resume_id,project_name,description,role,
				technologies,start_date,end_date,start_year,end_year,sort_order
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, tenantID, candidateID, resumeID, strings.TrimSpace(item.ProjectName),
			strings.TrimSpace(item.Description), strings.TrimSpace(item.Role), pq.Array(technologies),
			startDate, endDate, resumeYearFromPartialDate(item.StartDate, item.StartYear), resumeYearFromPartialDate(item.EndDate, item.EndYear), order); err != nil {
			return err
		}
	}

	return tx.Commit()
}
