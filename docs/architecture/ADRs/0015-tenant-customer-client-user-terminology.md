# ADR 0015 — Tenant, Customer, Client, and User Terminology

**Status:** Accepted  
**Date:** 2026-09-28  
**Decider:** Product owner  
**Related issue:** #79  
**Related ADRs:** ADR 0001, ADR 0002, ADR 0010, ADR 0014

## Context

Phase 9 introduces a SaaS control plane, subscriptions, tenant database provisioning, and login-time tenant/RBAC resolution. During the design of Phase 9, the terms **tenant**, **customer**, **client**, and **company** were being used too broadly and risked describing different business entities as though they were the same entity.

This distinction is fundamental because SkillSifter operates between two different business relationships:

1. **SkillSifter → recruitment/staffing firm**: the SaaS relationship.
2. **Recruitment/staffing firm → its recruitment clients**: the recruitment business relationship.

The existing architecture already establishes these as different concepts:

- ADR 0001 defines tenant identity as the security and isolation boundary.
- ADR 0002 defines a Client as an organization for which the recruitment firm performs recruitment work and explicitly places Client below Tenant.
- ADR 0010 defines SkillSifter as a recruitment/staffing platform for recruitment firms and their client-driven requirements.
- ADR 0014 defines the Phase 9 control plane around the subscribing company and its tenant database.

The terminology must therefore be frozen before physical tenant database provisioning and routing are implemented.

## Decision

### 1. Tenant

A **Tenant** is a recruitment/staffing company that subscribes to and operates its own account on SkillSifter.

The Tenant is the **SaaS customer** of SkillSifter.

The Tenant is the primary security and data-isolation boundary.

The authoritative tenant identity is an immutable tenant_id.

A tenant owns:

- users
- clients
- client contacts
- requirements
- candidates
- candidate × requirement recruitment transactions
- interviews
- selections
- offers
- joinings
- billing records
- reports and other tenant-owned operational data

A tenant receives one tenant data environment under the Phase 9 bridge tenancy model.

### 2. SaaS Customer

When the term **customer** is used in SkillSifter platform architecture, it means the **Tenant**, unless explicitly qualified otherwise.

Therefore:

> **SkillSifter customer = Tenant = subscribing recruitment/staffing company**

Technical architecture should prefer the term **Tenant** because it is unambiguous and directly expresses the security/data-isolation boundary.

### 3. User

A **User** is an individual who works for a Tenant and logs into SkillSifter.

In V1, a user belongs to one tenant.

A user's role is evaluated inside that tenant context.

Existing V1 roles remain:

- admin
- manager
- recruiter
- team_leader

Therefore:

~~~text
User
  ↓
Tenant
  ↓
RBAC
~~~

The User does not become a tenant merely because the user can administer the tenant.

### 4. Client

A **Client** is an organization for which the Tenant's recruitment/staffing firm performs recruitment work.

A Client is therefore **not** a SkillSifter tenant.

A Client is tenant-owned business data:

~~~text
Tenant
  └── Client
~~~

A Client may have:

- client contacts
- recruitment requirements
- recruitment activity
- candidates being considered against its requirements

The Client must never be used as the tenant security identity.

### 5. Requirement

A **Requirement** represents one Client's recruitment demand.

The ownership chain is:

~~~text
Tenant
  └── Client
        └── Requirement
~~~

The Requirement is therefore inside the Tenant's data boundary but belongs to a Client within that tenant.

### 6. Candidate

A Candidate is a candidate managed by the Tenant's recruitment operation.

A candidate may participate in multiple recruitment transactions:

~~~text
Tenant
  ├── Client A
  │     └── Requirement A1
  │           └── Candidate X
  │
  └── Client B
        └── Requirement B1
              └── Candidate X
~~~

The Candidate × Requirement relationship is the durable recruitment context, as established by ADR 0010.

## Canonical Business and Data Hierarchy

The following diagram is the canonical SkillSifter hierarchy:

~~~text
                         SKILLSIFTER
                       SaaS PLATFORM
                            │
                            ▼
                  ┌─────────────────────┐
                  │       TENANT        │
                  │                     │
                  │ SaaS Customer       │
                  │ Recruitment /       │
                  │ Staffing Firm       │
                  └──────────┬──────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
           Users          Clients       Candidates
              │              │
              │              ▼
              │       Client Contacts
              │              │
              │              ▼
              │        Requirements
              │              │
              │              ▼
              │    Candidate × Requirement
              │              │
              │              ▼
              │         Screening
              │              │
              │              ▼
              │         Submission
              │              │
              │              ▼
              │         Feedback
              │              │
              │              ▼
              │         Interview
              │              │
              │              ▼
              │         Selection
              │              │
              │              ▼
              │           Offer
              │              │
              │              ▼
              │           Joining
              │              │
              │              ▼
              │           Joined
              │              │
              │              ▼
              │           Billing
              │
              └────── RBAC inside Tenant
~~~

## Concrete Example

If R K Consulting subscribes to SkillSifter:

~~~text
SkillSifter
    │
    └── R K Consulting
          │
          │ Tenant / SaaS Customer
          │
          ├── Users
          │     ├── Admin
          │     ├── Manager
          │     └── Recruiters
          │
          ├── Samsung
          │     ├── Client Contact
          │     └── Requirements
          │
          ├── AMD
          │     └── Requirements
          │
          └── Other Clients
~~~

Here:

- **R K Consulting** is the Tenant and SaaS customer.
- **Users** work for R K Consulting.
- **Samsung** and **AMD** are Clients of R K Consulting.
- Samsung and AMD are not SkillSifter tenants.
- Their Requirements remain inside R K Consulting's tenant data boundary.

## SaaS Relationship vs Recruitment Relationship

These two relationships must remain separate.

### SaaS relationship

~~~text
SkillSifter
     │
     ▼
Tenant / SaaS Customer
     │
     ▼
Recruitment / Staffing Firm
~~~

### Recruitment relationship

~~~text
Recruitment / Staffing Firm
     │
     ▼
Client
     │
     ▼
Requirement
     │
     ▼
Candidate × Requirement
~~~

A Client must never become a separate SaaS tenant merely because the recruitment firm performs work for that Client.

## Phase 9 Control Plane Interpretation

The Phase 9 control plane operates at the **Tenant** level.

It represents:

~~~text
SkillSifter
    │
    └── Control Plane
           │
           ├── Tenant A
           │      ├── Subscription
           │      ├── Account status
           │      ├── Provisioning status
           │      └── Tenant DB routing
           │
           └── Tenant B
                  ├── Subscription
                  ├── Account status
                  ├── Provisioning status
                  └── Tenant DB routing
~~~

A tenant database belongs to exactly one Tenant.

It does **not** belong to one Client.

For example:

~~~text
Tenant DB: R K Consulting

    users
    clients
    client_contacts
    requirements
    candidates
    interviews
    selections
    offers
    joinings
    billing
    ...
~~~

The database may contain:

~~~text
Client: Samsung
Client: AMD
Client: Honeywell
Client: Other Client
~~~

All of those Clients remain within the R K Consulting tenant boundary.

## Authentication and Authorization Meaning

When a user logs in, the trusted platform context establishes:

~~~text
User
  ↓
Tenant
  ↓
Subscription / Account Status
  ↓
Tenant Database
  ↓
Role / RBAC
  ↓
Tenant Operation
~~~

The key questions are:

- **Tenant:** Which company's SkillSifter data is this user allowed to access?
- **RBAC:** What may this user do within that company's data?
- **Client:** Which organization is the recruitment firm performing work for?
- **Requirement:** What recruitment demand does that Client have?

These questions must not be collapsed into one identifier.

## Terminology Rules

The following terminology is mandatory for new architecture, code, API, schema, and documentation work:

| Term | Meaning | Security boundary |
|---|---|---|
| SkillSifter | SaaS product/platform | No |
| Tenant | Recruiting/staffing company subscribing to SkillSifter | **Yes** |
| SaaS Customer | Synonym for Tenant when discussing the commercial SaaS relationship | **Yes** |
| User | Person working for a Tenant and logging in | No; belongs to Tenant |
| Client | Organization for which the Tenant performs recruitment work | No; belongs to Tenant |
| Requirement | Recruitment demand belonging to a Client | No; belongs to Tenant through Client |
| Candidate | Candidate managed by the Tenant | No; belongs to Tenant |
| Candidate × Requirement | Recruitment transaction/context | No; belongs to Tenant |
| Company | Avoid as an ambiguous technical term; when referring to the SaaS account, use Tenant | — |

### Preferred terminology

Use:

- **Tenant** for SaaS isolation.
- **SaaS Customer** when discussing the commercial relationship with SkillSifter.
- **Client** for the recruitment firm's customer.
- **User** for a person operating the tenant account.

Avoid using **customer** without qualification when a technical entity is being described.

## Non-Goals

This ADR does not:

- create a new Client database model
- change the existing Client → Requirement relationship
- create a database for each Client
- make Clients SaaS tenants
- introduce multi-tenant users
- introduce cross-tenant administrators
- change the recruitment lifecycle
- change the existing RBAC roles
- introduce microservices

## Consequences

### Positive

- SaaS tenancy and recruitment-client relationships are unambiguous.
- Tenant databases have a single, clear ownership meaning.
- Client data cannot accidentally become a control-plane tenant.
- Login, subscription, database routing, and RBAC all operate at the correct boundary.
- Recruitment workflow remains inside the tenant while supporting multiple Clients.
- Future Phase 9 provisioning can be implemented without redefining Client semantics.

### Trade-offs

- The term "customer" must be qualified when necessary.
- Existing documentation and comments using "company", "customer", "client", and "tenant" interchangeably may require cleanup.
- Legacy compatibility fields such as company_name remain during the staged migration.

## Implementation Rule

Before Phase 9B physical tenant database provisioning begins:

1. Treat **Tenant** as the SaaS customer and isolation boundary.
2. Treat **Client** as a tenant-owned recruitment business entity.
3. Provision databases only for Tenants.
4. Never provision a database for a Client.
5. Resolve tenant identity from trusted authentication/control-plane context.
6. Resolve Client and Requirement only inside the authenticated Tenant context.
7. Keep existing recruitment functionality green throughout the migration.

This ADR is the terminology authority for Phase 9 and supersedes any informal use of these terms in implementation discussions where they conflict with this definition.
