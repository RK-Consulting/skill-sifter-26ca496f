package handlers

import (
	"strings"
	"testing"
)

func TestNormalizeTechnicalSkills(t *testing.T) {
	got := normalizeTechnicalSkills([]string{
		"golang", "Postgres", "postgresql", "React.js", "English", " Docker ", "",
	})
	want := []string{"Go", "PostgreSQL", "React", "Docker"}

	if len(got) != len(want) {
		t.Fatalf("normalized skills = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("normalized skills = %#v, want %#v", got, want)
		}
	}
}

func TestNormalizeResumeAIRejectsInvalidStructuredData(t *testing.T) {
	ai := resumeAIResult{
		Email: "not-an-email",
		EmploymentHistory: []resumeEmployment{{StartYear: 2025, EndYear: 2020}},
	}
	if err := normalizeResumeAI(&ai); err == nil {
		t.Fatal("expected invalid email/year data to be rejected")
	}
}

func TestNormalizeResumeAIAllowsMissingOptionalFields(t *testing.T) {
	ai := resumeAIResult{Name: "Ada Lovelace", Skills: []string{"golang", "English"}}
	if err := normalizeResumeAI(&ai); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if !strings.Contains(strings.Join(ai.Skills, ","), "Go") {
		t.Fatalf("normalized skills = %#v, want Go", ai.Skills)
	}
	if len(ai.Languages) != 0 {
		t.Fatalf("languages = %#v, want empty", ai.Languages)
	}
}

func TestNormalizeResumeAIRejectsInvalidDateDuringPersistenceValidation(t *testing.T) {
	ai := resumeAIResult{
		EmploymentHistory: []resumeEmployment{{
			Employer:  "Example Corp",
			StartDate: "not-a-date",
		}},
	}
	if err := normalizeResumeAI(&ai); err != nil {
		t.Fatalf("year validation should not reject a date-format error before persistence: %v", err)
	}
	if _, err := resumeDate("not-a-date"); err == nil {
		t.Fatal("expected invalid date to be rejected")
	}
}

func TestCallOllamaStructuredIntelligence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"response":"{\"name\":\"Ada Lovelace\",\"email\":\"ada@example.com\",\"phone\":\"123\",\"location\":\"London\",\"currentTitle\":\"Software Engineer\",\"professionalSummary\":\"Analytical engineer\",\"totalExperience\":\"7 years\",\"relevantExperience\":\"5 years\",\"skills\":[\"golang\",\"Postgres\"],\"languages\":[{\"name\":\"English\",\"proficiencyFramework\":\"CEFR\",\"proficiencyLevel\":\"C1\"}],\"employmentHistory\":[{\"employer\":\"Example Corp\",\"jobTitle\":\"Engineer\",\"startYear\":2020,\"isCurrent\":true}],\"education\":[{\"institution\":\"Example University\",\"degree\":\"BSc\",\"fieldOfStudy\":\"Mathematics\"}],\"certifications\":[{\"name\":\"AWS Certified Developer\",\"issuer\":\"AWS\",\"issueYear\":2024}],\"projects\":[{\"projectName\":\"Compiler\",\"technologies\":[\"golang\"]}]}"}`)
	}))
	defer server.Close()
	t.Setenv("OLLAMA_URL", server.URL)

	got, errText := callOllama("Ada resume")
	if errText != "" {
		t.Fatalf("structured Ollama parse error = %q", errText)
	}
	if got.Name != "Ada Lovelace" || got.Location != "London" || len(got.EmploymentHistory) != 1 ||
		len(got.Education) != 1 || len(got.Certifications) != 1 || len(got.Projects) != 1 ||
		len(got.Languages) != 1 {
		t.Fatalf("structured result = %+v", got)
	}
	if err := normalizeResumeAI(&got); err != nil {
		t.Fatalf("structured result validation failed: %v", err)
	}
	if got.Skills[0] != "Go" || got.Skills[1] != "PostgreSQL" {
		t.Fatalf("normalized skills = %#v", got.Skills)
	}
}
