package assignment

import "fmt"

// Status is the Recruitment Assignment lifecycle state defined by ADR 0003.
// It is intentionally distinct from candidate master status.
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

var terminalStatuses = map[Status]bool{
	StatusJoined:    true,
	StatusRejected:  true,
	StatusWithdrawn: true,
}

// allowedTransitions is the single authoritative workflow matrix.
// Business outcomes are explicit; no generic back-transition is permitted.
var allowedTransitions = map[Status][]Status{
	StatusDraft:        {StatusScreening},
	StatusScreening:    {StatusSubmitted, StatusRejected, StatusWithdrawn},
	StatusSubmitted:    {StatusInterviewing, StatusRejected, StatusWithdrawn},
	StatusInterviewing: {StatusOffered, StatusRejected, StatusWithdrawn},
	StatusOffered:      {StatusJoined, StatusRejected, StatusWithdrawn},
	StatusJoined:       {},
	StatusRejected:     {},
	StatusWithdrawn:    {},
}

func (s Status) Valid() bool {
	return allStatuses[s]
}

func (s Status) IsTerminal() bool {
	return terminalStatuses[s]
}

// CanTransition is the single source of truth for legal assignment transitions.
func CanTransition(from, to Status) bool {
	if !from.Valid() || !to.Valid() {
		return false
	}
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// AllowedTransitions returns a copy of the legal target states for from.
// Returning a copy prevents callers from mutating the workflow definition.
func AllowedTransitions(from Status) []Status {
	allowed := allowedTransitions[from]
	result := make([]Status, len(allowed))
	copy(result, allowed)
	return result
}

type TransitionError struct {
	From   Status
	To     Status
	Reason string
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("cannot transition assignment from %q to %q: %s", e.From, e.To, e.Reason)
}
