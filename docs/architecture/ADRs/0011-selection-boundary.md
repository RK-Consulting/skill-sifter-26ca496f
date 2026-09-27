# ADR 0011 — Agency-first Selection Boundary

**Status:** Accepted / Frozen  
**Scope:** Core SkillSifter recruitment workflow

## Decision

Selection is a **Candidate × Requirement** recruitment decision.

The Candidate × Requirement pair is the authoritative runtime context for the decision. Recruitment Assignment was the earlier implementation model and is no longer a runtime dependency.

Selection must never mutate Candidate master status.

Selection requires a completed interview for the same Candidate × Requirement pair. Exactly one Selection decision is allowed for that pair.

The Selection record contains:

- `candidate_id`
- `requirement_id`
- `decision` (`selected` or `rejected`)
- decision notes
- next action
- decision timestamp
- last modification timestamp

## API boundary

```text
GET  /api/v1/candidates/{candidateId}/requirements/{requirementId}/selection
POST /api/v1/candidates/{candidateId}/requirements/{requirementId}/selection
```

## Consequences

- A candidate can be selected for one requirement without being selected globally.
- Selection history is retained independently from Candidate state.
- Selection does not transition or mutate a Recruitment Assignment.
- Offer, Joining, and Billing remain separate downstream workflow slices.
- Enterprise HRMS concepts do not enter the core recruitment model.
