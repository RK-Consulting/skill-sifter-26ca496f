package assignment

import (
	"errors"
	"fmt"
	"testing"
)

// TestTransitionAssignment_ExhaustiveMatrix exercises every (from, to)
// pair across all 8 statuses (64 combinations) through the full
// TransitionAssignment service call — not just the pure CanTransition
// logic already covered exhaustively in status_test.go (checkpoint 2).
// This proves the service-level wiring (fetch -> TransitionTo -> persist)
// matches the domain's transition table exactly, including that rejected
// transitions leave the persisted status untouched.
func TestTransitionAssignment_ExhaustiveMatrix(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

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

	repo := NewPostgresRepository(db)
	svc := NewService(repo, db)

	for _, from := range all {
		for _, to := range all {
			from, to := from, to // capture
			t.Run(fmt.Sprintf("%s_to_%s", from, to), func(t *testing.T) {
				f := seedFixtures(t, db, "asg_tenant_a")
				a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
				if err != nil {
					t.Fatalf("seed CreateAssignment failed: %v", err)
				}

				// Force the assignment into `from` directly (bypassing
				// TransitionTo, since not every state is reachable via a
				// legal sequential path from draft — e.g. testing FROM
				// joined requires starting there).
				a.Status = from
				if err := repo.Update(a); err != nil {
					t.Fatalf("could not force status to %q: %v", from, err)
				}

				result, err := svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, to)
				want := allowed[from][to]

				if want {
					if err != nil {
						t.Fatalf("%s -> %s: expected success, got error: %v", from, to, err)
					}
					if result.Status != to {
						t.Errorf("%s -> %s: result status = %q, want %q", from, to, result.Status, to)
					}
					reloaded, rerr := svc.GetAssignment("asg_tenant_a", a.ID)
					if rerr != nil {
						t.Fatalf("reload failed: %v", rerr)
					}
					if reloaded.Status != to {
						t.Errorf("%s -> %s: persisted status = %q, want %q", from, to, reloaded.Status, to)
					}
				} else {
					if err == nil {
						t.Fatalf("%s -> %s: expected error (illegal transition), got success", from, to)
					}
					var transitionErr *TransitionError
					if !errors.As(err, &transitionErr) {
						t.Errorf("%s -> %s: error type = %T, want *TransitionError", from, to, err)
					}
					reloaded, rerr := svc.GetAssignment("asg_tenant_a", a.ID)
					if rerr != nil {
						t.Fatalf("reload failed: %v", rerr)
					}
					if reloaded.Status != from {
						t.Errorf("%s -> %s: persisted status = %q after rejected transition, want unchanged %q", from, to, reloaded.Status, from)
					}
				}
			})
		}
	}
}

// TestTransitionAssignment_CrossTenantReturnsNotFound verifies a
// transition attempt against another tenant's assignment ID fails closed.
func TestTransitionAssignment_CrossTenantReturnsNotFound(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fB := seedFixtures(t, db, "asg_tenant_b")
	svc := NewService(NewPostgresRepository(db), db)

	created, err := svc.CreateAssignment("asg_tenant_b", fB.userID, CreateInput{CandidateID: fB.candidateID, RequirementID: fB.requirementID})
	if err != nil {
		t.Fatalf("seed CreateAssignment failed: %v", err)
	}

	_, err = svc.TransitionAssignment("asg_tenant_a", fB.userID, created.ID, StatusScreening)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound (Tenant A must not transition Tenant B's assignment)", err)
	}

	// Confirm Tenant B's assignment was not modified by the cross-tenant
	// attempt.
	reloaded, rerr := svc.GetAssignment("asg_tenant_b", created.ID)
	if rerr != nil {
		t.Fatalf("reload failed: %v", rerr)
	}
	if reloaded.Status != StatusDraft {
		t.Errorf("Tenant B's assignment status = %q after cross-tenant transition attempt, want unchanged %q", reloaded.Status, StatusDraft)
	}
}
