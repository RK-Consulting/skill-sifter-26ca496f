# SkillSifter — Phase 9 Release Notes

## Phase 9 — SaaS Foundation, Security Hardening and Go-Live Readiness

Phase 9 completes the core SaaS architecture required to operate SkillSifter as a multi-tenant recruitment platform.

Implementation is consolidated on main. Final production activation remains subject to GitHub Issue #99.

## Highlights

### Admin User Management
- User listing, creation, updates and permitted deletion
- Seat usage and limits
- Tenant-admin protection
- Server-side authorization

### Privilege Enforcement
- V1 privilege matrix enforced across protected application domains.
- Frontend visibility is not treated as authorization.

### Tenant Isolation
- Authenticated tenant identity
- Tenant database routing
- Tenant ownership checks
- Control-plane / tenant-data separation
- Cross-tenant negative-test coverage

### Tenant Database Provisioning
- Deterministic tenant database identity
- Tenant schema migrations
- Tenant seed data
- Provisioning status
- READY gating
- Administrative retry

### Subscription Lifecycle

Plan → Checkout → Provider Webhook → Activation → Renewal → Payment Failure → Cancellation / Expiry.

Implemented: plans, checkout, webhook processing, activation, renewal, payment failure, cancellation, expiry, idempotency and server-side enforcement.

### Account & Subscription UI
- Account information
- Current plan/status
- Available plans
- Checkout handoff
- Admin-only cancellation

No accounting or financial ledger was introduced.

## Recruitment boundary

Requirement → Assignment → Screening → Submission → Feedback → Interview → Selection → Offer → Joining → JOINED → Billing.

Offer is made + accepted. Joining is joining date + joined.

HRMS-specific approval and employee-management functionality remains outside SkillSifter.

## Docker

No Docker architecture change is required for Phase 9. PostgreSQL remains independent; backend and frontend remain containerized; Docker Compose remains the local development/verification stack.

## CI and release gate

Formatting → Build → Vet → Test with PostgreSQL → Security Tests → UAT → Production Smoke → GO-LIVE.

Authoritative checklist: GitHub Issue #99 and docs/phase-9-release-gate.md.

## Scope discipline

No HRMS, microservices, accounting ledger, configurable permission builder, additional approval hierarchy, unnecessary recruitment entities or containerized PostgreSQL were introduced.

## Final release state

Implementation: complete on main.

Production: pending final evidence from CI, security/UAT and production smoke.

The project should only be declared GO-LIVE after all release-gate checks are verified.