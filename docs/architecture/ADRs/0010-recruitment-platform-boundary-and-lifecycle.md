# ADR 0010 — SkillSifter Recruitment Platform Boundary and Lifecycle

**Status:** Accepted / Frozen  
**Date:** 2026-09-27  
**Scope:** Core SkillSifter product architecture

## Context

SkillSifter is a recruitment/staffing platform for recruitment firms and their client-driven hiring requirements. It is not an internal corporate HRMS.

A corporate HRMS has a different domain model, including employee lifecycle management, organisation-specific HR policies, attendance, leave, payroll, performance management, internal approvals, and policy-driven workflows. Its AI requirements can also be substantially different because AI may need to reason over organisation-specific policies and employee processes.

Introducing those concerns into SkillSifter would broaden the core model without serving the primary recruitment workflow.

## Decision

SkillSifter is explicitly defined as a **recruitment intelligence platform**.

The canonical core recruitment lifecycle is:

```text
RECRUITMENT
───────────
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

The durable business transaction is the Candidate × Requirement Recruitment Assignment.

Department is a mandatory Requirement-level concept. It identifies the organisational context of the client requirement and supports correct requirement identity, duplicate avoidance, and reuse of a candidate across different department-specific requirements.

A candidate's outcome is always scoped to the relevant recruitment transaction. Rejection from one requirement or department does not globally reject the candidate.

## Core Interview boundary

The core Interview model represents an interview event for a Candidate against a Requirement. It does not model the client's internal interview organisation.

Core Interview data may include:

- candidate
- requirement
- interview round
- scheduled time
- status
- outcome
- recruiter/client feedback
- candidate feedback
- next action
- audit timestamps

The following are explicitly **not part of the core SkillSifter recruitment model**:

- `interviewer_user_id`
- interviewer panels
- hiring managers
- client interview panels
- internal approval workflows
- `submission_id` as an Interview requirement
- internal interview policies
- internal evaluation forms

These may belong to a separate future enterprise HRMS or an explicitly approved extension, but they must not be introduced into the core recruitment schema merely for enterprise-HR compatibility.

## AI boundary

Core SkillSifter AI is recruitment intelligence. It should support recruitment activities such as:

- resume understanding and extraction
- candidate-to-requirement matching
- skill and requirement interpretation
- screening assistance
- duplicate/relationship intelligence
- recruiter productivity

Corporate HR policy reasoning, employee-policy automation, internal appraisal intelligence, payroll intelligence, and similar HRMS AI concerns are outside the current product boundary.

## Commercial boundary

Billing is part of the SkillSifter recruitment business lifecycle, but it follows successful joining.

Selection alone does not make a recruitment transaction billable. The intended business sequence is:

```text
Selection
   ↓
Offer / Joining
   ↓
Candidate joins
   ↓
JOINED
   ↓
Billing eligibility
   ↓
Invoice / billing record
```

Billing belongs to the recruitment transaction, not to the Candidate globally.

## Consequences

1. Core SkillSifter remains focused on recruitment/staffing.
2. Candidate state is not confused with Candidate × Requirement transaction state.
3. A candidate can participate in multiple department-specific requirements simultaneously.
4. Internal corporate HR concepts cannot silently enter the core schema.
5. Future enterprise HRMS work can be designed independently with its own policy, workflow, security, and AI architecture.
6. Phase 5 must implement only the agency-neutral Interview workflow.
7. Selection/Offer/Joining and Billing remain subsequent phases and are not folded into Phase 5.

## Frozen rule

Any future feature proposed for SkillSifter must first answer:

> Does this directly support the recruitment/staffing lifecycle?

If not, it requires a separate approved product/domain decision rather than being added to the SkillSifter core model.
