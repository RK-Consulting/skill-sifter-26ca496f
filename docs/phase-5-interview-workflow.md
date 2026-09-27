## Phase 5 — Agency-first Interview Workflow

Implement the Interview stage of the frozen SkillSifter recruitment lifecycle.

### Frozen product boundary

SkillSifter is a **recruitment intelligence platform for recruitment/staffing firms and their client requirements**. It is not an internal corporate HRMS.

The canonical lifecycle is:

```text
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

### Phase 5 scope

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

Department remains a **mandatory Requirement-level field**. It is not an Interview field.

A Candidate may participate in multiple requirements, including different departments for the same client. Rejection from one requirement must not globally reject the Candidate.

### Explicit non-goals

Do **not** add any of the following to the core Interview model:

- `interviewer_user_id`
- interviewer panel
- hiring manager
- client interview panel
- internal approval
- `submission_id` as an Interview dependency
- internal interview policy
- internal evaluation form

These belong to a separate future enterprise HRMS/extension domain.

### Architectural relationship

```text
Candidate
   │
   └── Recruitment Assignment ── Requirement
                                   │
                                   ├── Department
                                   └── Job ID
                                         │
                                         └── Interview
```

Interview is a recruitment event against the Candidate × Requirement transaction. It does not redefine Candidate master state.

### Acceptance criteria

- Interview creation is tenant-scoped.
- Interview references an existing Candidate and Requirement.
- Interview scheduling requires the Recruitment Assignment to already be in `interviewing` status; lifecycle transitions remain under the existing audited Assignment transition service.
- Requirement must belong to the same tenant.
- Candidate must belong to the same tenant.
- Interview round is represented explicitly.
- Interview scheduling is supported.
- Interview status and outcome are represented explicitly.
- Recruiter/client feedback can be recorded.
- Candidate feedback can be recorded.
- Next action can be recorded.
- Interview history is preserved.
- Existing Job ID traceability through Requirement is retained.
- No internal-HR concepts are required by the core Interview API or schema.
- Existing interview history is preserved.
- Existing CI remains green.
- Tests cover tenant isolation, validation, create/read/update/history behavior, and lifecycle integration.

### Subsequent phases

Phase 5 does not implement:

- Selection workflow
- Offer workflow
- Joining workflow
- Billing

Those remain subsequent lifecycle phases.

See ADR 0010 and `docs/architecture/recruitment-lifecycle.md`.
