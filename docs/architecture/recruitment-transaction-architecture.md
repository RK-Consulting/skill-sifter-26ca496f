# Recruitment Transaction Architecture

**Status:** Accepted current architecture
**Supersedes:** the earlier Assignment-centric transaction model

## 1. Core domain model

SkillSifter is agency-first. A Candidate and Requirement form the recruitment context directly.

Client → Requirement → Candidate × Requirement
                         ├── Screening history
                         ├── Submission history
                         ├── Interview history
                         └── Selection decision

There is no runtime Recruitment Assignment entity.

The Requirement identifies the Client through its client relationship. The Candidate remains a reusable master record and is never globally selected or rejected because of one requirement.

## 2. Persistent Candidate controls

Candidate contains the small amount of recruitment control state that must survive process restarts:
- screening_count
- screening_limit

Screening admission uses an atomic database update. Completing, rejecting, or withdrawing an active screening releases one slot.

## 3. Screening

Screening is scoped directly by candidate_id + requirement_id. It records recruiter evidence and remains historical. Screening does not create an Assignment.

## 4. Submission

Submission is scoped directly by Candidate × Requirement. At submission time the system stores immutable Candidate and Requirement snapshots so later profile changes do not rewrite historical submission facts.

## 5. Interview

Interview is scoped directly by Candidate × Requirement. A candidate may interview for multiple requirements concurrently. Only one active interview for the same Candidate × Requirement is allowed.

## 6. Selection

Selection is a decision on Candidate × Requirement: selected or rejected. There is one Selection decision per Candidate × Requirement.

Selection does not modify Candidate master status and does not transition an Assignment. A candidate can therefore be selected for one requirement while remaining active in other recruitment contexts.

## 7. Tenant isolation

Every Candidate × Requirement operation verifies that both records belong to the authenticated tenant. Tenant identity comes from authenticated request context and is never accepted from the request payload.

## 8. API boundary

Current V1 recruitment paths are:
- /candidates/{candidateId}/screenings
- /candidates/{candidateId}/interviews
- /candidates/{candidateId}/requirements/{requirementId}/submissions
- /candidates/{candidateId}/requirements/{requirementId}/selection

There are no current /assignments endpoints.

Historical Assignment migrations remain only so existing databases can migrate without losing historical data. Migration 030 removes Assignment foreign-key dependencies from Screening, Submission, and Selection.

## 9. Non-goals

The core model does not introduce interviewer, panel, hiring manager, employee, payroll, attendance, leave, or enterprise HRMS workflow concepts.

## 10. Subsequent lifecycle

Requirement → Candidate × Requirement → Screening → Submission → Feedback → Interview → Selection → Offer / Joining → BILLING