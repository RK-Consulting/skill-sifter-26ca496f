# Phase 6 — Agency-first Selection Workflow

Phase 6 implements the Selection stage of the frozen recruitment lifecycle.

## Lifecycle position

```
Requirement
     ↓
Assignment
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

Selection is a decision on the **Candidate × Requirement Recruitment Assignment**.

It is not a Candidate-master decision.

A Candidate can therefore be selected for one Requirement while remaining active in other recruitment assignments.

## Selection record

The selection record is assignment-scoped and contains:

- Recruitment Assignment
- decision: `selected` or `rejected`
- decision notes / client feedback
- next action
- decision timestamp
- `last_modified`

There is no required Selection `created_at` field.

Exactly one selection decision is allowed for an assignment because the existing Assignment state machine makes the decision terminal with respect to the Selection stage.

## Lifecycle integration

Selection does not directly update `recruitment_assignments.status`.

The Selection operation runs inside one database transaction:

1. Validate the authenticated tenant.
2. Confirm that no Selection decision already exists.
3. Use the existing audited Assignment transition service:
   - `selected` → `offered`
   - `rejected` → `rejected`
4. Insert the Selection decision.
5. Commit both operations atomically.

If the Assignment is not `interviewing`, the existing state-machine transition rules reject the operation.

## API

### Record selection

`POST /api/v1/assignments/{id}/selection`

Request:

```json
{
  "decision": "selected",
  "decisionNotes": "Client confirmed selection.",
  "nextAction": "Prepare offer"
}
```

### Read selection

`GET /api/v1/assignments/{id}/selection`

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
- direct Assignment status mutation outside the existing transition service

## Acceptance criteria

- Tenant isolation is enforced.
- Selection is anchored to an existing Assignment.
- Only `interviewing` assignments can transition through Selection.
- Selected transitions to `offered` through the existing audited transition path.
- Rejected transitions to `rejected` through the existing audited transition path.
- Selection history is retained.
- Candidate master state is unchanged.
- Duplicate Selection decisions are rejected.
- Existing Interview history remains unchanged.
