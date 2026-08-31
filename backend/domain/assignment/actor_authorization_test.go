package assignment

import (
	"database/sql"
	"errors"
	"testing"
)

// TestTransitionAssignment_CrossTenantActorIsRejected verifies that an actor
// from another tenant cannot transition an assignment, even when the
// assignment itself belongs to the authenticated tenant. The transaction
// must roll back so neither the status nor an audit event changes.
func TestTransitionAssignment_CrossTenantActorIsRejected(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fA := seedFixtures(t, db, "actor_tenant_a")
	fB := seedFixtures(t, db, "actor_tenant_b")
	svc := NewService(NewPostgresRepository(db), db)

	created, err := svc.CreateAssignment("actor_tenant_a", fA.userID, CreateInput{
		CandidateID:   fA.candidateID,
		RequirementID: fA.requirementID,
	})
	if err != nil {
		t.Fatalf("seed CreateAssignment failed: %v", err)
	}

	beforeAuditCount := auditEventCount(t, db, "actor_tenant_a", created.ID)

	_, err = svc.TransitionAssignment("actor_tenant_a", fB.userID, created.ID, StatusScreening)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("error = %v, want ErrUserNotFound for cross-tenant actor", err)
	}

	reloaded, err := svc.GetAssignment("actor_tenant_a", created.ID)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if reloaded.Status != StatusDraft {
		t.Errorf("status = %q, want unchanged %q", reloaded.Status, StatusDraft)
	}

	afterAuditCount := auditEventCount(t, db, "actor_tenant_a", created.ID)
	if afterAuditCount != beforeAuditCount {
		t.Errorf("audit event count = %d, want unchanged %d", afterAuditCount, beforeAuditCount)
	}
}

// TestChangeOwner_CrossTenantActorIsRejected verifies the same invariant for
// owner changes. The target owner is valid in the assignment tenant; only the
// actor is cross-tenant, so the failure specifically exercises actor
// authorization rather than owner validation.
func TestChangeOwner_CrossTenantActorIsRejected(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fA := seedFixtures(t, db, "owner_actor_tenant_a")
	fB := seedFixtures(t, db, "owner_actor_tenant_b")
	svc := NewService(NewPostgresRepository(db), db)

	created, err := svc.CreateAssignment("owner_actor_tenant_a", fA.userID, CreateInput{
		CandidateID:   fA.candidateID,
		RequirementID: fA.requirementID,
	})
	if err != nil {
		t.Fatalf("seed CreateAssignment failed: %v", err)
	}

	beforeAuditCount := auditEventCount(t, db, "owner_actor_tenant_a", created.ID)

	_, err = svc.ChangeOwner("owner_actor_tenant_a", fB.userID, created.ID, fA.userID)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("error = %v, want ErrUserNotFound for cross-tenant actor", err)
	}

	reloaded, err := svc.GetAssignment("owner_actor_tenant_a", created.ID)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if reloaded.OwnerUserID != fA.userID {
		t.Errorf("owner_user_id = %d, want unchanged %d", reloaded.OwnerUserID, fA.userID)
	}

	afterAuditCount := auditEventCount(t, db, "owner_actor_tenant_a", created.ID)
	if afterAuditCount != beforeAuditCount {
		t.Errorf("audit event count = %d, want unchanged %d", afterAuditCount, beforeAuditCount)
	}
}

func auditEventCount(t *testing.T, db *sql.DB, tenantID string, entityID int) int {
	t.Helper()

	// Kept as a small local query helper so these regression tests verify the
	// externally visible invariant without coupling to audit repository code.
	var count int
	row := db.QueryRow(`SELECT COUNT(*) FROM audit_events WHERE tenant_id = $1 AND entity_type = 'recruitment_assignment' AND entity_id = $2`, tenantID, entityID)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("count audit events: %v", err)
	}
	return count
}
