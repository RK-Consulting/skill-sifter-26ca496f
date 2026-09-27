# ADR 0012: Persistent Candidate Recruitment State

## Status

Accepted

## Context

The recruitment workflow was previously modeled around a first-class recruitment_assignments entity with service-layer lifecycle transitions.

That model introduced unnecessary indirection for the agency-first workflow. The business rules that must be enforced are simpler:

- a candidate may be screened for multiple client requirements at the same time;
- simultaneous screening capacity is bounded by a persistent counter and limit;
- a candidate must not be scheduled for another interview while an interview is active;
- interview concurrency is therefore a single candidate-level lock;
- screening, interview, submission, and selection records are recruitment history;
- Candidate + Requirement already identifies the recruitment context.

## Decision

The current recruitment control state is stored directly on the Candidate row:

- screening_count
- screening_limit
- interview_locked

These are persistent database fields, not runtime state.

History records retain the event-specific context:

- Screening: candidate_id + requirement_id
- Interview: candidate_id + requirement_id
- Selection: candidate_id + requirement_id

The Requirement identifies its Client through client_id.

The application uses atomic database operations for gates:

- screening: increment only when screening_count < screening_limit;
- interview: set interview_locked = true only when currently false;
- screening completion/rejection: decrement the screening count;
- interview rejection: release interview_locked;
- selection is unique per Candidate + Requirement.

Multi-step changes to Candidate state and history are committed in one database transaction.

PostgreSQL row locking and transactional atomicity provide the concurrency boundary; the application does not reconstruct these states by repeatedly counting history records.

## Consequences

### Positive

- Simpler data model.
- Fewer domain abstractions.
- Fewer state reconstruction queries.
- Persistent state survives process restarts and deployments.
- Candidate-level gates are easy to inspect and reason about.
- Concurrent recruiters cannot exceed screening capacity or acquire two interview locks for the same candidate.

### Transitional

Existing recruitment_assignments data and APIs are retained temporarily for historical compatibility while the earlier workflow slices are migrated. New recruitment-state APIs do not depend on Assignment.

### Non-goals

This decision does not turn Candidate master status into a recruitment outcome. Candidate rejection or selection remains contextual to a specific Requirement.

It also does not introduce interviewer, panel, hiring-manager, employee, payroll, or other enterprise HRMS concepts into the core model.
