# SkillSifter — V1.0.0 Architecture

**Status:** Frozen for v1.0.0 production release  
**Release:** v1.0.0  
**Date:** 2026-09-30

## 1. Purpose

This document is the current runtime architecture reference for SkillSifter v1.0.0.

SkillSifter is a **multi-tenant recruitment intelligence platform** for staffing and recruitment teams. It is deliberately implemented as a modular application rather than a microservice estate.

The product boundary is:

> **SkillSifter = recruitment intelligence and recruitment operations. HRMS is a separate future product.**

This document supersedes older descriptions of shared-schema/company-name tenancy. Those descriptions are historical only.

## 2. System architecture

```text
                         INTERNET
                            │
                           HTTPS
                            │
                     Nginx / Frontend
                            │
                       React / Vite
                            │
                        REST / JSON
                            │
                     Go Modular API
                            │
              ┌─────────────┴─────────────┐
              │                           │
        Control Plane                 Tenant Resolver
              │                           │
     tenant / account /             authenticated
     subscription / plan /          tenant context
     provisioning state                  │
              │                           │
              └─────────────┬─────────────┘
                            │
                  Request-scoped Tenant DB
                            │
                    PostgreSQL cluster
                            │
                 ┌──────────┴──────────┐
                 │                     │
              Tenant A DB           Tenant B DB
```

### Components

- **Frontend:** React 18, TypeScript, Vite, TanStack Query, Tailwind CSS, shadcn/ui/Radix.
- **Backend:** Go REST API using gorilla-mux and PostgreSQL.
- **Control plane:** platform tenant/account/subscription/provisioning state.
- **Tenant database:** tenant-owned recruitment and operational data.
- **Infrastructure:** Nginx, systemd, Docker/Docker Compose and PostgreSQL.
- **CI/CD:** GitHub Actions.
- **Production browser verification:** Playwright.

The application remains a modular Go monolith and a single React SPA.

## 3. Control plane

The control plane owns platform-level state:

- tenant identity
- company/account information
- subscription plan and status
- subscription dates
- provisioning state
- tenant database routing metadata
- platform user/account identity
- opaque payment-provider references

Payment instruments and banking credentials are not stored.

## 4. Tenant database boundary

Each subscribed company receives a dedicated PostgreSQL database on the shared PostgreSQL infrastructure.

Tenant-owned data includes:

- candidates
- clients and contacts
- requirements
- candidate × requirement recruitment context
- screening
- submissions
- feedback
- interviews
- selections
- offers
- joinings
- billing worklist
- recruitment history
- tenant operational data

The authenticated tenant determines the database. The client cannot select a tenant ID, tenant database or role.

## 5. Authentication and authorization

Authentication uses JWT with bcrypt password verification.

The request context is established in this order:

```text
Credentials
   ↓
Authenticate user
   ↓
Resolve trusted tenant
   ↓
Check subscription/access state
   ↓
Resolve tenant DB
   ↓
Resolve fixed V1 role
   ↓
Apply server-side privilege rules
   ↓
Handler
```

V1 roles are fixed:

- Admin
- Manager
- Recruiter
- Team Leader

Frontend route visibility is not a security boundary.

## 6. Tenant isolation

Tenant isolation is enforced through:

1. trusted authenticated tenant identity;
2. tenant DB resolution;
3. request-scoped tenant DB access;
4. tenant ownership checks where applicable;
5. cross-tenant negative tests.

The V1 API explicitly applies both authentication and tenant DB routing middleware.

Control-plane queries remain on the control DB. Tenant-owned requests use the resolved tenant DB.

## 7. Subscription lifecycle

```text
Plan
 ↓
Checkout
 ↓
Provider Webhook
 ↓
Activation
 ↓
Renewal
 ↓
Payment Failure / Suspension
 ↓
Cancellation / Expiry
 ↓
Reactivation
```

Provider webhooks are authenticated and idempotent. Subscription state is enforced server-side before protected application access.

SkillSifter does not implement accounting, GST, banking, payment-instrument storage or a financial ledger.

## 8. Recruitment architecture

The V1 recruitment workflow is:

```text
Requirement
    ↓
Candidate × Requirement
    ↓
Screening
    ↓
Submission
    ↓
Feedback
    ↓
Interview
    ↓
Selection
    ↓
Offer
    ↓
Joining
    ↓
JOINED
    ↓
BILLING
```

The Candidate × Requirement relationship is the durable recruitment context.

### Offer

Offer state is intentionally limited to:

- offer made
- accepted

### Joining

Joining state is intentionally limited to:

- joining date
- joined

Billing is operational recruitment billing and is allowed after joined.

No HRMS workflow is introduced into this domain.

## 9. Schema architecture

The Go application owns the authoritative numbered schema definitions in:

The v1.0.0 runtime schema is split into two authoritative database planes:

- `backend/database/control-plane/001_baseline.sql`
- `backend/database/tenant-plane/001_baseline.sql`

The former `backend/database/migrations/` evolution chain is retired and is preserved only in Git history.

The schema engine:

- tracks applied definitions in `schema_versions`;
- verifies checksums for already-applied definitions;
- applies new definitions transactionally;
- uses PostgreSQL advisory locking;
- initializes tenant schemas during provisioning.

There is no second legacy deployment migration loop.

Schema definition 039 is part of the production schema history and must remain available for checksum/version verification.

## 10. Production security

v1.0.0 includes:

- production fail-closed JWT secret configuration;
- explicit production systemd environment;
- tenant DB routing on V1 API;
- fixed-role server-side authorization;
- tenant-admin protection;
- seat-limit enforcement;
- subscription state enforcement;
- webhook signature validation;
- webhook idempotency;
- cross-tenant denial tests;
- deployment preflight and health checks.

## 11. Release architecture boundary

The following are intentionally outside v1.0.0:

- microservices
- arbitrary RBAC
- permission-builder framework
- HRMS
- accounting/financial ledger
- GST/accounting system
- banking/payment-instrument storage
- custom payment gateway
- autonomous recruiting agents
- automated external sourcing/scraping
- candidate/client portals
- complex vector-database architecture
- advanced autonomous analytics

The architecture is frozen for the production release. Post-release changes require a new product/architecture decision.

## 12. Authoritative companion documents

- `docs/architecture/v1-baseline.md`
- `docs/product/v1-scope.md`
- `docs/authorization/v1-privilege-matrix.md`
- `docs/phase-9-saas-subscription-bridge-tenancy.md`
- `docs/phase-9-implementation-status.md`
- `docs/phase-9-security-uat-go-live.md`
- `docs/phase-9-release-gate.md`
- Architecture Decision Records under `docs/architecture/ADRs/`
