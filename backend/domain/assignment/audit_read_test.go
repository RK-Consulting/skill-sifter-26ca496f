package assignment

import "testing"

// TestGetByEntity_ReturnsWrittenAuditEvent is a minimal round-trip test for
// the audit-event read path (GetByEntity), ported forward from the
// checkpoint-6 branch as part of v0.5.4-dev.
//
// This intentionally does NOT port the full checkpoint-6 behavioral suite
// (correlation-ID sharing across an operation, transaction-rollback safety,
// PII-scrubbing in metadata, cross-tenant isolation of audit reads). Those
// tests depend on AuditEventAction-style naming and service-level behavior
// that diverged from what shipped on main, and deserve a deliberate,
// separate rewrite against the current codebase rather than a blind port.
// That rewrite is tracked as follow-up work, not dropped.
//
// This test only proves: an event written via the existing (unexported)
// recordAuditEventTx write path is retrievable via the new GetByEntity read
// path, with the same tenant, actor, and action intact.
func TestGetByEntity_ReturnsWrittenAuditEvent(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin failed: %v", err)
	}

	const entityID = 1
	metadata := map[string]interface{}{"candidateId": f.candidateID}
	if err := recordAuditEventTx(tx, "asg_tenant_a", f.userID, entityID, AuditAssignmentCreated, "test-correlation-1", metadata); err != nil {
		tx.Rollback()
		t.Fatalf("recordAuditEventTx failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("tx.Commit failed: %v", err)
	}

	repo := NewPostgresAuditEventRepository(db)
	events, err := repo.GetByEntity("asg_tenant_a", "recruitment_assignment", entityID)
	if err != nil {
		t.Fatalf("GetByEntity failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	if events[0].Action != string(AuditAssignmentCreated) {
		t.Errorf("action = %q, want %q", events[0].Action, AuditAssignmentCreated)
	}
	if events[0].TenantID != "asg_tenant_a" {
		t.Errorf("tenant_id = %q, want %q", events[0].TenantID, "asg_tenant_a")
	}
	if events[0].ActorUserID != f.userID {
		t.Errorf("actor_user_id = %d, want %d", events[0].ActorUserID, f.userID)
	}
}
