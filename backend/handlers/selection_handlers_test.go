package handlers

import "testing"

func TestValidateSelectionDecision(t *testing.T) {
	tests := []struct {
		decision string
		want     bool
	}{
		{decision: "selected", want: true},
		{decision: "rejected", want: true},
		{decision: "", want: false},
		{decision: "Selected", want: false},
		{decision: "offered", want: false},
		{decision: "withdrawn", want: false},
	}
	for _, tt := range tests {
		if got := validateSelectionDecision(tt.decision); got != tt.want {
			t.Fatalf("validateSelectionDecision(%q) = %v, want %v", tt.decision, got, tt.want)
		}
	}
}
