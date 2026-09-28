# ADR 0014 — Bridge Tenancy, Subscription Provisioning, and Login Context

**Status:** Accepted  
**Scope:** SaaS tenant architecture and account access  
**Related issue:** #79  
**Supersedes/extends:** ADR 0001 tenant identity and isolation, ADR 0005 V1 RBAC

## Context

SkillSifter is a multi-tenant SaaS product. The recruitment application must isolate each subscribing company's operational data while keeping infrastructure practical for a growing SaaS product.

The current implementation uses a shared PostgreSQL database with tenant identifiers on tenant-owned rows. ADR 0001 established immutable tenant identity and tenant-aware authentication, but did not select the final physical database tenancy model.

SkillSifter also needs an online subscription model. Subscription activation must establish a tenant environment, while payment processing and financial instrument data remain with an external payment provider.

The application must not become a payment ledger or financial-data store.

## Decision

SkillSifter will use a **bridge tenancy model**:

- A small **control-plane database** stores platform-level tenant/account metadata.
- Each subscribed company receives a dedicated **tenant PostgreSQL database**.
- Tenant databases initially share the same PostgreSQL infrastructure/cluster.
- Tenant recruitment and operational data is stored only in that tenant's database.
- The architecture must allow a future tenant to be moved to dedicated PostgreSQL infrastructure without changing application-level tenant contracts.
- Microservices are not required by this decision. The existing Go application remains the application boundary unless a later ADR explicitly approves service decomposition.

### Control plane

The control plane is authoritative for:

- tenant ID
- company name
- company email
- company phone
- company address
- subscription/plan
- subscription status
- subscription dates
- provisioning status
- tenant database identifier/routing metadata
- selected payment provider
- opaque external provider references required to communicate with that provider

The control plane must not store payment-card data, CVV, bank credentials, UPI credentials, payment instruments, or other sensitive financial information.

### Tenant data plane

A tenant database contains the operational SkillSifter data for exactly one subscribing company, including:

- users and roles
- candidates
- requirements
- clients
- recruitment lifecycle data
- reports/operational data
- other tenant-owned application records

A tenant database must never contain data belonging to another tenant.

## Subscription and provisioning flow

The successful SaaS subscription flow is:

~~~text
Customer registration
       ↓
Company/account created in control plane
       ↓
Payment provider checkout
       ↓
Provider confirms subscription/payment
       ↓
SkillSifter activates subscription
       ↓
Provision tenant PostgreSQL database
       ↓
Run current tenant migrations
       ↓
Create initial tenant admin
       ↓
Tenant becomes ACTIVE
       ↓
User can log in
~~~

Provisioning must be deterministic and idempotent. A retry must not create a second tenant database or duplicate the initial tenant administrator.

Payment-provider adapters remain replaceable. The control plane stores only the provider name and opaque references required for provider communication.

## Login and request context

Login is a control-plane operation first.

The authenticated session must establish:

1. user identity
2. tenant identity
3. subscription/account status
4. role/RBAC context
5. tenant database routing context

The protected application request then follows:

~~~text
Request
  ↓
Authentication
  ↓
Resolve tenant from trusted identity
  ↓
Check subscription/account access
  ↓
Resolve tenant database
  ↓
Establish tenant DB context
  ↓
RBAC/domain-action authorization
  ↓
Tenant-scoped operation
~~~

Client-supplied tenant IDs, company names, roles, or database identifiers cannot override trusted authentication/control-plane context.

A user with valid credentials but an inactive/expired/suspended subscription cannot enter normal tenant application workflows.

## RBAC

Existing V1 RBAC remains:

- admin
- manager
- recruiter
- team_leader

RBAC is always evaluated inside the authenticated tenant context.

Tenant identity and role authorization remain separate security boundaries:

- tenant identity answers **which company's data may be accessed**
- RBAC answers **what the authenticated user may do within that company**

Admin privileges never cross tenant boundaries.

## Payment boundary

SkillSifter does not process or store sensitive payment instruments.

External providers own:

- card/payment instrument details
- UPI/payment credentials
- bank/payment credentials
- financial transaction processing
- provider-side payment records

SkillSifter consumes normalized provider events such as:

- subscription activated
- payment succeeded
- payment failed
- subscription renewed
- subscription cancelled
- subscription expired

The payment provider is replaceable through an adapter/provider interface.

## Database lifecycle

Tenant database lifecycle is tied to subscription/account lifecycle, but subscription state and database lifecycle are separate concepts.

Examples:

- ACTIVE → tenant database available
- SUSPENDED/EXPIRED → application access blocked; tenant database retained according to the approved retention policy
- REACTIVATED → existing tenant database reused
- TERMINATED after retention policy → controlled tenant-data deletion

A subscription renewal must not create a new tenant database.

## Consequences

### Positive

- Strong physical database isolation between subscribing companies.
- Simple tenant routing model for application code.
- Central control plane provides a reliable place for subscription and tenant metadata.
- Payment providers can be changed without redesigning the subscription domain.
- Premium tenants can later be moved to dedicated infrastructure.
- Existing recruitment lifecycle remains independent of SaaS subscription/payment infrastructure.
- No financial instrument data is stored by SkillSifter.

### Trade-offs

- More operational work than a single shared database.
- Tenant database provisioning and migration automation become platform responsibilities.
- Backups, monitoring, connection management, and schema upgrades must account for multiple tenant databases.
- Cross-tenant reporting requires deliberate control-plane or reporting architecture rather than unrestricted joins across tenant databases.

## Explicit non-goals

This ADR does not introduce:

- microservices
- a custom payment gateway
- a financial ledger
- invoice/GST/accounting functionality
- a banking system
- a per-customer cloud server by default
- custom enterprise permission builders

## Implementation rule

Phase 9 implementation must proceed in small, reviewable increments:

1. control-plane tenant/subscription model
2. tenant database provisioning abstraction
3. tenant DB connection/routing
4. login-time subscription + tenant + RBAC resolution
5. provider adapter/webhook normalization
6. frontend account/subscription integration
7. migration, isolation, provisioning, and login tests

Existing recruitment functionality must remain green throughout the migration.