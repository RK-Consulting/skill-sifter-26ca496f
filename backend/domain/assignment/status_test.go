package assignment

import "testing"

func TestCanTransition_ValidPaths(t *testing.T) {
	valid := []struct{ from, to Status }{
		{StatusDraft, StatusScreening},
		{StatusScreening, StatusSubmitted},
		{StatusSubmitted, StatusInterviewing},
		{StatusInterviewing, StatusOffered},
		{StatusOffered, StatusJoined},
		{StatusScreening, StatusRejected},
		{StatusSubmitted, StatusRejected},
		{StatusInterviewing, StatusRejected},
		{StatusOffered, StatusRejected},
		{StatusScreening, StatusWithdrawn},
		{StatusSubmitted, StatusWithdrawn},
		{StatusInterviewing, StatusWithdrawn},
		{StatusOffered, StatusWithdrawn},
	}
	for _, tc := range valid {
		if !CanTransition(tc.from, tc.to) {
			t.Errorf("CanTransition(%q, %q) = false, want true", tc.from, tc.to)
		}
	}
}

// TestCanTransition_ExhaustiveMatrix checks every possible (from, to) pair
// across all 8 statuses (64 combinations) against the exact ADR 0003
// section 4 transition table, so any accidental future change to
// allowedTransitions that adds an unintended edge is caught immediately.
func TestCanTransition_ExhaustiveMatrix(t *testing.T) {
	allowed := map[Status]map[Status]bool{
		StatusDraft:        {StatusScreening: true},
		StatusScreening:    {StatusSubmitted: true, StatusRejected: true, StatusWithdrawn: true},
		StatusSubmitted:    {StatusInterviewing: true, StatusRejected: true, StatusWithdrawn: true},
		StatusInterviewing: {StatusOffered: true, StatusRejected: true, StatusWithdrawn: true},
		StatusOffered:      {StatusJoined: true, StatusRejected: true, StatusWithdrawn: true},
		StatusJoined:       {},
		StatusRejected:     {},
		StatusWithdrawn:    {},
	}

	all := []Status{StatusDraft, StatusScreening, StatusSubmitted, StatusInterviewing, StatusOffered, StatusJoined, StatusRejected, StatusWithdrawn}

	for _, from := range all {
		for _, to := range all {
			want := allowed[from][to]
			got := CanTransition(from, to)
			if got != want {
				t.Errorf("CanTransition(%q, %q) = %v, want %v", from, to, got, want)
			}
		}
	}
}

func TestCanTransition_NoBackwardOrSkippedTransitions(t *testing.T) {
	invalid := []struct{ from, to Status }{
		{StatusDraft, StatusSubmitted},     // skips screening
		{StatusDraft, StatusInterviewing},  // skips multiple steps
		{StatusScreening, StatusDraft},     // backward
		{StatusSubmitted, StatusScreening}, // backward
		{StatusOffered, StatusSubmitted},   // backward
		{StatusInterviewing, StatusDraft},  // backward, skips
		{StatusDraft, StatusOffered},       // skips everything
		{StatusDraft, StatusJoined},        // skips everything
	}
	for _, tc := range invalid {
		if CanTransition(tc.from, tc.to) {
			t.Errorf("CanTransition(%q, %q) = true, want false (skip/backward transition must be rejected)", tc.from, tc.to)
		}
	}
}

func TestCanTransition_TerminalStatesHaveNoOutgoingTransitions(t *testing.T) {
	all := []Status{StatusDraft, StatusScreening, StatusSubmitted, StatusInterviewing, StatusOffered, StatusJoined, StatusRejected, StatusWithdrawn}
	for _, terminal := range []Status{StatusJoined, StatusRejected, StatusWithdrawn} {
		for _, to := range all {
			if CanTransition(terminal, to) {
				t.Errorf("CanTransition(%q, %q) = true, want false — terminal status must have no outgoing transitions", terminal, to)
			}
		}
	}
}

func TestStatus_IsTerminal(t *testing.T) {
	terminal := []Status{StatusJoined, StatusRejected, StatusWithdrawn}
	nonTerminal := []Status{StatusDraft, StatusScreening, StatusSubmitted, StatusInterviewing, StatusOffered}

	for _, s := range terminal {
		if !s.IsTerminal() {
			t.Errorf("%q.IsTerminal() = false, want true", s)
		}
	}
	for _, s := range nonTerminal {
		if s.IsTerminal() {
			t.Errorf("%q.IsTerminal() = true, want false", s)
		}
	}
}

func TestStatus_Valid(t *testing.T) {
	valid := []Status{StatusDraft, StatusScreening, StatusSubmitted, StatusInterviewing, StatusOffered, StatusJoined, StatusRejected, StatusWithdrawn}
	for _, s := range valid {
		if !s.Valid() {
			t.Errorf("%q.Valid() = false, want true", s)
		}
	}

	invalid := []Status{"pending", "approved", "", "Draft", "DRAFT"}
	for _, s := range invalid {
		if s.Valid() {
			t.Errorf("%q.Valid() = true, want false", s)
		}
	}
}

func TestAssignment_TransitionTo_ValidPathMutatesStatus(t *testing.T) {
	a := &Assignment{Status: StatusDraft}
	if err := a.TransitionTo(StatusScreening); err != nil {
		t.Fatalf("TransitionTo(screening) from draft failed: %v", err)
	}
	if a.Status != StatusScreening {
		t.Errorf("a.Status = %q, want %q", a.Status, StatusScreening)
	}
}

func TestAssignment_TransitionTo_InvalidTransitionRejectedAndStatusUnchanged(t *testing.T) {
	a := &Assignment{Status: StatusDraft}
	err := a.TransitionTo(StatusOffered) // skips everything
	if err == nil {
		t.Fatal("TransitionTo(offered) from draft succeeded, want error")
	}
	var transitionErr *TransitionError
	if _, ok := err.(*TransitionError); !ok {
		t.Errorf("error type = %T, want *TransitionError", err)
	}
	_ = transitionErr
	if a.Status != StatusDraft {
		t.Errorf("a.Status = %q after rejected transition, want unchanged %q", a.Status, StatusDraft)
	}
}

func TestAssignment_TransitionTo_TerminalStateBlocksFurtherTransitions(t *testing.T) {
	a := &Assignment{Status: StatusRejected}
	err := a.TransitionTo(StatusScreening)
	if err == nil {
		t.Fatal("TransitionTo succeeded from a terminal state, want error")
	}
	if a.Status != StatusRejected {
		t.Errorf("a.Status = %q, want unchanged %q (terminal state must not move)", a.Status, StatusRejected)
	}
}

func TestAssignment_TransitionTo_RejectsUnknownStatus(t *testing.T) {
	a := &Assignment{Status: StatusScreening}
	err := a.TransitionTo(Status("cancelled")) // not in the ADR 0003 vocabulary
	if err == nil {
		t.Fatal("TransitionTo(\"cancelled\") succeeded, want error — not a recognized status")
	}
	if a.Status != StatusScreening {
		t.Errorf("a.Status = %q, want unchanged %q", a.Status, StatusScreening)
	}
}

func TestAssignment_TransitionTo_FullHappyPathToJoined(t *testing.T) {
	a := &Assignment{Status: StatusDraft}
	path := []Status{StatusScreening, StatusSubmitted, StatusInterviewing, StatusOffered, StatusJoined}
	for _, next := range path {
		if err := a.TransitionTo(next); err != nil {
			t.Fatalf("TransitionTo(%q) failed at status %q: %v", next, a.Status, err)
		}
	}
	if a.Status != StatusJoined {
		t.Errorf("final status = %q, want %q", a.Status, StatusJoined)
	}
	// And now it must be stuck.
	if err := a.TransitionTo(StatusDraft); err == nil {
		t.Error("transition out of joined succeeded, want error (terminal)")
	}
}
