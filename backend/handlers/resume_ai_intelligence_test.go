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
