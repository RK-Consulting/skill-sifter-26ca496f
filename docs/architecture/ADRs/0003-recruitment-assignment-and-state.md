# ADR 0003: Recruitment Assignment and Transaction State

**Status:** Accepted  
**Decider:** Product owner  
**Related issues:** V04-04, V04-06, Issue #35

## Context

V1 must connect a candidate and a client requirement through a durable transaction. Candidate state and recruitment transaction state are separate concerns.

The existing schema has no durable candidate-to-requirement relationship. The Recruitment Assignment therefore becomes the first-class transaction connecting a candidate to a specific requirement and, through the requirement, to a client.

## Decision

### 1. Recruitment Assignment is a first-class domain entity

Introduce a tenant-scoped `recruitment_assignments` entity connecting candidates and requirements.

The relationship is many-to-many through the assignment entity:

- One candidate may have many assignments.
- One requirement may have many candidate assignments.
- A candidate and requirement pair has one active business transaction represented by one assignment record.

The assignment is therefore the durable transaction boundary for recruitment activity.

### 2. Candidate/requirement uniqueness

A candidate must not have multiple simultaneous assignment records for the same requirement. The implementation should enforce uniqueness for the `(candidate_id, requirement_id)` pair.

A completed or terminated assignment is historical. Reopening the same recruitment need is not performed by silently resetting a terminal assignment; any future re-attempt requires an explicit product decision/versioning mechanism.

### 3. Assignment ownership

Each assignment records both:

- `created_by_user_id` — the user who created the transaction.
- `owner_user_id` — the user currently responsible for the transaction.

Both users must belong to the assignment tenant. Tenant identity is authoritative from authenticated request context and is never accepted from client-supplied payload data.

### 4. Assignment lifecycle

The controlled lifecycle is:

```text
draft
  -> screening
  -> submitted
  -> interviewing
  -> offered
  -> joined
```

Terminal outcomes are:

```text
rejected
withdrawn
joined
```

Valid transitions include:

- `draft -> screening`
- `screening -> submitted`
- `submitted -> interviewing`
- `interviewing -> offered`
- `offered -> joined`
- `screening -> rejected`
- `submitted -> rejected`
- `interviewing -> rejected`
- `offered -> rejected`
- `screening -> withdrawn`
- `submitted -> withdrawn`
- `interviewing -> withdrawn`
- `offered -> withdrawn`

Terminal states must not be moved backward through ordinary API operations. The implementation must validate state transitions rather than accepting arbitrary status values.

### 5. Candidate state is independent

Assignment lifecycle never overwrites candidate status. A candidate may participate in multiple assignments while retaining one independent candidate-level status.

Candidate eligibility for a new assignment may be restricted by candidate status, but existing assignment history remains valid.

### 6. Historical snapshots

At formal submission, the assignment records immutable recruitment-facing snapshots of the candidate and requirement information necessary to explain what was submitted at that time.

Snapshots are historical evidence and do not replace the current candidate or requirement records.

At minimum, the submission snapshot must preserve the relevant candidate identity/profile, skills/language information, and requirement information such as title, location, work arrangement, description, required skills, experience, compensation, headcount, and language requirement.

### 7. Assignment boundary versus Job ID master-index flow

Recruitment Assignment remains the authoritative transaction for the candidate-to-requirement recruitment workflow and its submission lifecycle. It is not the master index for Interview, Billing, or Reports.

The Job ID master-index flow is:

```text
Requirement
   |
   | Job ID
   v
Interview
   |
   v
Billing
   |
   v
Reports
```

Rules:

- Job ID is entered on Requirement and is unique within the tenant.
- Interview setup selects the Requirement/Job ID.
- Interview stores the Requirement reference and derives Job ID from the Requirement.
- Billing and recruitment-transaction reports use the Requirement/Job ID relationship.
- Candidate Submission / Recruitment Assignment does not need Job ID for this flow.
- Client does not reference Job ID.

Existing interviews may have no Requirement reference because they predate the Job ID model. They must not be automatically remapped without reliable historical evidence. New and updated Interview workflows require a Requirement with a Job ID.

### 8. Tenant isolation

Every assignment is tenant-owned and must enforce that candidate, requirement, and assignment owner belong to the same tenant.

All reads and writes are scoped by the authenticated `tenant_id`. Client-supplied tenant identifiers are ignored for authorization purposes.

## Consequences

### Positive

- Establishes a durable recruitment transaction boundary.
- Supports multiple requirements per candidate and multiple candidates per requirement.
- Separates candidate state from transaction state.
- Provides a stable transaction boundary for candidate recruitment workflow state and submission history.
- Preserves historical submission context through immutable snapshots.
- Provides a clear tenant-isolation boundary.

### Trade-offs

- Assignment lifecycle requires explicit transition validation.
- Submission snapshots introduce immutable historical data that must be maintained separately from current profiles.
- Existing interviews cannot be cleanly migrated without additional historical mapping decisions.
- A future re-submission model may require an explicit assignment versioning/re-attempt design.

## Implementation Boundary

This ADR defines the architecture for Issue #35. At the time of approval, it explicitly prevented automatic migration of historical legacy Jobs records and dual-write behavior while the Requirement domain was being introduced.

That historical constraint has since been resolved by retiring the legacy Jobs resource. The current system uses Requirements as the authoritative recruitment-demand model. Historical migration SQL and ADR context remain preserved for traceability; they are not instructions to recreate or dual-write a Jobs resource.

Existing interviews still require separate historical mapping decisions where a reliable Requirement relationship cannot be established.
