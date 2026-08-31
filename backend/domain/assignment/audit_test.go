package assignment

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
)

// auditRow is a minimal query-side representation of one audit_events row,
// used only by these tests.
type auditRow struct {
	ID            int
	TenantID      string
	ActorUserID   int
	EntityType    string
	EntityID      int
	Action        string
	CorrelationID sql.NullString
	Metadata      json.RawMessage
}

func queryAuditEvents(t *testing.T, db *sql.DB, tenantID string, entityID int) []auditRow {
	t.Helper()
	rows, err := db.Query(`
		SELECT id, tenant_id, actor_user_id, entity_type, entity_id, action, correlation_id, metadata
		FROM audit_events WHERE tenant_id = $1 AND entity_id = $2 ORDER BY id`, tenantID, entityID)
	if err != nil {
		t.Fatalf("queryAuditEvents failed: %v", err)
	}
	defer rows.Close()

	var results []auditRow
	for rows.Next() {
		var r auditRow
		if err := rows.Scan(&r.ID, &r.TenantID, &r.ActorUserID, &r.EntityType, &r.EntityID, &r.Action, &r.CorrelationID, &r.Metadata); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		results = append(results, r)
	}
	return results
}

// TestAudit_CreateEmitsAssignmentCreated verifies CreateAssignment records
// exactly one assignment.created event with the correct actor/tenant/
// entity fields and candidateId/requirementId metadata.
func TestAudit_CreateEmitsAssignmentCreated(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)

	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}

	events := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	e := events[0]
	if e.Action != string(AuditAssignmentCreated) {
		t.Errorf("action = %q, want %q", e.Action, AuditAssignmentCreated)
	}
	if e.TenantID != "asg_tenant_a" {
		t.Errorf("tenant_id = %q, want %q", e.TenantID, "asg_tenant_a")
	}
	if e.ActorUserID != f.userID {
		t.Errorf("actor_user_id = %d, want %d", e.ActorUserID, f.userID)
	}
	if e.EntityType != "recruitment_assignment" {
		t.Errorf("entity_type = %q, want %q", e.EntityType, "recruitment_assignment")
	}
	if e.EntityID != a.ID {
		t.Errorf("entity_id = %d, want %d", e.EntityID, a.ID)
	}
	if !e.CorrelationID.Valid || e.CorrelationID.String == "" {
		t.Error("correlation_id is empty")
	}

	var meta map[string]int
	if err := json.Unmarshal(e.Metadata, &meta); err != nil {
		t.Fatalf("could not unmarshal metadata: %v", err)
	}
	if meta["candidateId"] != f.candidateID {
		t.Errorf("metadata.candidateId = %d, want %d", meta["candidateId"], f.candidateID)
	}
	if meta["requirementId"] != f.requirementID {
		t.Errorf("metadata.requirementId = %d, want %d", meta["requirementId"], f.requirementID)
	}
}

// TestAudit_AllLifecycleTransitionsEmitCorrectEvent walks every reachable
// transition and checks the exact event name + from/to metadata.
func TestAudit_AllLifecycleTransitionsEmitCorrectEvent(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	cases := []struct {
		to     Status
		action AuditAction
	}{
		{StatusScreening, AuditAssignmentScreeningStarted},
		{StatusSubmitted, AuditAssignmentSubmitted},
		{StatusInterviewing, AuditAssignmentInterviewingStarted},
		{StatusOffered, AuditAssignmentOffered},
		{StatusJoined, AuditAssignmentJoined},
	}

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}

	fromStatus := StatusDraft
	for _, c := range cases {
		t.Run(string(c.to), func(t *testing.T) {
			if _, err := svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, c.to); err != nil {
				t.Fatalf("transition to %q failed: %v", c.to, err)
			}

			events := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
			var found *auditRow
			for i := range events {
				if events[i].Action == string(c.action) {
					found = &events[i]
				}
			}
			if found == nil {
				t.Fatalf("no %q event found among %d events", c.action, len(events))
			}

			var meta map[string]string
			json.Unmarshal(found.Metadata, &meta)
			if meta["from"] != string(fromStatus) {
				t.Errorf("metadata.from = %q, want %q", meta["from"], fromStatus)
			}
			if meta["to"] != string(c.to) {
				t.Errorf("metadata.to = %q, want %q", meta["to"], c.to)
			}

			fromStatus = c.to
		})
	}
}

// TestAudit_RejectedAndWithdrawnEmitCorrectEvents covers the two terminal
// events not reached by the happy-path walk above.
func TestAudit_RejectedAndWithdrawnEmitCorrectEvents(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()
	svc := NewService(NewPostgresRepository(db), db)

	t.Run("rejected", func(t *testing.T) {
		f := seedFixtures(t, db, "asg_tenant_a")
		a, _ := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
		svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening)
		if _, err := svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusRejected); err != nil {
			t.Fatalf("transition failed: %v", err)
		}
		events := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
		if !containsAction(events, AuditAssignmentRejected) {
			t.Errorf("no %q event found among %d events", AuditAssignmentRejected, len(events))
		}
	})

	t.Run("withdrawn", func(t *testing.T) {
		f := seedFixtures(t, db, "asg_tenant_a")
		a, _ := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
		svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening)
		if _, err := svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusWithdrawn); err != nil {
			t.Fatalf("transition failed: %v", err)
		}
		events := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
		if !containsAction(events, AuditAssignmentWithdrawn) {
			t.Errorf("no %q event found among %d events", AuditAssignmentWithdrawn, len(events))
		}
	})
}

func containsAction(events []auditRow, action AuditAction) bool {
	for _, e := range events {
		if e.Action == string(action) {
			return true
		}
	}
	return false
}

// TestAudit_SubmittedProducesExactlyTwoEventsSharingCorrelationID is the
// specific requirement that submitted + snapshot_created are both
// recorded, share one correlation_id, and snapshot_created's metadata
// contains no candidate/requirement PII.
func TestAudit_SubmittedProducesExactlyTwoEventsSharingCorrelationID(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening)

	if _, err := svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusSubmitted); err != nil {
		t.Fatalf("submission failed: %v", err)
	}

	all := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
	var submittedEvents, snapshotEvents []auditRow
	for _, e := range all {
		switch e.Action {
		case string(AuditAssignmentSubmitted):
			submittedEvents = append(submittedEvents, e)
		case string(AuditAssignmentSnapshotCreated):
			snapshotEvents = append(snapshotEvents, e)
		}
	}

	if len(submittedEvents) != 1 {
		t.Fatalf("assignment.submitted event count = %d, want 1", len(submittedEvents))
	}
	if len(snapshotEvents) != 1 {
		t.Fatalf("assignment.snapshot_created event count = %d, want 1", len(snapshotEvents))
	}

	submitted, snapshot := submittedEvents[0], snapshotEvents[0]
	if !submitted.CorrelationID.Valid || !snapshot.CorrelationID.Valid {
		t.Fatal("correlation_id missing on one or both events")
	}
	if submitted.CorrelationID.String != snapshot.CorrelationID.String {
		t.Errorf("correlation_id mismatch: submitted=%q snapshot_created=%q, want equal", submitted.CorrelationID.String, snapshot.CorrelationID.String)
	}

	// snapshot_created metadata must contain no candidate/requirement PII
	// — check that the actual candidate identifying data (name, email)
	// does not appear anywhere in the metadata JSON.
	var meta map[string]interface{}
	if err := json.Unmarshal(snapshot.Metadata, &meta); err != nil {
		t.Fatalf("could not unmarshal snapshot_created metadata: %v", err)
	}
	if len(meta) != 0 {
		t.Errorf("snapshot_created metadata = %v, want empty (no candidate/requirement content)", meta)
	}
	metaStr := string(snapshot.Metadata)
	if metaStr == "" {
		t.Error("snapshot_created metadata is empty string, want at least '{}'")
	}
	if strings.Contains(metaStr, "Test Candidate") || strings.Contains(metaStr, "cand@test.com") {
		t.Errorf("snapshot_created metadata appears to contain candidate PII: %s", metaStr)
	}
}

// TestAudit_OtherTransitionsDoNotProduceSnapshotCreated verifies
// snapshot_created is ONLY emitted at the screening->submitted boundary,
// never on any other transition.
func TestAudit_OtherTransitionsDoNotProduceSnapshotCreated(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}

	if _, err := svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening); err != nil {
		t.Fatalf("draft->screening failed: %v", err)
	}

	events := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
	if containsAction(events, AuditAssignmentSnapshotCreated) {
		t.Error("assignment.snapshot_created emitted on draft->screening, want only at screening->submitted")
	}
}

// TestAudit_OwnerChangedEvent verifies ChangeOwner records an
// assignment.owner_changed event with previousOwnerUserId/newOwnerUserId.
func TestAudit_OwnerChangedEvent(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	var newOwnerID int
	db.QueryRow(`INSERT INTO users (username, email, password, role, tenant_id, company_name) VALUES ('newowner_audit', 'newowner_audit@test.com', 'x', 'manager', 'asg_tenant_a', 'asg_tenant_a') RETURNING id`).Scan(&newOwnerID)

	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}

	if _, err := svc.ChangeOwner("asg_tenant_a", f.userID, a.ID, newOwnerID); err != nil {
		t.Fatalf("ChangeOwner failed: %v", err)
	}

	events := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
	var found *auditRow
	for i := range events {
		if events[i].Action == string(AuditAssignmentOwnerChanged) {
			found = &events[i]
		}
	}
	if found == nil {
		t.Fatal("no assignment.owner_changed event found")
	}
	if found.ActorUserID != f.userID {
		t.Errorf("actor_user_id = %d, want %d", found.ActorUserID, f.userID)
	}

	var meta map[string]int
	json.Unmarshal(found.Metadata, &meta)
	if meta["previousOwnerUserId"] != f.userID {
		t.Errorf("metadata.previousOwnerUserId = %d, want %d", meta["previousOwnerUserId"], f.userID)
	}
	if meta["newOwnerUserId"] != newOwnerID {
		t.Errorf("metadata.newOwnerUserId = %d, want %d", meta["newOwnerUserId"], newOwnerID)
	}
}

// TestAudit_IllegalTransitionProducesNoEvent verifies a rejected transition
// attempt writes nothing to audit_events.
func TestAudit_IllegalTransitionProducesNoEvent(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	beforeCount := len(queryAuditEvents(t, db, "asg_tenant_a", a.ID))

	// draft -> offered skips everything; must be rejected.
	_, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusOffered)
	if err == nil {
		t.Fatal("expected illegal transition to fail")
	}

	afterCount := len(queryAuditEvents(t, db, "asg_tenant_a", a.ID))
	if afterCount != beforeCount {
		t.Errorf("audit event count changed from %d to %d after an illegal transition, want unchanged", beforeCount, afterCount)
	}
}

// TestAudit_CrossTenantOperationProducesNoEvent verifies a cross-tenant
// transition attempt writes nothing to audit_events, in either tenant.
func TestAudit_CrossTenantOperationProducesNoEvent(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fB := seedFixtures(t, db, "asg_tenant_b")
	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_b", fB.userID, CreateInput{CandidateID: fB.candidateID, RequirementID: fB.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	beforeCount := len(queryAuditEvents(t, db, "asg_tenant_b", a.ID))

	_, err = svc.TransitionAssignment("asg_tenant_a", fB.userID, a.ID, StatusScreening)
	if err != ErrNotFound {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}

	afterCount := len(queryAuditEvents(t, db, "asg_tenant_b", a.ID))
	if afterCount != beforeCount {
		t.Errorf("Tenant B's audit event count changed from %d to %d after a cross-tenant attempt, want unchanged", beforeCount, afterCount)
	}

	// And confirm nothing was written under Tenant A's tenant_id either.
	leaked := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
	if len(leaked) != 0 {
		t.Errorf("found %d audit events under asg_tenant_a for a cross-tenant attempt, want 0", len(leaked))
	}
}

// TestAudit_TenantIsolationOnQuery verifies Tenant A cannot read Tenant B's
// audit events even by entity_id collision (different tenants can have
// assignments with the same auto-increment id in principle, though in
// practice ids are global here — this test still confirms the tenant_id
// filter is what actually matters).
func TestAudit_TenantIsolationOnQuery(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fA := seedFixtures(t, db, "asg_tenant_a")
	fB := seedFixtures(t, db, "asg_tenant_b")
	svc := NewService(NewPostgresRepository(db), db)

	aA, _ := svc.CreateAssignment("asg_tenant_a", fA.userID, CreateInput{CandidateID: fA.candidateID, RequirementID: fA.requirementID})
	aB, _ := svc.CreateAssignment("asg_tenant_b", fB.userID, CreateInput{CandidateID: fB.candidateID, RequirementID: fB.requirementID})

	eventsForA := queryAuditEvents(t, db, "asg_tenant_a", aA.ID)
	for _, e := range eventsForA {
		if e.TenantID != "asg_tenant_a" {
			t.Errorf("query scoped to asg_tenant_a returned a row with tenant_id = %q", e.TenantID)
		}
	}

	// Querying Tenant A's tenant_id with Tenant B's assignment id must
	// return nothing.
	crossResult := queryAuditEvents(t, db, "asg_tenant_a", aB.ID)
	if len(crossResult) != 0 {
		t.Errorf("cross-tenant query returned %d rows, want 0", len(crossResult))
	}
}

// TestAudit_TransactionRollbackIncludesAuditRows reuses the poison-trigger
// technique from checkpoint 5 to force a mid-transaction failure on a
// submission, and verifies NO audit event (not "submitted", not
// "snapshot_created") was left behind by the rolled-back attempt.
func TestAudit_TransactionRollbackIncludesAuditRows(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	db.Exec(`
		CREATE OR REPLACE FUNCTION test_audit_reject_poisoned_submission() RETURNS TRIGGER AS $$
		BEGIN
			IF NEW.candidate_snapshot->>'name' = 'AUDIT_ROLLBACK_POISON' THEN
				RAISE EXCEPTION 'simulated failure for audit rollback test';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`)
	db.Exec(`DROP TRIGGER IF EXISTS trg_audit_reject_poisoned_submission ON recruitment_assignments`)
	db.Exec(`
		CREATE TRIGGER trg_audit_reject_poisoned_submission
		BEFORE UPDATE ON recruitment_assignments
		FOR EACH ROW EXECUTE FUNCTION test_audit_reject_poisoned_submission();
	`)
	defer db.Exec(`DROP TRIGGER IF EXISTS trg_audit_reject_poisoned_submission ON recruitment_assignments`)

	f := seedFixtures(t, db, "asg_tenant_a")
	db.Exec(`UPDATE candidates SET name = 'AUDIT_ROLLBACK_POISON' WHERE id = $1`, f.candidateID)

	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	if _, err := svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening); err != nil {
		t.Fatalf("draft->screening failed: %v", err)
	}

	beforeCount := len(queryAuditEvents(t, db, "asg_tenant_a", a.ID))

	_, err = svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusSubmitted)
	if err == nil {
		t.Fatal("expected the poisoned submission to fail")
	}

	afterCount := len(queryAuditEvents(t, db, "asg_tenant_a", a.ID))
	if afterCount != beforeCount {
		t.Errorf("audit event count changed from %d to %d after a rolled-back submission, want unchanged (no partial submitted/snapshot_created rows)", beforeCount, afterCount)
	}

	events := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
	if containsAction(events, AuditAssignmentSubmitted) {
		t.Error("assignment.submitted event present despite rolled-back transaction")
	}
	if containsAction(events, AuditAssignmentSnapshotCreated) {
		t.Error("assignment.snapshot_created event present despite rolled-back transaction")
	}
}

// TestAudit_CreateTransactionRollback verifies that if CreateAssignment's
// persist step fails, no assignment AND no assignment.created audit event
// remain — tested via the duplicate-(candidate,requirement) constraint,
// which fails inside the same transaction the audit INSERT would occur in.
func TestAudit_CreateTransactionRollback(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)

	first, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("first CreateAssignment failed: %v", err)
	}
	firstEventCount := len(queryAuditEvents(t, db, "asg_tenant_a", first.ID))
	if firstEventCount != 1 {
		t.Fatalf("first assignment event count = %d, want 1", firstEventCount)
	}

	// Second attempt with the same (candidate, requirement) pair must fail
	// on the UNIQUE constraint, inside the transaction, before commit.
	_, err = svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != ErrDuplicateAssignment {
		t.Fatalf("error = %v, want ErrDuplicateAssignment", err)
	}

	// The first assignment's audit trail must be completely unaffected,
	// and no new "phantom" assignment or audit event should exist for a
	// failed second attempt.
	var totalAssignments int
	db.QueryRow(`SELECT COUNT(*) FROM recruitment_assignments WHERE tenant_id = 'asg_tenant_a' AND candidate_id = $1 AND requirement_id = $2`,
		f.candidateID, f.requirementID).Scan(&totalAssignments)
	if totalAssignments != 1 {
		t.Errorf("total assignments for this (candidate, requirement) pair = %d, want 1", totalAssignments)
	}

	stillFirstEventCount := len(queryAuditEvents(t, db, "asg_tenant_a", first.ID))
	if stillFirstEventCount != 1 {
		t.Errorf("first assignment event count after failed second create = %d, want unchanged 1", stillFirstEventCount)
	}
}

// TestAudit_NoUpdateOrDeletePathExists is a structural test: this package
// exposes no function capable of updating or deleting an audit_events row.
// Enforced by code review / the absence of such a function, documented
// here so the invariant is explicit and discoverable rather than implicit.
// (There is deliberately nothing to call here — see audit.go's comment on
// recordAuditEventTx for the actual enforcement.)
func TestAudit_NoUpdateOrDeletePathExists(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}

	events := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	originalAction := events[0].Action

	// Perform several more operations on the same assignment and confirm
	// the original event's action is never altered — there is no code
	// path in this package that would do so, and this pins that fact.
	svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusScreening)
	svc.TransitionAssignment("asg_tenant_a", f.userID, a.ID, StatusSubmitted)

	reloadedEvents := queryAuditEvents(t, db, "asg_tenant_a", a.ID)
	for _, e := range reloadedEvents {
		if e.ID == events[0].ID && e.Action != originalAction {
			t.Errorf("original assignment.created event's action changed from %q to %q", originalAction, e.Action)
		}
	}
}
