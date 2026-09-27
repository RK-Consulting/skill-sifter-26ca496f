# ADR 0012: Persistent Candidate Recruitment State

## Status

Accepted

## Context

The recruitment workflow was previously modeled around a first-class recruitment_assignments entity with service-layer lifecycle transitions.

That model introduced unnecessary indirection for the agency-first workflow. The business rules that must be enforced are simpler:

- a candidate may be screened for multiple client requirements at the same time;
- simultaneous screening capacity is bounded by a persistent counter and limit;
- a candidate may interview for multiple requirements concurrently;
- only one active interview is allowed for the same Candidate × Requirement pair;
- screening, interview, submission, and selection records are recruitment history;
- Candidate + Requirement already identifies the recruitment context.

## Decision

The current recruitment control state is stored directly on the Candidate row:

- screening_count
- screening_limit

These are persistent database fields, not runtime state.

History records retain the event-specific context:

- Screening: candidate_id + requirement_id
- Interview: candidate_id + requirement_id
- Selection: candidate_id + requirement_id

The Requirement identifies its Client through client_id.

The application uses database-backed gates:

- screening: increment only when `screening_count < screening_limit`;
- interview: allow only one active interview for the same Candidate × Requirement pair;
- screening completion/rejection/withdrawal: decrement the screening count when an active screening is released;
- selection: one decision per Candidate + Requirement.

Interview concurrency is enforced by the Candidate × Requirement relationship; there is no global candidate interview lock.

Multi-step changes to Candidate state and history are committed in one database transaction where state and history must change together.

PostgreSQL row locking and transactional atomicity provide the concurrency boundary; the application does not reconstruct these states by repeatedly counting history records.

## Consequences

### Positive

- Simpler data model.
- Fewer domain abstractions.
- Fewer state reconstruction queries.
- Persistent screening capacity survives process restarts and deployments.
- Candidate-level screening capacity is easy to inspect and reason about.
- Interview concurrency is isolated correctly to Candidate × Requirement, so a candidate can progress through multiple requirements concurrently.

### Transitional

Historical `recruitment_assignments` data and the earlier migrations are retained for database upgrade compatibility. The current runtime does not expose Assignment APIs and the new screening, submission, interview, feedback, and selection paths do not depend on Assignment.

### Non-goals

This decision does not turn Candidate master status into a recruitment outcome. Candidate rejection or selection remains contextual to a specific Requirement.

It also does not introduce interviewer, panel, hiring-manager, employee, payroll, or other enterprise HRMS concepts into the core model.
