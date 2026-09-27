package matching

import (
	"regexp"
	"strconv"
	"strings"
)

type Status string

const (
	Matched       Status = "matched"
	Missing       Status = "missing"
	Unknown       Status = "unknown"
	NotApplicable Status = "not_applicable"
)

type CriterionEvidence struct {
	Criterion string   `json:"criterion"`
	Required  string   `json:"required,omitempty"`
	Evidence  string   `json:"evidence,omitempty"`
	Status    Status   `json:"status"`
	Details   []string `json:"details,omitempty"`
}

type CandidateEvidence struct {
	TechnicalSkills  []string
	Languages        []LanguageEvidence
	Certifications   []string
	Experience       string
	Location         string
	NoticePeriod     string
	WorkArrangement  string
}

type LanguageEvidence struct {
	Language    string
	Proficiency string
}

type RequirementEvidence struct {
	RequiredSkills         string
	MandatoryRequirements  string
	ExperienceRequired     string
	LanguageRequirements   string
	CertificationsRequired string
	Location             string
	NoticePeriod         string
	WorkArrangement      string
}

type MatchResult struct {
	Status   Status              `json:"status"`
	Criteria []CriterionEvidence `json:"criteria"`
	Summary  string              `json:"summary"`
}

var nonWord = regexp.MustCompile(`[^a-z0-9+#.]+`)
var yearsPattern = regexp.MustCompile(`(?i)([0-9]+(?:\\.[0-9]+)?)\\s*\\+?\\s*(?:years?|yrs?)`)

func Normalize(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "c sharp", "c#")
	value = strings.ReplaceAll(value, "c plus plus", "c++")
	value = nonWord.ReplaceAllString(value, " ")
	return strings.Join(strings.Fields(value), " ")
}

func SplitRequirements(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '|'
	})
	result := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n := Normalize(part)
		if n != "" && !seen[n] {
			seen[n] = true
			result = append(result, part)
		}
	}
	return result
}

func containsNormalized(values []string, required string) (bool, string) {
	r := Normalize(required)
	if r == "" {
		return false, ""
	}
	for _, value := range values {
		n := Normalize(value)
		if n == r || strings.Contains(n, r) || strings.Contains(r, n) {
			return true, value
		}
	}
	return false, ""
}

func MatchSkills(required string, skills []string) CriterionEvidence {
	if strings.TrimSpace(required) == "" {
		return CriterionEvidence{Criterion: "skills", Status: NotApplicable}
	}
	requiredItems := SplitRequirements(required)
	if len(requiredItems) == 0 {
		return CriterionEvidence{Criterion: "skills", Status: Unknown, Required: required}
	}
	missing := []string{}
	matched := []string{}
	for _, item := range requiredItems {
		if ok, evidence := containsNormalized(skills, item); ok {
			matched = append(matched, item+" -> "+evidence)
		} else {
			missing = append(missing, item)
		}
	}
	status := Matched
	if len(missing) > 0 {
		status = Missing
	}
	return CriterionEvidence{Criterion: "skills", Required: required, Status: status, Details: append([]string{"matched: "+strings.Join(matched, ", ")}, "missing: "+strings.Join(missing, ", "))}
}

func MatchLanguages(required string, languages []LanguageEvidence) CriterionEvidence {
	if strings.TrimSpace(required) == "" {
		return CriterionEvidence{Criterion: "languages", Status: NotApplicable}
	}
	items := SplitRequirements(required)
	if len(items) == 0 {
		return CriterionEvidence{Criterion: "languages", Status: Unknown, Required: required}
	}
	values := make([]string, 0, len(languages))
	for _, l := range languages {
		values = append(values, l.Language+" "+l.Proficiency)
	}
	missing := []string{}
	matched := []string{}
	for _, item := range items {
		if ok, evidence := containsNormalized(values, item); ok {
			matched = append(matched, item+" -> "+evidence)
		} else {
			missing = append(missing, item)
		}
	}
	status := Matched
	if len(missing) > 0 {
		status = Missing
	}
	return CriterionEvidence{Criterion: "languages", Required: required, Status: status, Details: []string{"matched: "+strings.Join(matched, ", "), "missing: "+strings.Join(missing, ", ")}}
}

func MatchCertifications(required string, certifications []string) CriterionEvidence {
	if strings.TrimSpace(required) == "" {
		return CriterionEvidence{Criterion: "certifications", Status: NotApplicable}
	}
	items := SplitRequirements(required)
	if len(items) == 0 {
		return CriterionEvidence{Criterion: "certifications", Status: Unknown, Required: required}
	}
	missing := []string{}
	matched := []string{}
	for _, item := range items {
		if ok, evidence := containsNormalized(certifications, item); ok {
			matched = append(matched, item+" -> "+evidence)
		} else {
			missing = append(missing, item)
		}
	}
	status := Matched
	if len(missing) > 0 {
		status = Missing
	}
	return CriterionEvidence{Criterion: "certifications", Required: required, Status: status, Details: []string{"matched: "+strings.Join(matched, ", "), "missing: "+strings.Join(missing, ", ")}}
}

func parseYears(value string) (float64, bool) {
	matches := yearsPattern.FindStringSubmatch(strings.ToLower(value))
	if len(matches) > 1 {
		n, err := strconv.ParseFloat(matches[1], 64)
		return n, err == nil
	}
	return 0, false
}

func MatchExperience(required, candidate string) CriterionEvidence {
	if strings.TrimSpace(required) == "" {
		return CriterionEvidence{Criterion: "experience", Status: NotApplicable}
	}
	requiredYears, requiredKnown := parseYears(required)
	candidateYears, candidateKnown := parseYears(candidate)
	if !requiredKnown || !candidateKnown {
		return CriterionEvidence{Criterion: "experience", Required: required, Evidence: candidate, Status: Unknown, Details: []string{"Experience could not be deterministically parsed."}}
	}
	status := Matched
	if candidateYears < requiredYears {
		status = Missing
	}
	return CriterionEvidence{Criterion: "experience", Required: required, Evidence: candidate, Status: status, Details: []string{"required years: "+strconv.FormatFloat(requiredYears, 'f', -1, 64), "candidate years: "+strconv.FormatFloat(candidateYears, 'f', -1, 64)}}
}

func MatchText(criterion, required, candidate string) CriterionEvidence {
	if strings.TrimSpace(required) == "" {
		return CriterionEvidence{Criterion: criterion, Status: NotApplicable}
	}
	if strings.TrimSpace(candidate) == "" {
		return CriterionEvidence{Criterion: criterion, Required: required, Status: Unknown}
	}
	r := Normalize(required)
	c := Normalize(candidate)
	status := Missing
	if c == r || strings.Contains(c, r) || strings.Contains(r, c) {
		status = Matched
	}
	return CriterionEvidence{Criterion: criterion, Required: required, Evidence: candidate, Status: status}
}

func MatchWorkArrangement(required, candidate string) CriterionEvidence {
	if strings.TrimSpace(required) == "" {
		return CriterionEvidence{Criterion: "workArrangement", Status: NotApplicable}
	}
	if strings.TrimSpace(candidate) == "" {
		return CriterionEvidence{Criterion: "workArrangement", Required: required, Status: Unknown, Details: []string{"Candidate work arrangement is not currently stored in candidate intelligence."}}
	}
	return MatchText("workArrangement", required, candidate)
}

func Evaluate(req RequirementEvidence, candidate CandidateEvidence) MatchResult {
	criteria := []CriterionEvidence{
		MatchSkills(req.RequiredSkills, candidate.TechnicalSkills),
		MatchSkills(req.MandatoryRequirements, candidate.TechnicalSkills),
		MatchExperience(req.ExperienceRequired, candidate.Experience),
		MatchLanguages(req.LanguageRequirements, candidate.Languages),
		MatchCertifications(req.CertificationsRequired, candidate.Certifications),
		MatchText("location", req.Location, candidate.Location),
		MatchText("noticePeriod", req.NoticePeriod, candidate.NoticePeriod),
		MatchWorkArrangement(req.WorkArrangement, candidate.WorkArrangement),
	}
	status := Matched
	for _, criterion := range criteria {
		switch criterion.Status {
		case Missing:
			status = Missing
		case Unknown:
			if status != Missing {
				status = Unknown
			}
		}
	}
	return MatchResult{
		Status: status,
		Criteria: criteria,
		Summary: summary(status),
	}
}

func summary(status Status) string {
	switch status {
	case Matched:
		return "All applicable criteria have deterministic evidence."
	case Missing:
		return "At least one required criterion is missing from the candidate evidence."
	case Unknown:
		return "No required criterion is known to be missing, but at least one criterion cannot be determined from available evidence."
	default:
		return "No applicable matching criteria are defined."
	}
}
