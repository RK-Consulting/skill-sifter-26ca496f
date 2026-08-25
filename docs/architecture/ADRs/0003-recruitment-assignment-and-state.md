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

### 7. Downstream transaction ownership

Recruitment Assignment is the authoritative parent transaction for downstream recruitment activity.

Conceptually:

```text
Candidate
   |
   v
Recruitment Assignment
   |
   +--> Interviews
   +--> Offer
   +--> Joining
   +--> Future commercial records
```

Future domain records should reference the assignment rather than independently reconstructing a candidate/requirement relationship.

Existing `interviews` records currently reference candidates and use a free-text position field. Existing interview data must not be remapped automatically under this ADR because there is no reliable historical requirement relationship. Any historical interview migration is a separate migration decision.

### 8. Tenant isolation

Every assignment is tenant-owned and must enforce that candidate, requirement, and assignment owner belong to the same tenant.

All reads and writes are scoped by the authenticated `tenant_id`. Client-supplied tenant identifiers are ignored for authorization purposes.

## Consequences

### Positive

- Establishes a durable recruitment transaction boundary.
- Supports multiple requirements per candidate and multiple candidates per requirement.
- Separates candidate state from transaction state.
- Provides a stable parent for interviews, offers, joining, and future commercial records.
- Preserves historical submission context through immutable snapshots.
- Provides a clear tenant-isolation boundary.

### Trade-offs

- Assignment lifecycle requires explicit transition validation.
- Submission snapshots introduce immutable historical data that must be maintained separately from current profiles.
- Existing interviews cannot be cleanly migrated without additional historical mapping decisions.
- A future re-submission model may require an explicit assignment versioning/re-attempt design.

## Implementation Boundary

This ADR defines the architecture for Issue #35. It does **not** authorize automatic migration of historical `jobs` or `interviews` data, dual-write behavior between legacy `jobs` and requirements, synthetic clients, or changes to existing recruitment transaction history.
