package feedback

import "testing"

func TestOutcomeValid(t *testing.T) {
	tests := []struct {
		name    string
		outcome Outcome
		valid   bool
	}{
		{"shortlist", OutcomeShortlist, true},
		{"hold", OutcomeHold, true},
		{"reject", OutcomeReject, true},
		{"invalid", Outcome("unknown"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.outcome.Valid(); got != tt.valid {
				t.Fatalf("Outcome.Valid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestReasonCodeValid(t *testing.T) {
	tests := []struct {
		code  ReasonCode
		valid bool
	}{
		{ReasonCodeSkillsGap, true},
		{ReasonCodeExperienceGap, true},
		{ReasonCodeCompensationMismatch, true},
		{ReasonCodeLocationMismatch, true},
		{ReasonCodeNoticePeriod, true},
		{ReasonCodeCandidateNotInterested, true},
		{ReasonCodeAvailability, true},
		{ReasonCodeProfileMismatch, true},
		{ReasonCodeOther, true},
		{ReasonCode("custom"), false},
	}
	for _, tt := range tests {
		if got := tt.code.Valid(); got != tt.valid {
			t.Errorf("ReasonCode(%q).Valid() = %v, want %v", tt.code, got, tt.valid)
		}
	}
}
