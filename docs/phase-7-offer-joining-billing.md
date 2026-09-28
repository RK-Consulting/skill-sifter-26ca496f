# Phase 7 — Offer, Joining, and Billing Workflow

**Status:** Architecture baseline  
**Runtime model:** Candidate × Requirement  
**Prerequisite:** Phase 6 Selection

## Lifecycle

```
Selection
   ↓
Offer exists
   ↓
Offer accepted
   ↓
Joining date
   ↓
Joined
   ↓
Billing
```

## Offer

The Offer record itself means **offer made**.

Fields:
- Candidate × Requirement
- Selection reference
- `accepted: boolean`
- audit timestamps

Rules:
1. Selection must be `selected`.
2. One Offer per Candidate × Requirement.
3. `accepted = false` or `true`.
4. No Offer status state machine.

## Joining

Fields:
- Candidate × Requirement
- Offer reference
- `joining_date: date`
- `joined: boolean`
- audit timestamps

Rules:
1. Offer must be accepted.
2. One Joining per Candidate × Requirement.
3. `joined = true` requires joining_date.
4. No Joining status state machine.

## Billing

Billing is eligible only when:

```
Offer.accepted = true
        ↓
Joining.joined = true
        ↓
Billing eligible
```

Billing date = joining_date.

Billing remains recruitment-firm commercial functionality. It does not introduce accounting, payroll, GST/tax, or HRMS functionality.

## API

```
GET  /api/v1/candidates/{candidateId}/requirements/{requirementId}/offer
POST /api/v1/candidates/{candidateId}/requirements/{requirementId}/offer
PUT  /api/v1/candidates/{candidateId}/requirements/{requirementId}/offer

GET  /api/v1/candidates/{candidateId}/requirements/{requirementId}/joining
POST /api/v1/candidates/{candidateId}/requirements/{requirementId}/joining
PUT  /api/v1/candidates/{candidateId}/requirements/{requirementId}/joining

GET  /api/v1/candidates/{candidateId}/requirements/{requirementId}/billing
POST /api/v1/candidates/{candidateId}/requirements/{requirementId}/billing
```

## Acceptance criteria

- Offer cannot be created without selected Candidate × Requirement.
- Offer acceptance is a boolean.
- Joining cannot be created without accepted Offer.
- Joining has only joining_date and joined as business facts.
- joined=true requires joining_date.
- Billing cannot be created before joined=true.
- Candidate master state remains unchanged.
- Cross-tenant access is rejected.
- Historical Assignment migrations remain untouched.
- CI remains green.
