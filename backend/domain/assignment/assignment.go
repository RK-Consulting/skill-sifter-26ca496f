package assignment

import "time"

// Assignment is the Recruitment Assignment domain model — the durable
// transaction connecting a Candidate to a Requirement (ADR 0003 section 1).
// This is a domain type, deliberately separate from any HTTP request/
// response shape: handlers (a later checkpoint) will translate to/from
// this type rather than operating on it directly as a wire format.
type Assignment struct {
	ID              int
	TenantID        string
	CandidateID     int
	RequirementID   int
	Status          Status
	CreatedByUserID int
	OwnerUserID     int

	// Submission snapshots (ADR 0003 section 6). Nil/zero until formal
	// submission. Populated by a later checkpoint's submission logic, not
	// by TransitionTo itself — capturing a snapshot is a distinct concern
	// from validating a lifecycle transition, and the caller (service
	// layer) is responsible for setting these before persisting a
	// transition into StatusSubmitted.
	CandidateSnapshot   []byte // raw JSON
	RequirementSnapshot []byte // raw JSON
	SnapshotCreatedAt   *time.Time

	CreatedAt    time.Time
	LastModified time.Time
}

// TransitionTo attempts to move the assignment to newStatus, enforcing
// ADR 0003's transition rules via CanTransition. It mutates the in-memory
// Assignment on success and returns a *TransitionError on failure. It does
// NOT persist the change — the caller (service layer) is responsible for
// persisting via the repository after a successful transition. This
// mirrors the "assignment.TransitionTo(submitted)" shape requested: the
// handler/service tells the domain object what it wants, and the domain
// object is the only place that knows whether that's allowed.
func (a *Assignment) TransitionTo(newStatus Status) error {
	if !newStatus.Valid() {
		return &TransitionError{From: a.Status, To: newStatus, Reason: "target status is not a recognized assignment status"}
	}
	if a.Status.IsTerminal() {
		return &TransitionError{From: a.Status, To: newStatus, Reason: "assignment is in a terminal state and cannot transition further"}
	}
	if !CanTransition(a.Status, newStatus) {
		return &TransitionError{From: a.Status, To: newStatus, Reason: "not a valid transition"}
	}
	a.Status = newStatus
	return nil
}
