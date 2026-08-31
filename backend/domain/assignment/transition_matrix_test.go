package assignment

import "testing"

func TestAllowedTransitions_ReturnsDefensiveCopy(t *testing.T) {
	got := AllowedTransitions(StatusScreening)
	if len(got) != 3 {
		t.Fatalf("len(AllowedTransitions(screening)) = %d, want 3", len(got))
	}

	got[0] = StatusJoined
	if CanTransition(StatusScreening, StatusJoined) {
		t.Fatal("mutating returned transitions changed the authoritative workflow matrix")
	}
	if !CanTransition(StatusScreening, StatusSubmitted) {
		t.Fatal("mutating returned transitions removed a valid transition")
	}
}

func TestCanTransition_UnknownStatesAreRejected(t *testing.T) {
	cases := []struct {
		name string
		from Status
		to   Status
	}{
		{"unknown source", Status("unknown"), StatusScreening},
		{"unknown target", StatusDraft, Status("unknown")},
		{"both unknown", Status("unknown"), Status("other")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if CanTransition(tc.from, tc.to) {
				t.Fatalf("CanTransition(%q, %q) = true, want false", tc.from, tc.to)
			}
		})
	}
}

func TestAllowedTransitions_TerminalStatesAreEmpty(t *testing.T) {
	for _, status := range []Status{StatusJoined, StatusRejected, StatusWithdrawn} {
		if got := AllowedTransitions(status); len(got) != 0 {
			t.Errorf("AllowedTransitions(%q) = %v, want empty", status, got)
		}
	}
}
