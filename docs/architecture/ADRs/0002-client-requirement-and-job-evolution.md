# ADR 0002: Client, Requirement, and Existing Job Evolution

**Status:** Accepted  
**Decider:** Product owner  
**Related issue:** #17

## Context

The current `requirements` resource is a generic job record and does not model the recruitment-agency relationship between a client and its recruitment demand. The current schema stores `company_name` on requirements and exposes fields such as title, department, location, status, description, requirements, and posting dates, but there is no explicit client/contact/requirement domain. fileciteturn28file0

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

The current Requirement lifecycle is:

```text
Open → On Hold → Closed
 │
 └────→ Cancelled
```

The current implementation uses the status values `open`, `on_hold`, `closed`, and `cancelled`. `open` is the database default for newly created Requirements.

A Requirement is the authoritative representation of one client's recruitment demand. Its recruitment-demand attributes include Job Type, Job Title, Department, Experience Required, Budget, Language Requirements, Certifications Required, Notice Period, Work Arrangement, Mandatory Requirements, Job Description, Job Location, and Open Positions.
## Candidate Relationship Boundary

A Requirement does not directly own candidates.

The candidate-to-requirement recruitment transaction belongs to `Recruitment Assignment`, which is the central V1 transaction connecting one candidate to one client requirement. This preserves the ability for one candidate to be considered for multiple requirements and preserves transaction-specific history. fileciteturn34file0

Issue #18 will define that transaction in detail.

## V0.5 Migration Strategy

Migration will be staged:

1. Introduce the Client, Client Contact, and Requirement domain model.
2. Establish client relationships for existing operational data where the current data permits reliable mapping.
3. Map existing `requirements` records into requirements.
4. Provide controlled compatibility for legacy job APIs during migration where necessary.
5. Migrate frontend/service consumers from requirements to requirements.
6. Validate data and tenant ownership.
7. Retire the legacy `requirements` model and compatibility paths through a separate approved implementation issue.

No destructive replacement or uncontrolled rename is authorized as part of this architecture decision.

### Implementation status (Issue #34)

Issue #34 implemented stage 1 only: the Client and Requirement domain model
(`clients` and `requirements` tables, tenant-scoped, with full CRUD and
tenant isolation). Stages 2-7 above are **not** implemented and were not
attempted.

**Legacy `requirements` remains authoritative for existing recruitment
transactions. Client/Requirement is a new, independent v0.4 domain.**
Migration of historical `requirements` records into `requirements` (stage 3) is
intentionally deferred, not merely unstarted: `requirements.company_name`
identifies the *recruiting tenant itself*, not a client of that tenant, so
it does not encode the client relationship `requirements.client_id`
requires. There is no reliable existing data from which to derive real
client identities for historical job records. Treat this as an open
architectural question requiring an explicit decision — including whether
synthetic per-tenant clients would be created, how `interviews` and
`daily_requirements` (which currently reference `requirements`, not `requirements`) would
be affected, and what happens to `requirements.created_by_user_id` and other
fields with no `requirements` equivalent — not as a default follow-up task
to "just migrate the requirements."

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
- Requirements represent recruitment demand rather than generic requirements.
- Client contacts can support real recruitment workflows without overloading the client entity.
- Candidate-to-requirement transactions remain separate and auditable.
- Existing job data can be migrated without a big-bang rewrite.

### Trade-offs

- V0.5 temporarily carries legacy `requirements` compatibility.
- Data migration requires explicit mapping and validation.
- Existing job APIs and UI consumers must eventually be migrated.
- Client and requirement domain modeling introduces additional entities compared with the current simple requirements model.

## Implementation Boundary

This ADR defines architecture only. It does **not** authorize database schema changes, migration scripts, API rewrites, or frontend changes under Issue #17.

Those changes must be implemented through separately scoped GitHub issues after the v0.4 architecture foundation is complete.## Legacy Jobs Decision and Current Implementation

At the time this ADR was written, SkillSifter had a legacy Jobs resource and the architecture considered a staged migration from that resource to a Client → Requirement model. That historical decision is retained here for traceability.

The implementation has since completed that migration boundary:

- Requirements are now the single authoritative recruitment-demand model.
- The legacy Jobs API, handlers, service, model, routes, and UI have been removed.
- The legacy `jobs` table is retired by the authoritative schema definitions.
- No legacy Jobs data is required or retained by the current Requirement implementation.
- Historical migration SQL and comments documenting the Jobs retirement remain part of the schema history.
- **Daily Jobs** remains a separate operational domain and is not part of the retired Jobs resource.

Therefore, the staged migration language below is historical context, not a current implementation plan.
## Candidate Relationship Boundary

A Requirement does not directly own candidates.

The candidate-to-requirement recruitment transaction belongs to `Recruitment Assignment`, which is the central V1 transaction connecting one candidate to one client requirement. This preserves the ability for one candidate to be considered for multiple requirements and preserves transaction-specific history. fileciteturn34file0

Issue #18 will define that transaction in detail.

## V0.5 Migration Strategy

Migration will be staged:

1. Introduce the Client, Client Contact, and Requirement domain model.
2. Establish client relationships for existing operational data where the current data permits reliable mapping.
3. Map existing `requirements` records into requirements.
4. Provide controlled compatibility for legacy job APIs during migration where necessary.
5. Migrate frontend/service consumers from requirements to requirements.
6. Validate data and tenant ownership.
7. Retire the legacy `requirements` model and compatibility paths through a separate approved implementation issue.

No destructive replacement or uncontrolled rename is authorized as part of this architecture decision.

### Implementation status (Issue #34)

Issue #34 implemented stage 1 only: the Client and Requirement domain model
(`clients` and `requirements` tables, tenant-scoped, with full CRUD and
tenant isolation). Stages 2-7 above are **not** implemented and were not
attempted.

**Legacy `requirements` remains authoritative for existing recruitment
transactions. Client/Requirement is a new, independent v0.4 domain.**
Migration of historical `requirements` records into `requirements` (stage 3) is
intentionally deferred, not merely unstarted: `requirements.company_name`
identifies the *recruiting tenant itself*, not a client of that tenant, so
it does not encode the client relationship `requirements.client_id`
requires. There is no reliable existing data from which to derive real
client identities for historical job records. Treat this as an open
architectural question requiring an explicit decision — including whether
synthetic per-tenant clients would be created, how `interviews` and
`daily_requirements` (which currently reference `requirements`, not `requirements`) would
be affected, and what happens to `requirements.created_by_user_id` and other
fields with no `requirements` equivalent — not as a default follow-up task
to "just migrate the requirements."

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
- Requirements represent recruitment demand rather than generic requirements.
- Client contacts can support real recruitment workflows without overloading the client entity.
- Candidate-to-requirement transactions remain separate and auditable.
- Existing job data can be migrated without a big-bang rewrite.

### Trade-offs

- V0.5 temporarily carries legacy `requirements` compatibility.
- Data migration requires explicit mapping and validation.
- Existing job APIs and UI consumers must eventually be migrated.
- Client and requirement domain modeling introduces additional entities compared with the current simple requirements model.

## Implementation Boundary

This ADR defines architecture only. It does **not** authorize database schema changes, migration scripts, API rewrites, or frontend changes under Issue #17.

Those changes must be implemented through separately scoped GitHub issues after the v0.4 architecture foundation is complete.
