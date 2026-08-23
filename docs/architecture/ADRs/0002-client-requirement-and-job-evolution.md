# ADR 0002: Client, Requirement, and Existing Job Evolution

**Status:** Accepted  
**Decider:** Product owner  
**Related issue:** #17

## Context

The current `jobs` resource is a generic job record and does not model the recruitment-agency relationship between a client and its recruitment demand. The current schema stores `company_name` on jobs and exposes fields such as title, department, location, status, description, requirements, and posting dates, but there is no explicit client/contact/requirement domain. fileciteturn28file0

The V1 architecture requires explicit `clients`, `client_contacts`, and `requirements` and places Requirement after Client in the core workflow. Candidate-to-requirement work is intentionally separated into Recruitment Assignment. fileciteturn34file0

## Decision

SkillSifter V1 will introduce **Client**, **Client Contact**, and **Requirement** as explicit domain concepts.

### Client

A Client represents an organization for which the recruitment firm performs recruitment work.

Ownership:

```text
Tenant
  └── Client
```

A client is tenant-owned. A client must have a stable technical identity and must not be used as the tenant security identity.

Lifecycle:

```text
Prospect → Active → Inactive
```

Lead and opportunity qualification remain part of Business Development. Conversion into a Client is a separate business transition and is not implicitly performed by creating a requirement.

Core attributes are expected to include:

- client identity
- legal/business name
- status
- primary business/contact metadata where applicable
- created/updated timestamps
- tenant identity

The exact physical schema is deferred to implementation issues.

### Client Contact

A Client Contact represents an individual contact within a Client organization.

Relationship:

```text
Client 1 ──── * ClientContact
```

A contact belongs to exactly one client. Core attributes are expected to include:

- contact identity
- client identity
- name
- designation/title where available
- email
- phone/contact number
- status
- created/updated timestamps

Multiple contacts are supported because recruitment operations may involve HR, hiring managers, coordinators, and commercial contacts.

### Requirement

A Requirement is the authoritative representation of one client's recruitment demand/JD.

Relationship:

```text
Client 1 ──── * Requirement
```

A requirement belongs to exactly one client and must be tenant-scoped through the client/tenant relationship.

Core requirement attributes are expected to include:

- requirement identity
- client identity
- role/title
- department/function where applicable
- location/work arrangement where applicable
- lifecycle/status
- opened/created date
- description/JD content
- required qualifications/skills
- experience expectations
- compensation information where provided
- opening/headcount information where applicable
- language requirements where applicable, using the generic language + proficiency framework defined by the V1 architecture
- last-modified/audit timestamps

These are domain-level requirements, not a final database schema. Physical fields are to be finalized by the implementation issue.

## Requirement Lifecycle

The V1 requirement lifecycle is:

```text
Draft → Open → On Hold → Filled
                 │
                 └────→ Cancelled
```

The implementation may represent equivalent states with approved naming conventions, but the lifecycle must distinguish an active requirement from one that is paused, fulfilled, or cancelled.

A requirement's lifecycle must not be confused with the lifecycle of a candidate's Recruitment Assignment. Recruitment Assignment is defined separately in Issue #18.

## Jobs-to-Requirements Decision

`jobs` will **not** remain the long-term V1 domain model.

A new explicit `requirements` model will be introduced during V0.5. The existing `jobs` model will be retained temporarily as a compatibility/legacy model while existing data and consumers are migrated.

The intended direction is:

```text
Current
Client intent → generic Job

V1
Client → Requirement
```

Existing job records will be migrated into requirements using an explicit field mapping. Legacy APIs may remain temporarily where required to avoid an uncontrolled breaking change, but new domain work must target `requirements`, not extend `jobs` as a parallel long-term model.

## Candidate Relationship Boundary

A Requirement does not directly own candidates.

The candidate-to-requirement recruitment transaction belongs to `Recruitment Assignment`, which is the central V1 transaction connecting one candidate to one client requirement. This preserves the ability for one candidate to be considered for multiple requirements and preserves transaction-specific history. fileciteturn34file0

Issue #18 will define that transaction in detail.

## V0.5 Migration Strategy

Migration will be staged:

1. Introduce the Client, Client Contact, and Requirement domain model.
2. Establish client relationships for existing operational data where the current data permits reliable mapping.
3. Map existing `jobs` records into requirements.
4. Provide controlled compatibility for legacy job APIs during migration where necessary.
5. Migrate frontend/service consumers from jobs to requirements.
6. Validate data and tenant ownership.
7. Retire the legacy `jobs` model and compatibility paths through a separate approved implementation issue.

No destructive replacement or uncontrolled rename is authorized as part of this architecture decision.

## Historical and Audit Requirements

Requirement changes must not rewrite historical recruitment transaction facts. Recruitment Assignment, submission, interview, offer, joining, and commercial records must be able to preserve the requirement context that was relevant at the time of the transaction. The exact snapshot strategy is defined by the Recruitment Assignment and downstream architecture issues.

Client and requirement changes must also produce auditable activity where required by the V1 audit architecture.

## Tenant Isolation

Client, Client Contact, and Requirement are tenant-owned data. Their names, statuses, or other business attributes must never be treated as tenant security identity.

All access must be scoped to the authenticated tenant identity established by ADR 0001.

Cross-tenant access is a security defect.

## Consequences

### Positive

- Recruitment operations gain an explicit client-to-requirement domain model.
- Requirements represent recruitment demand rather than generic jobs.
- Client contacts can support real recruitment workflows without overloading the client entity.
- Candidate-to-requirement transactions remain separate and auditable.
- Existing job data can be migrated without a big-bang rewrite.

### Trade-offs

- V0.5 temporarily carries legacy `jobs` compatibility.
- Data migration requires explicit mapping and validation.
- Existing job APIs and UI consumers must eventually be migrated.
- Client and requirement domain modeling introduces additional entities compared with the current simple jobs model.

## Implementation Boundary

This ADR defines architecture only. It does **not** authorize database schema changes, migration scripts, API rewrites, or frontend changes under Issue #17.

Those changes must be implemented through separately scoped GitHub issues after the v0.4 architecture foundation is complete.
