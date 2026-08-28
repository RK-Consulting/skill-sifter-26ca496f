package assignment

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

// TestSnapshot_CapturedOnlyAtSubmission verifies the snapshot is nil
// before formal submission (screening) and populated exactly at
// screening -> submitted, per ADR 0003 section 6 ("at formal submission").
func TestSnapshot_CapturedOnlyAtSubmission(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)

	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	if a.SnapshotCreatedAt != nil {
		t.Fatal("newly created assignment already has a snapshot")
	}

	a, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening)
	if err != nil {
		t.Fatalf("draft->screening failed: %v", err)
	}
	if a.SnapshotCreatedAt != nil {
		t.Error("snapshot captured before formal submission (at draft->screening) — must only capture at screening->submitted")
	}

	before := time.Now()
	a, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusSubmitted)
	if err != nil {
		t.Fatalf("screening->submitted failed: %v", err)
	}
	after := time.Now()

	if a.SnapshotCreatedAt == nil {
		t.Fatal("snapshot not captured at formal submission (screening->submitted)")
	}
	if a.SnapshotCreatedAt.Before(before) || a.SnapshotCreatedAt.After(after) {
		t.Errorf("SnapshotCreatedAt = %v, want between %v and %v", a.SnapshotCreatedAt, before, after)
	}
	if a.CandidateSnapshot == nil {
		t.Error("CandidateSnapshot is nil after submission")
	}
	if a.RequirementSnapshot == nil {
		t.Error("RequirementSnapshot is nil after submission")
	}
}

// TestSnapshot_CandidateDataCapturedCorrectly verifies the candidate
// snapshot content matches the candidate row at the moment of submission.
func TestSnapshot_CandidateDataCapturedCorrectly(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	// Overwrite the seeded candidate with fully-populated, distinctive
	// values so the assertions below can't accidentally pass on defaults.
	db.Exec(`UPDATE candidates SET name = 'Snapshot Candidate', phone = '555-1234', position = 'Backend Engineer',
		location = 'Bangalore', experience = '5 years', currentctc = '20 LPA', expectedctc = '28 LPA',
		noticeperiod = '30 days', jlptlanguage = 'N2', skills = 'Go, Postgres' WHERE id = $1`, f.candidateID)

	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening)
	a, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusSubmitted)
	if err != nil {
		t.Fatalf("submission failed: %v", err)
	}

	var snap CandidateSnapshotData
	if err := json.Unmarshal(a.CandidateSnapshot, &snap); err != nil {
		t.Fatalf("could not unmarshal candidate snapshot: %v", err)
	}

	cases := map[string]struct{ got, want string }{
		"name":         {snap.Name, "Snapshot Candidate"},
		"phone":        {snap.Phone, "555-1234"},
		"position":     {snap.Position, "Backend Engineer"},
		"location":     {snap.Location, "Bangalore"},
		"experience":   {snap.Experience, "5 years"},
		"currentCtc":   {snap.CurrentCTC, "20 LPA"},
		"expectedCtc":  {snap.ExpectedCTC, "28 LPA"},
		"noticePeriod": {snap.NoticePeriod, "30 days"},
		"jlptLanguage": {snap.JLPTLanguage, "N2"},
		"skills":       {snap.Skills, "Go, Postgres"},
		"status":       {snap.Status, "active"},
	}
	for field, c := range cases {
		if c.got != c.want {
			t.Errorf("snapshot.%s = %q, want %q", field, c.got, c.want)
		}
	}
	if snap.ID != f.candidateID {
		t.Errorf("snapshot.id = %d, want %d", snap.ID, f.candidateID)
	}
}

// TestSnapshot_RequirementDataCapturedCorrectly verifies the requirement
// snapshot content matches the requirement row at the moment of
// submission, covering every field ADR 0003 section 6 lists as minimum.
func TestSnapshot_RequirementDataCapturedCorrectly(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	db.Exec(`UPDATE requirements SET title = 'Staff Engineer', location = 'Remote', work_arrangement = 'hybrid',
		description = 'Own the platform', required_skills = 'Go, K8s', experience_required = '8+ years',
		compensation = '40 LPA', headcount = 3, language_requirement = 'English fluent' WHERE id = $1`, f.requirementID)

	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening)
	a, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusSubmitted)
	if err != nil {
		t.Fatalf("submission failed: %v", err)
	}

	var snap RequirementSnapshotData
	if err := json.Unmarshal(a.RequirementSnapshot, &snap); err != nil {
		t.Fatalf("could not unmarshal requirement snapshot: %v", err)
	}

	cases := map[string]struct{ got, want string }{
		"title":               {snap.Title, "Staff Engineer"},
		"location":            {snap.Location, "Remote"},
		"workArrangement":     {snap.WorkArrangement, "hybrid"},
		"description":         {snap.Description, "Own the platform"},
		"requiredSkills":      {snap.RequiredSkills, "Go, K8s"},
		"experienceRequired":  {snap.ExperienceRequired, "8+ years"},
		"compensation":        {snap.Compensation, "40 LPA"},
		"languageRequirement": {snap.LanguageRequirement, "English fluent"},
	}
	for field, c := range cases {
		if c.got != c.want {
			t.Errorf("snapshot.%s = %q, want %q", field, c.got, c.want)
		}
	}
	if snap.Headcount != 3 {
		t.Errorf("snapshot.headcount = %d, want 3", snap.Headcount)
	}
	if snap.ID != f.requirementID {
		t.Errorf("snapshot.id = %d, want %d", snap.ID, f.requirementID)
	}
}

// TestSnapshot_UnchangedAfterSourceRecordsModified is the core immutability
// guarantee (ADR 0003 section 6: "Snapshots are historical evidence and do
// not replace the current candidate or requirement records"). After
// submission, editing the live candidate/requirement rows must not affect
// the already-captured snapshot.
func TestSnapshot_UnchangedAfterSourceRecordsModified(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	db.Exec(`UPDATE candidates SET name = 'Original Name', skills = 'Original Skills' WHERE id = $1`, f.candidateID)
	db.Exec(`UPDATE requirements SET title = 'Original Title', compensation = 'Original Comp' WHERE id = $1`, f.requirementID)

	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening)
	a, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusSubmitted)
	if err != nil {
		t.Fatalf("submission failed: %v", err)
	}

	// Now mutate the live source records AFTER the snapshot was taken.
	db.Exec(`UPDATE candidates SET name = 'Changed Name', skills = 'Changed Skills' WHERE id = $1`, f.candidateID)
	db.Exec(`UPDATE requirements SET title = 'Changed Title', compensation = 'Changed Comp' WHERE id = $1`, f.requirementID)

	// Advance the assignment further (interviewing) to prove later
	// transitions don't touch the snapshot either.
	a, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusInterviewing)
	if err != nil {
		t.Fatalf("submitted->interviewing failed: %v", err)
	}

	var candSnap CandidateSnapshotData
	json.Unmarshal(a.CandidateSnapshot, &candSnap)
	if candSnap.Name != "Original Name" || candSnap.Skills != "Original Skills" {
		t.Errorf("candidate snapshot changed after source edit: name=%q skills=%q, want original values", candSnap.Name, candSnap.Skills)
	}

	var reqSnap RequirementSnapshotData
	json.Unmarshal(a.RequirementSnapshot, &reqSnap)
	if reqSnap.Title != "Original Title" || reqSnap.Compensation != "Original Comp" {
		t.Errorf("requirement snapshot changed after source edit: title=%q compensation=%q, want original values", reqSnap.Title, reqSnap.Compensation)
	}

	// Also reload from the database directly, not just the in-memory
	// struct, to prove the persisted snapshot is unaffected too.
	reloaded, err := svc.GetAssignment("asg_tenant_a", a.ID)
	if err != nil {
		t.Fatalf("GetAssignment failed: %v", err)
	}
	var reloadedCandSnap CandidateSnapshotData
	json.Unmarshal(reloaded.CandidateSnapshot, &reloadedCandSnap)
	if reloadedCandSnap.Name != "Original Name" {
		t.Errorf("persisted candidate snapshot name = %q after source edit, want unchanged %q", reloadedCandSnap.Name, "Original Name")
	}
}

// TestSnapshot_NotOverwrittenOnSubsequentTransitions verifies the snapshot
// captured at submission is never recaptured or altered by later
// transitions (interviewing, offered, joined), even though those
// transitions do update other columns (status, last_modified) on the same
// row.
func TestSnapshot_NotOverwrittenOnSubsequentTransitions(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)

	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening)
	a, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusSubmitted)
	if err != nil {
		t.Fatalf("submission failed: %v", err)
	}
	// Reload from the DB so the baseline we compare against has the same
	// (microsecond) precision Postgres actually persists, rather than the
	// nanosecond-precision in-memory time.Now() value TransitionAssignment
	// returns directly — otherwise every subsequent DB round-trip would
	// spuriously look "changed" purely from precision truncation, not an
	// actual overwrite.
	a, err = svc.GetAssignment("asg_tenant_a", a.ID)
	if err != nil {
		t.Fatalf("GetAssignment after submission failed: %v", err)
	}
	originalSnapshotTime := *a.SnapshotCreatedAt
	originalCandidateSnapshot := string(a.CandidateSnapshot)

	for _, next := range []Status{StatusInterviewing, StatusOffered, StatusJoined} {
		a, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, next)
		if err != nil {
			t.Fatalf("transition to %q failed: %v", next, err)
		}
		if !a.SnapshotCreatedAt.Equal(originalSnapshotTime) {
			t.Errorf("at status %q: SnapshotCreatedAt changed from %v to %v", next, originalSnapshotTime, a.SnapshotCreatedAt)
		}
		if string(a.CandidateSnapshot) != originalCandidateSnapshot {
			t.Errorf("at status %q: candidate snapshot content changed", next)
		}
	}
}

// TestSnapshot_TransactionRollbackLeavesNoPartialSnapshot forces a failure
// partway through the submission transaction (after the snapshot content
// has been computed, via a trigger that rejects the specific poisoned
// candidate name used only in this test) and verifies the entire attempt
// rolls back: the assignment's status remains "screening" and both
// snapshot columns remain NULL — never a partial state where, say, the
// candidate snapshot was written but the status wasn't updated, or
// vice versa.
func TestSnapshot_TransactionRollbackLeavesNoPartialSnapshot(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	// A trigger that deterministically rejects the UPDATE this test
	// forces, simulating a failure between snapshot computation and
	// commit. Scoped to this test only; not part of any migration.
	db.Exec(`
		CREATE OR REPLACE FUNCTION test_reject_poisoned_submission() RETURNS TRIGGER AS $$
		BEGIN
			IF NEW.candidate_snapshot->>'name' = 'ROLLBACK_POISON' THEN
				RAISE EXCEPTION 'simulated failure for rollback test';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`)
	db.Exec(`DROP TRIGGER IF EXISTS trg_reject_poisoned_submission ON recruitment_assignments`)
	db.Exec(`
		CREATE TRIGGER trg_reject_poisoned_submission
		BEFORE UPDATE ON recruitment_assignments
		FOR EACH ROW EXECUTE FUNCTION test_reject_poisoned_submission();
	`)
	defer db.Exec(`DROP TRIGGER IF EXISTS trg_reject_poisoned_submission ON recruitment_assignments`)

	f := seedFixtures(t, db, "asg_tenant_a")
	db.Exec(`UPDATE candidates SET name = 'ROLLBACK_POISON' WHERE id = $1`, f.candidateID)

	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	if _, err := svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening); err != nil {
		t.Fatalf("draft->screening failed: %v", err)
	}

	_, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusSubmitted)
	if err == nil {
		t.Fatal("expected the poisoned submission to fail, but it succeeded")
	}

	reloaded, rerr := svc.GetAssignment("asg_tenant_a", a.ID)
	if rerr != nil {
		t.Fatalf("GetAssignment after failed submission failed: %v", rerr)
	}
	if reloaded.Status != StatusScreening {
		t.Errorf("status = %q after rolled-back submission, want unchanged %q (transaction must roll back the status change too)", reloaded.Status, StatusScreening)
	}
	if reloaded.CandidateSnapshot != nil {
		t.Error("candidate_snapshot is non-nil after a rolled-back submission — partial snapshot leaked")
	}
	if reloaded.RequirementSnapshot != nil {
		t.Error("requirement_snapshot is non-nil after a rolled-back submission — partial snapshot leaked")
	}
	if reloaded.SnapshotCreatedAt != nil {
		t.Error("snapshot_created_at is non-nil after a rolled-back submission — partial snapshot leaked")
	}

	// And confirm the assignment can still be submitted normally once the
	// poison condition is removed, proving the row itself wasn't left in
	// some corrupted, permanently-stuck state by the failed attempt.
	db.Exec(`UPDATE candidates SET name = 'No Longer Poisoned' WHERE id = $1`, f.candidateID)
	final, err := svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusSubmitted)
	if err != nil {
		t.Fatalf("submission after removing poison condition failed: %v", err)
	}
	if final.SnapshotCreatedAt == nil {
		t.Error("snapshot still not captured after a successful retry")
	}
}

// TestSnapshot_CrossTenantSubmissionDoesNotLeakOrCapture verifies a
// cross-tenant transition attempt (which should already fail with
// ErrNotFound, per checkpoint 4) also does not somehow capture a snapshot
// from another tenant's data as a side effect.
func TestSnapshot_CrossTenantSubmissionDoesNotLeakOrCapture(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fB := seedFixtures(t, db, "asg_tenant_b")
	svc := NewService(NewPostgresRepository(db), db)

	a, err := svc.CreateAssignment("asg_tenant_b", fB.userID, CreateInput{CandidateID: fB.candidateID, RequirementID: fB.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	if _, err := svc.TransitionAssignment("asg_tenant_b", fB.userID, a.ID, StatusScreening); err != nil {
		t.Fatalf("draft->screening failed: %v", err)
	}

	// Tenant A attempts to submit Tenant B's assignment by ID.
	_, err = svc.TransitionAssignment("asg_tenant_a", fB.userID, a.ID, StatusSubmitted)
	if err != ErrNotFound {
		t.Errorf("error = %v, want ErrNotFound", err)
	}

	// Confirm Tenant B's assignment was not transitioned or snapshotted by
	// the cross-tenant attempt.
	reloaded, rerr := svc.GetAssignment("asg_tenant_b", a.ID)
	if rerr != nil {
		t.Fatalf("GetAssignment failed: %v", rerr)
	}
	if reloaded.Status != StatusScreening {
		t.Errorf("Tenant B's assignment status = %q after cross-tenant submission attempt, want unchanged %q", reloaded.Status, StatusScreening)
	}
	if reloaded.SnapshotCreatedAt != nil {
		t.Error("Tenant B's assignment was snapshotted by a cross-tenant Tenant A submission attempt")
	}
}

var _ = sql.ErrNoRows // keep database/sql import even if future edits trim direct usage above
