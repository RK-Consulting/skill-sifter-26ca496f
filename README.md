# SkillSifter

**Multi-tenant recruitment intelligence platform for staffing and recruitment teams.**

[![Backend CI](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/backend-ci.yml)
[![Frontend CI](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/frontend-ci.yml/badge.svg)](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/frontend-ci.yml)

> **Current baseline:** Phase 9 implementation complete on main  
 > **Release state:** Final CI, security, UAT and production-smoke verification before go-live  
 > **Product boundary:** SkillSifter is recruitment intelligence. HRMS is a separate future product.

SkillSifter combines a Go REST API, React + TypeScript frontend, and PostgreSQL to manage recruitment in a tenant-isolated environment.

## Recruitment workflow

~~~text
Requirement
    ↓
Assignment
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
~~~

## Current capabilities

- Candidate management and expertise
- Client management
- Requirements with the defined recruitment fields
- Assignments and Daily Tasks
- Screening, Submission, Feedback and Interviews
- Selection, Offer and Joining
- Billing worklist
- Resume intelligence foundation
- Recruitment history and auditability
- JWT authentication and server-side RBAC
- Multi-tenant database provisioning and routing
- Admin User Management
- SaaS subscription plans, checkout and lifecycle
- Account and Subscription UI

### Offer and Joining

- Offer = made + accepted
- Joining = joining date + joined

### SaaS subscription

~~~text
Plan
 ↓
Checkout
 ↓
Provider Webhook
 ↓
Activation / Renewal
 ↓
Payment Failure / Cancellation / Expiry
~~~

SkillSifter does not implement an accounting or financial ledger.

## Multi-tenancy

Phase 9 separates platform control data from tenant-owned recruitment data.

- Control DB: authentication, tenant registry/provisioning state, account and subscription state
- Tenant DB: candidates, clients, requirements, assignments, recruitment workflow and billing data
- Tenant database provisioning and migrations
- Authenticated tenant database routing
- Tenant ownership checks as defense in depth
- Cross-tenant negative-test coverage
- READY gating before tenant login

## Authorization

Supported roles: Admin, Manager, Recruiter and Team Leader.

Authorization is enforced server-side. Frontend visibility is not a security boundary.

## Admin User Management

- List, create, update and permitted delete operations
- Seat usage and limits
- Tenant-admin protection
- Server-side authorization

## Architecture boundary

SkillSifter intentionally does not add HRMS functionality, microservices, an accounting ledger, a configurable permission-builder framework, or unnecessary approval hierarchies.

## Technology

- Go / gorilla-mux / PostgreSQL / JWT
- React / TypeScript / Vite / TanStack Query / Vitest
- Docker / Docker Compose / Nginx / GitHub Actions

## Docker

No Phase 9 Docker architecture change is required.

- PostgreSQL remains independent infrastructure.
- Backend remains containerized.
- Frontend remains containerized.
- Docker Compose remains the local development and verification stack.
- Tenant DB provisioning and routing remain backend responsibilities.

## Verification

Backend:
~~~bash
gofmt -l .
go build ./...
go vet ./...
go test ./...
~~~

Frontend:
~~~bash
npm ci
npm run lint
npm run test
npm run build
~~~

Combined local gate:
~~~bash
bash infra/scripts/test.sh
~~~

Release sequence:
~~~text
Formatting → Build → Vet → Test with PostgreSQL → Security → UAT → Production Smoke → GO-LIVE
~~~

## Phase 9 release gate

Implementation is consolidated on main. Final release evidence is tracked in GitHub Issue #99 and docs/phase-9-release-gate.md.

- [x] Admin User Management
- [x] Privilege enforcement
- [x] Tenant isolation implementation
- [x] Tenant DB provisioning
- [x] Tenant DB routing
- [x] Subscription lifecycle
- [x] Account & Subscription UI
- [ ] CI evidence on release commit
- [ ] Security negative-test execution
- [ ] Production UAT
- [ ] Production smoke test
- [ ] GO-LIVE

## Documentation

- CHANGELOG.md
- RELEASE_NOTES.md
- docs/features.md
- docs/phase-9-release-gate.md
- backend/docs/swagger.yaml

## Post-go-live

Once Phase 9 passes its release gate, freeze the architecture. New functionality becomes post-go-live product work and should be driven by real usage evidence.

## License

A project license has not yet been finalized. Until a license is added, the source should not be assumed to be available for unrestricted redistribution or commercial use.