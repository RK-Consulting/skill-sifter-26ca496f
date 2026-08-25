# Recruitment Transaction Architecture

**Status:** Accepted architectural baseline for v0.4 development  
**Related ADRs:** 0003, 0004, 0006  
**Primary issue:** #35

## 1. Purpose

This document consolidates the architectural decisions required to introduce the Recruitment Assignment transaction into SkillSifter.

The goal is to establish a durable relationship between a Candidate and a Client Requirement without conflating candidate profile state with recruitment transaction state, while preserving tenant isolation and historical traceability.

## 2. Core domain model

The v0.4 recruitment model is:

```text
Client
   |
   +--> Requirement
             |
             | N..1
             v
      Recruitment Assignment
             ^
             | 1..N
             |
         Candidate
```

A Recruitment Assignment is the transaction joining a candidate to a specific requirement.

### Cardinality

- One Candidate may have many Recruitment Assignments.
- One Requirement may have many Recruitment Assignments.
- A Candidate/Requirement pair represents one business transaction and must not have duplicate simultaneous assignment records.

The assignment is therefore the durable transaction boundary for recruitment activity.

## 3. Candidate state versus transaction state

These are independent state machines:

```text
Candidate status
       !=
Assignment lifecycle
```

A candidate can be `active` while having multiple assignments in different states, for example:

```text
Assignment A -> interviewing
Assignment B -> submitted
Assignment C -> rejected
```

Changing candidate status does not rewrite historical assignment state.

Candidate status vocabulary:

- `active`
- `inactive`
- `blacklisted`
- `archived`

See ADR 0004 for the authoritative candidate-state decision.

## 4. Recruitment Assignment lifecycle

The controlled lifecycle is:

```text
draft
  -> screening
  -> submitted
  -> interviewing
  -> offered
  -> joined
```

Terminal outcomes:

```text
rejected
withdrawn
joined
```

Ordinary API operations must validate transitions and reject invalid state changes. Terminal states cannot be moved backward through the normal workflow.

See ADR 0003 for the authoritative transition model.

## 5. Assignment ownership

Each assignment records:

- `created_by_user_id` — user who created the transaction.
- `owner_user_id` — user currently responsible for the transaction.
- `tenant_id` — authoritative tenant identity.

All ownership relationships are tenant constrained.

Tenant identity is derived from authenticated context and never from client-supplied authorization data.

## 6. Submission snapshots

At formal submission, the assignment captures immutable recruitment-facing snapshots of the candidate and requirement information necessary to explain what was submitted at that time.

The snapshot is historical evidence, not a replacement for the current Candidate or Requirement record.

Candidate snapshot information may include identity, skills, language expertise, and other recruitment-facing profile data. Requirement snapshot information may include title, location, work arrangement, description, required skills, experience, compensation, headcount, and language requirement.

The implementation must avoid copying unnecessary sensitive data into snapshots.

## 7. Downstream transaction ownership

Recruitment Assignment is the authoritative parent transaction for future recruitment activity:

```text
Recruitment Assignment
        |
        +--> Interviews
        +--> Offer
        +--> Joining
        +--> Future commercial records
```

Future records should reference the assignment rather than reconstructing candidate/requirement relationships independently.

### Legacy interview boundary

Existing `interviews` rows currently reference candidates and contain a free-text `position` value rather than a reliable requirement relationship. Existing interviews therefore must not be automatically remapped to assignments under #35.

Any historical interview migration requires a separate data-mapping and migration decision.

## 8. Audit architecture

New recruitment transaction audit coverage uses the audit-event model defined by ADR 0006.

Conceptual event structure:

```text
Audit Event
├── id
├── tenant_id
├── actor_user_id
├── entity_type
├── entity_id
├── action
├── occurred_at
├── correlation_id
└── metadata
```

Audit events are append-only, tenant-scoped, and must avoid full-row serialization of sensitive entities.

Recruitment Assignment lifecycle events include creation, screening start, submission, interview start, offer, joining, rejection, withdrawal, ownership changes, and submission snapshot creation.

The legacy `activity_logs` mechanism remains outside the scope of #35 and is not extended merely for convenience.

## 9. Tenant isolation

All recruitment transaction entities are tenant-owned.

The following relationships must remain inside one tenant:

```text
Assignment.tenant_id
Candidate.tenant_id
Requirement.tenant_id
Assignment.owner_user.tenant_id
```

Every API read/write must scope by authenticated `tenant_id`.

Cross-tenant candidate, requirement, owner, assignment, interview, or audit-event access must fail closed.

## 10. Legacy domain boundaries

The new Client/Requirement domain is additive to the legacy Jobs domain.

Existing `jobs` records remain authoritative for historical job transactions. `jobs.company_name` does not represent a client relationship and therefore cannot be safely converted into `requirements.client_id` automatically.

The following are explicitly outside the current recruitment transaction implementation:

- automatic `jobs -> requirements` migration
- synthetic clients for legacy jobs
- dual-write between jobs and requirements
- automatic historical interview remapping
- automatic historical daily-job remapping
- legacy `activity_logs` migration

These require separate decisions and migration plans.

## 11. Implementation sequence

The architecture is implemented in the following order:

```text
ADR 0004
Candidate state/language
        |
        v
ADR 0003
Recruitment Assignment
        |
        v
ADR 0006
Audit Events
        |
        v
Issue #35
Recruitment Assignment implementation
```

The three ADRs are the authoritative architectural decisions. This document is the consolidated implementation-oriented view and must remain consistent with those ADRs.

## 12. Non-goals

This architecture does not define:

- commercial billing rules
- offer compensation policy
- joining workflow details beyond assignment ownership
- client contact management
- historical migration of legacy recruitment transactions
- replacement of the legacy activity reporting system
- a universal language proficiency taxonomy beyond the current candidate-language direction

Those areas require separate domain or product decisions.
