package assignment

import "fmt"

// Status is the Recruitment Assignment lifecycle state, per ADR 0003
// section 4. It is intentionally its own type (not a bare string) so the
// compiler catches accidental use of an arbitrary string where a validated
// Status is expected.
type Status string

const (
	StatusDraft        Status = "draft"
	StatusScreening    Status = "screening"
	StatusSubmitted    Status = "submitted"
	StatusInterviewing Status = "interviewing"
	StatusOffered      Status = "offered"
	StatusJoined       Status = "joined"
	StatusRejected     Status = "rejected"
	StatusWithdrawn    Status = "withdrawn"
)

// allStatuses is the complete, closed set of valid values — must stay in
// sync with the recruitment_assignments_status_valid CHECK constraint in
// 008_recruitment_assignment_domain.sql.
var allStatuses = map[Status]bool{
	StatusDraft:        true,
	StatusScreening:    true,
	StatusSubmitted:    true,
	StatusInterviewing: true,
	StatusOffered:      true,
	StatusJoined:       true,
	StatusRejected:     true,
	StatusWithdrawn:    true,
}

// terminalStatuses are the outcomes ADR 0003 section 4 designates as
// terminal: they must not be moved backward through ordinary API
// operations.
var terminalStatuses = map[Status]bool{
	StatusJoined:    true,
	StatusRejected:  true,
	StatusWithdrawn: true,
}

// allowedTransitions encodes ADR 0003 section 4's transition table exactly.
// This is the single source of truth for what transitions are legal — no
// other code (handlers, services) should independently decide this.
var allowedTransitions = map[Status][]Status{
	StatusDraft:        {StatusScreening},
	StatusScreening:    {StatusSubmitted, StatusRejected, StatusWithdrawn},
	StatusSubmitted:    {StatusInterviewing, StatusRejected, StatusWithdrawn},
	StatusInterviewing: {StatusOffered, StatusRejected, StatusWithdrawn},
	StatusOffered:      {StatusJoined, StatusRejected, StatusWithdrawn},
	StatusJoined:       {}, // terminal
	StatusRejected:     {}, // terminal
	StatusWithdrawn:    {}, // terminal
}

// Valid reports whether s is one of the closed set of known statuses.
func (s Status) Valid() bool {
	return allStatuses[s]
}

// IsTerminal reports whether s is a terminal outcome (ADR 0003 section 4).
func (s Status) IsTerminal() bool {
	return terminalStatuses[s]
}

// CanTransition reports whether moving from `from` to `to` is a legal
// transition per ADR 0003 section 4. This is the conceptual
// `CanTransition(from, to) bool` mechanism: the one place that knows the
// workflow rules, so handlers and other callers never need to.
func CanTransition(from, to Status) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// TransitionError describes why a requested transition was rejected. It
// carries the attempted from/to so callers (eventually HTTP handlers) can
// render a specific, useful error without re-deriving the workflow logic
// themselves.
type TransitionError struct {
	From   Status
	To     Status
	Reason string
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("cannot transition assignment from %q to %q: %s", e.From, e.To, e.Reason)
}
