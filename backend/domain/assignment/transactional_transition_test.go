package assignment

import "testing"

func TestTransitionAssignmentTx_RollbackRestoresAssignment(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_atomic")
	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_atomic", f.userID, CreateInput{
		CandidateID: f.candidateID, RequirementID: f.requirementID,
	})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	if _, err := svc.TransitionAssignmentTx(tx, "asg_tenant_atomic", f.userID, a.ID, StatusScreening); err != nil {
		t.Fatalf("TransitionAssignmentTx failed: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	reloaded, err := svc.GetAssignment("asg_tenant_atomic", a.ID)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if reloaded.Status != StatusDraft {
		t.Fatalf("status after rollback = %q, want %q", reloaded.Status, StatusDraft)
	}
}

func TestTransitionAssignmentTx_CommitPersistsAssignment(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_atomic_commit")
	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_atomic_commit", f.userID, CreateInput{
		CandidateID: f.candidateID, RequirementID: f.requirementID,
	})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	if _, err := svc.TransitionAssignmentTx(tx, "asg_tenant_atomic_commit", f.userID, a.ID, StatusScreening); err != nil {
		t.Fatalf("TransitionAssignmentTx failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	reloaded, err := svc.GetAssignment("asg_tenant_atomic_commit", a.ID)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if reloaded.Status != StatusScreening {
		t.Fatalf("status after commit = %q, want %q", reloaded.Status, StatusScreening)
	}
}
