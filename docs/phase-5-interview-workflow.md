# Phase 5 — Agency-first Interview Workflow

## Frozen product boundary

SkillSifter is a recruitment intelligence platform for recruitment/staffing firms and their client requirements. It is not an internal corporate HRMS.

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

## Phase 5 scope

Implement an agency-neutral Interview workflow around Candidate + Requirement.

The core Interview record supports:

- candidate
- requirement
- interview round
- scheduled time
- status
- outcome
- feedback
- candidate feedback
- next action
- audit timestamps

Department remains a mandatory Requirement-level field. It is not an Interview field.

A Candidate may participate in multiple requirements, including different departments for the same client.

## Runtime relationship

```text
Candidate ───────── Candidate × Requirement ───────── Requirement
                                                       │
                                                       ├── Department
                                                       └── Job ID
                                                             │
                                                             └── Interview
```

Historical Recruitment Assignment records are not required by the runtime Interview workflow.

## Interview gates

Scheduling an Interview requires:

1. a tenant-valid Candidate;
2. a tenant-valid Requirement;
3. a completed Screening for the same Candidate × Requirement;
4. an existing Submission for the same Candidate × Requirement;
5. recorded client feedback for that recruitment context;
6. no other active Interview for the same Candidate × Requirement.

A candidate may have active interviews for different requirements.

## Explicit non-goals

Do not add any of the following to the core Interview model:

- `interviewer_user_id`
- interviewer panel
- hiring manager
- client interview panel
- internal approval
- `submission_id` as an Interview dependency
- internal interview policy
- internal evaluation form

## API

```text
GET  /api/v1/interviews
POST /api/v1/interviews
GET  /api/v1/interviews/{id}
PUT  /api/v1/interviews/{id}

GET  /api/v1/candidates/{candidateId}/interviews
```

## Acceptance criteria

- Interview creation is tenant-scoped.
- Candidate and Requirement references are validated.
- Interview round is explicit.
- Interview status and outcome are explicit.
- Recruiter/client feedback can be recorded.
- Candidate feedback can be recorded.
- Next action can be recorded.
- Interview history is preserved.
- Job ID traceability remains through Requirement.
- No internal-HR concepts are required by the core Interview API or schema.
- Tests cover tenant isolation, validation, create/read/update/history behavior, and lifecycle gates.
