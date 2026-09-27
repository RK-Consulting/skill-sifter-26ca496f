# SkillSifter Recruitment Lifecycle

**Status:** Frozen architectural/product baseline  
**Related ADR:** ADR 0010  
**Primary Phase:** Phase 5 and subsequent recruitment workflow phases

## Product definition

SkillSifter is a **recruitment intelligence platform** for recruitment/staffing firms and their client-driven hiring requirements.

It is deliberately **not** an internal corporate HRMS.

## Canonical lifecycle

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

## Requirement

A Requirement represents a client's recruitment demand.

Department is mandatory Requirement data because the same client may have separate requirements for different departments. Department therefore contributes to requirement context and prevents unrelated requirements from being collapsed into one generic job record.

Typical Requirement data includes:

- Client
- Department
- Job Title
- Job Type
- Experience Required
- Budget
- Language Requirements
- Certifications Required
- Notice Period
- Mode of Work
- Mandatory Requirements
- Job Description
- Status
- Job Location
- Number of Open Positions
- Job ID where applicable

## Assignment

A Recruitment Assignment is the durable Candidate × Requirement transaction.

This is critical because Candidate state is global while recruitment state is contextual.

Example:

```text
Candidate A
   ├── Client X / Finance Requirement → Rejected
   └── Client X / Technology Requirement → Interested / Interviewing
```

A rejection in one requirement must not globally reject the candidate.

## Screening

Screening records recruiter-side evidence and assessment for a specific assignment.

## Submission

Submission records the formal presentation of a candidate against a requirement when the recruitment workflow requires it.

The Interview domain does not depend on `submission_id`. Interview is anchored directly to Candidate + Requirement and remains usable for the core recruitment workflow.

## Feedback

Feedback records client-side recruitment feedback associated with the recruitment transaction. Feedback does not mutate Candidate master data.

## Interview

Interview records the recruitment interview event.

The core model is intentionally agency-neutral:

```text
Interview
├── Candidate
├── Requirement
├── Round
├── Scheduled At
├── Duration
├── Status
├── Outcome
├── Feedback
├── Candidate Feedback
├── Next Action
└── Audit timestamps
```

The model does not require the identity of a client interviewer, panel, hiring manager, internal approver, or enterprise evaluation policy.

## Selection / Offer / Joining

These are subsequent recruitment lifecycle stages.

Selection means the client has selected the candidate. Offer/Joining captures the transition toward actual employment with the client.

The candidate becomes `joined` only when the business workflow records the joining event.

## Billing

Billing is downstream of joining.

The intended business rule is:

> A recruitment transaction becomes billing-eligible when the candidate actually joins, according to the firm's billing policy.

Billing is therefore associated with the relevant Recruitment Assignment / Requirement / Client transaction, never with the Candidate globally.

## Product boundary

### In scope

- Recruitment/staffing workflow
- Candidate management
- Requirements
- Department-aware recruitment demand
- Recruitment Assignments
- Screening
- Submission
- Client feedback
- Interviews
- Selection
- Offer / Joining
- Billing
- Recruitment intelligence AI

### Explicitly out of core scope

- Employee HRMS
- Attendance
- Leave
- Payroll
- Employee benefits
- Performance management
- Appraisals
- Corporate HR policy engines
- Internal approval hierarchies
- Client interview panels
- Internal hiring-manager workflows
- Enterprise interview evaluation forms
- HRMS-specific AI

A future enterprise HRMS must be treated as a separate product/domain rather than an expansion of the SkillSifter core recruitment model.
