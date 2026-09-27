# Phase 6 — Agency-first Selection Workflow

## Lifecycle position

```text
Requirement
     ↓
Candidate × Requirement
     ↓
Screening
     ↓
Submission
     ↓
Feedback
     ↓
Interview
     ↓
Selection
     ↓
Offer / Joining
     ↓
JOINED
     ↓
BILLING
```

## Boundary

Selection is a decision on **Candidate × Requirement**.

It is not a Candidate-master decision and it is not an Assignment-state transition.

A Candidate can therefore be selected for one Requirement while remaining active in other recruitment contexts.

## Selection record

The selection record contains:

- Candidate
- Requirement
- decision: `selected` or `rejected`
- decision notes
- next action
- decision timestamp
- `last_modified`

Exactly one selection decision is allowed for a Candidate × Requirement pair.

## Lifecycle integration

Selection requires a completed Interview for the same Candidate × Requirement pair.

The operation:

1. validates the authenticated tenant;
2. validates Candidate and Requirement ownership;
3. verifies a completed Interview exists for the pair;
4. verifies no prior Selection decision exists;
5. inserts the Selection decision.

Selection does not mutate Candidate master state and does not transition a Recruitment Assignment.

## API

### Record selection

`POST /api/v1/candidates/{candidateId}/requirements/{requirementId}/selection`

Request:

```json
{
  "decision": "selected",
  "decisionNotes": "Client confirmed selection.",
  "nextAction": "Prepare offer"
}
```

### Read selection

`GET /api/v1/candidates/{candidateId}/requirements/{requirementId}/selection`

Cross-tenant access behaves as not found.

## Explicit non-goals

Phase 6 does not implement:

- offer generation
- offer acceptance
- joining
- billing
- employee records
- payroll
- HRMS workflows
- candidate-global selection state
- interviewer/panel/hiring-manager concepts

## Acceptance criteria

- Tenant isolation is enforced.
- Selection is anchored to Candidate × Requirement.
- A completed Interview is required.
- Exactly one Selection decision is allowed per pair.
- Candidate master state remains unchanged.
- Offer/Joining remains a subsequent workflow.
- Existing Interview history remains unchanged.
