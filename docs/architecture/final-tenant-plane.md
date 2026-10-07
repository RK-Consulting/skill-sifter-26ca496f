# Final Tenant Plane Architecture

## 1. Purpose

The tenant plane is the operational database for exactly one SkillSifter customer.

It contains recruitment data only. It has no dependency on the control-plane database for referential integrity and no cross-database foreign keys.

The tenant database is created from the current tenant baseline when a tenant is provisioned.

## 2. Tenant boundary

One tenant database represents one customer:

`tenant_id -> tenant database -> recruitment data`

The database itself is the primary isolation boundary.

`tenant_id` may remain on operational rows as immutable identity/audit metadata, but it is NOT a foreign key to the control plane.

The application obtains the tenant context from authenticated routing and must never accept a caller-supplied tenant identity as authority.

## 3. Tenant-owned domains

The final tenant plane contains only these domains:

### Identity and authorization
- `users`
- `roles`

These are tenant-local application accounts and authorization data.

They are not the SaaS control-plane account registry.

### Recruitment master data
- `clients`
- `client contacts/partner information` where required by the existing product model
- `requirements`

Requirements are the sole recruitment-demand model. There is no Jobs domain.

### Candidate intelligence
- `candidates`
- candidate language expertise
- candidate technical expertise
- resume/profile intelligence and related provenance
- employment, education, certification and project intelligence where already part of Resume AI

AI-derived data remains tenant data. AI assists recruiters; it does not own recruitment decisions.

### Recruitment execution
- recruitment execution records (screening, submission, feedback, interview, selection, offer, joining, billing)
- screening state/history
- submissions
- client feedback
- interviews
- selection decisions
- offers
- joining
- recruitment billing records

The workflow is:

`Requirement -> Assignment -> Screening -> Submission -> Feedback -> Interview -> Selection -> Offer -> Joining -> JOINED -> Billing`

## 4. Explicitly excluded

The tenant plane does not contain:

- platform tenants
- platform plans
- platform subscriptions
- checkout records
- payment-provider events
- registration registry
- pending SaaS registrations
- SaaS verification codes
- provisioning state
- tenant database routing
- subscription eligibility
- payment instruments or financial transaction records
- `companies` as a separate customer-root table

The control plane owns those concerns.

## 5. Tenant user model

`users` is tenant-local.

A user row contains the credentials and role needed by the tenant application. It does not reference `platform_tenants` with a PostgreSQL foreign key.

The control plane may retain a routing/access record for the platform account, but synchronization between the two planes is explicit application behavior.

There is no cross-database FK.

## 6. Tenant schema integrity

PostgreSQL is responsible for physical integrity:

- primary keys
- NOT NULL
- uniqueness
- foreign keys between tenant-local tables
- simple CHECK constraints
- positive counts
- valid enumerated values
- basic date/value validity

Go is responsible for domain decisions:

- authorization
- tenant routing
- workflow transitions
- eligibility
- subscription access
- screening limits
- selection
- offer acceptance
- joining rules
- billing eligibility
- audit event creation
- conflict interpretation

A database constraint must prevent impossible storage, not silently implement a business workflow.

## 7. Cross-entity tenant consistency

Because the database is physically tenant-specific, cross-entity relationships are tenant-local.

Where `tenant_id` is retained on rows, Go writes the authenticated tenant identity and verifies it.

Do not create FKs from tenant tables to control-plane tables.

Do not create triggers that reach across domains to enforce workflow behavior.

## 8. Recruitment workflow state

The application owns workflow state.

The database may constrain the allowed state values, but state transitions are performed explicitly by Go.

Important race-sensitive operations must use an atomic PostgreSQL operation and return a deterministic result to Go:

- screening-limit consumption
- submission
- feedback update
- interview outcome
- selection
- offer creation/acceptance
- joining
- billing creation

Two simultaneous requests must never produce two logically incompatible successful outcomes.

## 9. Offer and joining

Keep these deliberately small.

### Offer

`candidate_id + requirement_id + selection_id + accepted`

Offer creation is the existence of an offer.

Acceptance is a boolean.

### Joining

`candidate_id + requirement_id + offer_id + joining_date + joined`

No employee/HRMS model is introduced.

If `joined = true`, Go requires a joining date.

## 10. Billing

Tenant billing is recruitment billing only.

It records the commercial recruitment event needed by the recruitment firm.

It is NOT:

- SaaS subscription billing
- payment processing
- accounting
- tax
- employee payroll
- invoice generation by SkillSifter

SaaS subscription/payment lifecycle belongs exclusively to the control plane and external payment provider.

## 11. Audit

Recruitment audit history is tenant-owned.

Audit records are written explicitly by Go at successful domain operations.

There are no generic INSERT/UPDATE/DELETE triggers generating business activity.

Audit rows are append-only by application contract.

## 12. Triggers

The final tenant baseline contains no generic business triggers.

Triggers must not:

- generate recruitment activity
- advance workflow state
- create related recruitment records
- enforce authorization
- enforce tenant routing
- implement subscription behavior

If a future trigger is ever required, it must be justified as a physical database invariant and documented separately.

## 13. Legacy objects excluded from the final baseline

The following historical/shared objects are not part of the final tenant baseline:

- `companies` customer-root usage
- `jobs`
- `daily_jobs` legacy runtime model
- legacy activity-log trigger infrastructure
- `platform_tenants` references
- control-plane subscription tables
- shared-database compatibility structures
- migration-time backfill logic
- cross-plane foreign keys
- runtime schema cleanup/drop logic

Existing production migration history remains historical evidence. It is not the definition of a newly provisioned tenant database.

## 14. Final migration architecture

The repository will maintain two independent schema roots:

```
backend/database/control-plane/
backend/database/tenant-plane/
```

Each has its own baseline and version history.

A new tenant receives the current tenant baseline.

A control database receives the current control-plane baseline.

There is no numeric relationship such as:

`tenant migration <= 034`

and no migration cutoff that determines which old migrations belong to which database.

## 15. Provisioning invariant

Provisioning succeeds only when:

1. the control plane has created the tenant identity;
2. a tenant database has been created;
3. the current tenant baseline has been applied successfully;
4. the tenant database identity has been recorded by the control plane;
5. routing can resolve the tenant to that database;
6. tenant-local authentication can operate.

If any step fails, Go records a deterministic provisioning failure state and the operation can be retried safely.

## 16. Routing invariant

Every authenticated tenant request follows:

```
request
  -> authenticate
  -> resolve tenant
  -> control-plane routing lookup
  -> open/use tenant DB
  -> execute tenant-domain operation
```

The tenant database connection is selected by server-side routing state.

A request cannot choose an arbitrary database by supplying a tenant ID.

## 17. Final invariant

> The tenant plane knows everything required to operate one customer's recruitment business. It does not know how SaaS registration, subscription, payment, provisioning, or other customers are managed.

## 18. Acceptance criteria

The tenant-plane architecture is accepted only when:

- a tenant database can be created without the control database schema;
- `companies` is absent as a customer-root dependency;
- no tenant table has an FK to `platform_tenants`;
- no tenant table has an FK to any control-plane table;
- Requirements are the only demand model;
- Jobs are absent;
- recruitment workflow is represented by tenant-local data;
- Offer is accepted boolean;
- Joining is joining date + joined boolean;
- recruitment billing is tenant-local;
- audit is explicit application behavior;
- generic business triggers are absent;
- all workflow races have atomic DB operations;
- tenant migrations are independently versioned;
- a new tenant can be provisioned from the current baseline without replaying historical compatibility migrations.

## Decision

**Freeze the tenant-plane architecture.**

The next implementation step is to construct the tenant-plane baseline from the final domain model and then audit the Go repositories/handlers against that physical boundary.
