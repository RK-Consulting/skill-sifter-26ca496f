# ADR 0001: Tenant Identity and Isolation

**Status:** Accepted  
**Decider:** Product owner  
**Related issue:** #16

## Context

The current shared-schema implementation has a `companies` table with an existing `id` primary key, but tenant-owned records and authenticated requests currently use mutable `company_name` values as the effective tenant identity. Individual handlers are responsible for applying tenant filters. V1 requires a reliable, immutable, testable tenant security boundary.

The existing implementation therefore contains the beginnings of a tenant identity model, but the technical tenant identifier is not consistently used across the application.

## Decision

SkillSifter V1 will use an **immutable tenant identifier** as the authoritative tenant identity and security boundary.

The existing `companies.id` will be evaluated and retained as the tenant identifier if its current semantics and data characteristics satisfy the migration requirements. A second competing tenant identity will not be introduced without an explicit architectural decision.

`company_name` will no longer serve as the authorization or isolation key. It remains business/display data and may change without changing tenant identity.

The migration from the current `company_name` model to tenant-ID-based isolation will be **staged and compatibility-aware**, rather than implemented as a big-bang schema rewrite.

## Target Identity Model

```text
Tenant
  ├── tenant_id   <- immutable technical identity
  └── name        <- mutable business/display name
        │
        ├── Users
        ├── Candidates
        ├── Jobs
        ├── Daily Jobs
        ├── Interviews
        └── Business Development
```

Tenant-owned records will ultimately reference `tenant_id`. The tenant name must never be used to determine authorization or cross-tenant access.

## Authentication and Request Context

The authenticated identity will ultimately carry:

- user identity
- tenant identity
- role

The tenant identity propagated from authentication is authoritative for tenant-scoped operations. Request handlers and data-access code must not accept a client-supplied tenant name or tenant identifier as an authorization override.

## Enforcement Model

Tenant isolation will be enforced as an application-wide security invariant:

1. Authentication establishes the authenticated tenant identity.
2. The tenant identity is propagated through the request context/service boundary.
3. Tenant-owned data access is scoped to that tenant identity.
4. Client-provided tenant identifiers or names cannot override the authenticated tenant.
5. Cross-tenant access must fail without leaking the existence of another tenant's resources.

The implementation must avoid relying solely on individual developers remembering to add tenant predicates to every query. V1 implementation work must establish consistent tenant-scoping conventions and an authoritative enforcement point.

## Migration Strategy

Migration will occur in stages:

### Stage A — Establish the authoritative tenant identity

Use the existing `companies.id` as the candidate immutable tenant identifier, subject to validation during implementation. Preserve `companies.name` as mutable business data.

### Stage B — Introduce tenant identity into tenant-owned records

Add and populate tenant identity for existing tenant-owned records, initially retaining `company_name` where necessary for compatibility during the transition.

The affected domain areas include users, candidates, jobs, daily jobs, interviews, and business development.

### Stage C — Migrate authentication

Move authenticated tenant context from `company_name` to the authoritative tenant identifier. JWT claims and request context should treat tenant ID as the security identity and company name as display/business data only.

### Stage D — Enforce tenant-scoped access consistently

Update service/data-access conventions and handlers so tenant-owned operations are always scoped to the authenticated tenant identity.

### Stage E — Retire `company_name` as an isolation key

After migration and validation, remove authorization/isolation dependencies on `company_name`. Retain it only where legitimately required as business/display information.

## Compatibility and Migration Risks

The implementation must account for:

- existing records identified only by `company_name`
- tenant names being renamed
- existing JWTs containing `companyName`
- API consumers that currently send or receive company names
- duplicate or inconsistent company-name data
- rollback during staged migration
- queries and handlers that currently rely on manual `company_name` predicates

Existing data must not become inaccessible or cross tenant boundaries during migration.

## Cross-Tenant Isolation Requirements

Cross-tenant isolation is a security requirement, not merely a functional test.

At minimum, tests must verify that a user authenticated for Tenant A cannot read, modify, or delete Tenant B resources, even when the resource ID is known.

The test matrix must cover every V1 tenant-owned resource as implementation work progresses.

A cross-tenant request must not disclose whether the requested resource exists in another tenant.

## Implementation Boundary

This ADR defines architecture only. It does **not** authorize schema changes, API changes, migration scripts, or handler refactoring as part of Issue #16.

Those changes must be implemented through separate, explicitly scoped issues derived from this decision.

## Consequences

### Positive

- Tenant identity is independent of mutable business names.
- Tenant renaming does not change security identity.
- Cross-tenant isolation becomes a first-class architectural invariant.
- Authentication and data access have a clear tenant boundary.
- Migration can be staged without an immediate destructive rewrite.

### Negative / Trade-offs

- The migration requires coordinated database, authentication, API, and data-access changes.
- Temporary compatibility fields increase implementation complexity during the transition.
- Existing handlers that manually filter by `company_name` must eventually be migrated.
- Cross-tenant security testing becomes a mandatory part of the V1 quality gate.
