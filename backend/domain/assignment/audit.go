package assignment

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
)

// AuditAction is one of the exact event names ADR 0006 section 4 defines
// for Recruitment Assignment. This is the closed set — nothing outside
// this list is written by this package.
type AuditAction string

const (
	AuditAssignmentCreated             AuditAction = "assignment.created"
	AuditAssignmentScreeningStarted    AuditAction = "assignment.screening_started"
	AuditAssignmentSubmitted           AuditAction = "assignment.submitted"
	AuditAssignmentInterviewingStarted AuditAction = "assignment.interviewing_started"
	AuditAssignmentOffered             AuditAction = "assignment.offered"
	AuditAssignmentJoined              AuditAction = "assignment.joined"
	AuditAssignmentRejected            AuditAction = "assignment.rejected"
	AuditAssignmentWithdrawn           AuditAction = "assignment.withdrawn"
	AuditAssignmentOwnerChanged        AuditAction = "assignment.owner_changed"
	AuditAssignmentSnapshotCreated     AuditAction = "assignment.snapshot_created"
)

// auditActionForStatus maps a lifecycle Status to the transition event ADR
// 0006 section 4 names for it. Only statuses reachable via
// TransitionAssignment need an entry; StatusDraft has none (creation has
// its own event, AuditAssignmentCreated, emitted by CreateAssignment, not
// this mapping).
var auditActionForStatus = map[Status]AuditAction{
	StatusScreening:    AuditAssignmentScreeningStarted,
	StatusSubmitted:    AuditAssignmentSubmitted,
	StatusInterviewing: AuditAssignmentInterviewingStarted,
	StatusOffered:      AuditAssignmentOffered,
	StatusJoined:       AuditAssignmentJoined,
	StatusRejected:     AuditAssignmentRejected,
	StatusWithdrawn:    AuditAssignmentWithdrawn,
}

// newCorrelationID generates a short random identifier so multiple audit
// events emitted by a single outer service-method call (e.g. "submitted"
// and "snapshot_created" from one TransitionAssignment call) can be
// grouped as belonging to the same business operation. No new dependency
// is pulled in for this — 16 random bytes, hex-encoded, is sufficient for
// grouping/correlation purposes and doesn't need to be a formal UUID.
func newCorrelationID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// recordAuditEventTx inserts one audit_events row within tx. tenantID and
// actorUserID must always come from authenticated context — this function
// has no way to enforce that itself (it just writes what it's given), so
// every caller in this package is responsible for that, matching how
// tenant/actor identity is handled everywhere else in this codebase (see
// #33/#34: never accepted from request payloads).
//
// metadata is marshaled to JSON here; pass a small struct or map — never
// an entire Assignment/Candidate/Requirement (ADR 0006 section 3
// explicitly forbids full-entity serialization into audit metadata).
//
// There is deliberately no corresponding update/delete function anywhere
// in this package: audit events are append-only (ADR 0006 section 5), and
// the absence of any UPDATE/DELETE-capable function against audit_events
// is how that's enforced at the application layer, consistent with how
// the table's own migration comment documents this.
func recordAuditEventTx(tx *sql.Tx, tenantID string, actorUserID int, entityID int, action AuditAction, correlationID string, metadata interface{}) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO audit_events (tenant_id, actor_user_id, entity_type, entity_id, action, correlation_id, metadata)
		VALUES ($1, $2, 'recruitment_assignment', $3, $4, $5, $6)`,
		tenantID, actorUserID, entityID, string(action), correlationID, metadataJSON,
	)
	return err
}
