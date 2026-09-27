package handlers

import "testing"

func TestValidateInterviewStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{status: "scheduled", want: true},
		{status: "completed", want: true},
		{status: "cancelled", want: true},
		{status: "rescheduled", want: true},
		{status: "no_show", want: true},
		{status: "", want: false},
		{status: "Scheduled", want: false},
		{status: "selected", want: false},
		{status: "rejected", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := validateInterviewStatus(tt.status); got != tt.want {
				t.Fatalf("validateInterviewStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
