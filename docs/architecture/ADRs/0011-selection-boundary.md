# ADR 0011 — Agency-first Selection Boundary

**Status:** Accepted / Frozen
**Scope:** Core SkillSifter recruitment workflow

## Decision

Selection is an assignment-scoped recruitment decision.

The authoritative transaction remains:

`Candidate × Requirement = Recruitment Assignment`

Selection must never mutate Candidate master status.

The existing Assignment state machine remains the only authority for lifecycle transitions:

- `interviewing → offered` for a selected candidate
- `interviewing → rejected` for a rejected candidate

The Selection decision and the Assignment transition are committed atomically.

## Consequences

- A candidate can be selected for one requirement without being selected globally.
- Assignment lifecycle remains centralized and auditable.
- Selection history is retained independently from Candidate state.
- Offer, Joining, and Billing remain separate downstream workflow slices.
- Enterprise HRMS concepts do not enter the core recruitment model.
