# SkillSifter

**Multi-tenant recruitment intelligence platform for staffing and recruitment teams.**

[![Backend CI](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/backend-ci.yml)
[![Frontend CI](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/frontend-ci.yml/badge.svg)](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/frontend-ci.yml)
[![Lint](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/lint.yml/badge.svg?branch=main&label=lint)](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/lint.yml)
[![Release](https://img.shields.io/github/v/release/RK-Consulting/skill-sifter-26ca496f?display_name=tag)](https://github.com/RK-Consulting/skill-sifter-26ca496f/releases)
[![Version](https://img.shields.io/github/package-json/v/RK-Consulting/skill-sifter-26ca496f)](https://github.com/RK-Consulting/skill-sifter-26ca496f/releases/latest)
[![GitHub issues](https://img.shields.io/github/issues/RK-Consulting/skill-sifter-26ca496f)](https://github.com/RK-Consulting/skill-sifter-26ca496f/issues)

> **Current release:** v1.0.0 — Production release
> **Release date:** 2026-09-30
> **Product boundary:** SkillSifter is recruitment intelligence. HRMS is a separate future product.

SkillSifter is a recruitment intelligence and operations platform for staffing and recruitment teams. It combines a Go REST API, React + TypeScript frontend, PostgreSQL, tenant-isolated data, server-side RBAC, recruitment workflow management, AI-assisted resume capabilities, and SaaS subscription controls.

## Recruitment workflow

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

The recruitment lifecycle is deliberately simple and frozen for V1.

- **Offer:** made + accepted
- **Joining:** joining date + joined
- **Billing:** operational billing worklist after joined
- HRMS employee-management and approval workflows are outside SkillSifter.

## V1.0.0 capabilities

### Recruitment operations
- Dashboard and operational reporting
- Client management and client contacts
- Requirements management with recruitment-specific fields
- Candidate database and candidate expertise
- Candidate × Requirement recruitment context
- Assignments and daily recruitment tasks
- Screening
- Submission and client review
- Feedback
- Interview scheduling and interview workflow
- Selection
- Offer management
- Joining management
- Joined-candidate billing worklist
- Recruitment history and auditability

### Resume and AI-assisted recruitment
- Resume ingestion and resume intelligence foundation
- Resume extraction
- Recruiter-assisted candidate-to-requirement matching
- AI-assisted search/reporting surfaces
- AI remains assistive; recruiters make selection and submission decisions.

### SaaS platform
- Public tenant registration
- Initial tenant Admin creation
- Control-plane tenant/account management
- Subscription plan catalog
- External checkout handoff
- Provider webhook processing and signature validation
- Activation
- Renewal
- Payment failure/suspension
- Cancellation
- Expiry
- Reactivation/resumption
- Webhook idempotency
- Server-side subscription enforcement
- Account and Subscription UI

### Multi-tenancy and security
- Dedicated tenant PostgreSQL database per subscribed company
- Deterministic tenant DB provisioning
- Tenant schema initialization and migration
- READY provisioning gate
- Authenticated tenant DB routing
- Request-scoped tenant DB access
- Tenant ownership checks as defense in depth
- Cross-tenant negative-test coverage
- Control-plane / tenant-data separation
- JWT authentication
- Production fail-closed JWT secret handling
- Fixed V1 roles: Admin, Manager, Recruiter, Team Leader
- Server-side privilege matrix enforcement
- Admin User Management
- Server-side seat limits
- Tenant-admin protection
- Protected subscription access

### Administration
- Admin user list
- User creation and updates
- Permitted user deletion
- Role enforcement
- Seat usage and limits
- Tenant administrator protection

## SaaS subscription model

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
```

SkillSifter does not store payment instruments and does not implement an accounting or financial ledger. Formal invoicing, GST/tax processing, banking and client payment collection remain outside the product.

## Multi-tenancy architecture

```text
                         SkillSifter API
                              │
                    ┌─────────┴─────────┐
                    │                   │
               Control Plane       Authentication
                    │                   │
          Tenant / Subscription /       │
          Provisioning / Account        │
                    └─────────┬─────────┘
                              │
                       Tenant Resolver
                              │
                    ┌─────────┴─────────┐
                    │                   │
                 Tenant A DB         Tenant B DB
```

- Control DB: tenant registry, authentication/account state, subscription and provisioning state.
- Tenant DB: recruitment and operational data.
- The authenticated tenant identity determines the tenant DB.
- The client cannot select a tenant, tenant DB or role.
- Tenant isolation is enforced server-side; frontend visibility is never a security boundary.

## Authorization

Supported roles:

| Role | Scope |
|---|---|
| Admin | Tenant administration and permitted platform operations |
| Manager | V1 manager privileges |
| Recruiter | V1 recruiter privileges |
| Team Leader | V1 team-leader privileges |

The V1 privilege matrix is fixed. SkillSifter does not provide an arbitrary role designer or permission-builder framework.

## Technology

- **Backend:** Go, gorilla-mux, PostgreSQL, JWT, bcrypt
- **Frontend:** React 18, TypeScript, Vite, TanStack Query, Vitest
- **UI:** Tailwind CSS, shadcn/ui, Radix UI
- **Infrastructure:** Docker, Docker Compose, Nginx, systemd
- **CI/CD:** GitHub Actions
- **Browser verification:** Playwright production smoke tests

## Database and schema

The Go application owns the authoritative numbered schema definitions under `backend/database/migrations/`.

- `schema_versions` tracks applied definitions.
- Applied definitions are checksum-verified.
- New definitions are applied transactionally.
- PostgreSQL advisory locking serializes schema initialization.
- Tenant provisioning applies the tenant schema before the tenant becomes READY.
- There is no separate legacy deployment migration loop.

## Verification

Backend:

```bash
gofmt -l .
go build ./...
go vet ./...
go test ./...
```

Frontend:

```bash
npm ci
npm run lint
npm run test
npm run build
```

Production smoke:

```bash
cd e2e
npm ci
npm run smoke
```

Final production smoke for v1.0.0:

```text
4 passed (19.1s)
```

Release sequence:

```text
Formatting
   ↓
Build
   ↓
Vet
   ↓
PostgreSQL-backed Tests
   ↓
Security / Isolation
   ↓
UAT
   ↓
Production Deployment
   ↓
Production Smoke
   ↓
v1.0.0
```

## V1.0.0 release gate

- [x] Admin User Management
- [x] V1 privilege enforcement
- [x] Tenant isolation
- [x] Tenant DB provisioning
- [x] Tenant DB routing
- [x] Subscription lifecycle
- [x] Account & Subscription UI
- [x] Production deployment
- [x] Security hardening
- [x] Production smoke: 4/4 passed
- [x] Go-live release evidence

## Architecture boundary

V1.0.0 intentionally does **not** introduce:

- HRMS functionality
- microservices
- custom permission builders
- arbitrary RBAC
- accounting or financial ledger
- GST/accounting module
- banking/payment-instrument storage
- custom payment gateway
- unnecessary approval hierarchies
- autonomous recruiting agents
- LinkedIn/job-portal scraping
- WhatsApp automation
- candidate/client portals
- complex vector-database architecture
- advanced autonomous analytics/workflows

Post-v1.0.0 work is new product work and must be driven by real usage evidence.

## Documentation

- [Release Notes](RELEASE_NOTES.md)
- [Changelog](CHANGELOG.md)
- [Architecture](docs/architecture.md)
- [V1 Architecture Baseline](docs/architecture/v1-baseline.md)
- [V1 Product Scope](docs/product/v1-scope.md)
- [Phase 9 Implementation Status](docs/phase-9-implementation-status.md)
- [Phase 9 SaaS / Tenancy](docs/phase-9-saas-subscription-bridge-tenancy.md)
- [Phase 9 Security / UAT / Go-Live](docs/phase-9-security-uat-go-live.md)
- [Phase 9 Release Gate](docs/phase-9-release-gate.md)
- [Feature Reference](docs/features.md)
- [Production Topology & Recovery](docs/operations/production-topology.md)
- [API Specification](backend/docs/swagger.yaml)

## License

A project license has not yet been finalized. Until a license is added, the source should not be assumed to be available for unrestricted redistribution or commercial use.
