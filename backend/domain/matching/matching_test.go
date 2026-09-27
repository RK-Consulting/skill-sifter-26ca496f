package matching

import "testing"

func TestMatchSkills(t *testing.T) {
	result := MatchSkills("Go, PostgreSQL", []string{"Go", "PostgreSQL", "Docker"})
	if result.Status != Matched {
		t.Fatalf("status = %s, want %s", result.Status, Matched)
	}
}

func TestMatchSkillsMissing(t *testing.T) {
	result := MatchSkills("Go, Rust", []string{"Go"})
	if result.Status != Missing {
		t.Fatalf("status = %s, want %s", result.Status, Missing)
	}
}

func TestMatchExperience(t *testing.T) {
	result := MatchExperience("5+ years", "7 years")
	if result.Status != Matched {
		t.Fatalf("status = %s, want %s", result.Status, Matched)
	}
}

func TestMatchExperienceUnknown(t *testing.T) {
	result := MatchExperience("5+ years", "")
	if result.Status != Unknown {
		t.Fatalf("status = %s, want %s", result.Status, Unknown)
	}
}

func TestMatchLanguage(t *testing.T) {
	result := MatchLanguages("English, Japanese", []LanguageEvidence{{Language: "English"}, {Language: "Japanese", Proficiency: "N2"}})
	if result.Status != Matched {
		t.Fatalf("status = %s, want %s", result.Status, Matched)
	}
}

func TestMatchCertificationMissing(t *testing.T) {
	result := MatchCertifications("AWS Solutions Architect", []string{"PMP"})
	if result.Status != Missing {
		t.Fatalf("status = %s, want %s", result.Status, Missing)
	}
}

func TestMatchLocationUnknown(t *testing.T) {
	result := MatchText("location", "Bangalore", "")
	if result.Status != Unknown {
		t.Fatalf("status = %s, want %s", result.Status, Unknown)
	}
}

func TestWorkArrangementUnknown(t *testing.T) {
	result := MatchWorkArrangement("remote", "")
	if result.Status != Unknown {
		t.Fatalf("status = %s, want %s", result.Status, Unknown)
	}
}

func TestEvaluateMissingTakesPrecedence(t *testing.T) {
	result := Evaluate(RequirementEvidence{
		RequiredSkills: "Go",
		ExperienceRequired: "5+ years",
	}, CandidateEvidence{
		TechnicalSkills: []string{"Python"},
		Experience: "7 years",
	})
	if result.Status != Missing {
		t.Fatalf("status = %s, want %s", result.Status, Missing)
	}
}
