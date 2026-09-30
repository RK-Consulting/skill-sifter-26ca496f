# Changelog

All notable changes to SkillSifter are documented in this file.

## [1.0.0] - 2026-09-30

Production release.

### Added

- Complete Phase 9 SaaS platform foundation.
- Public tenant registration with initial Admin account.
- Control-plane tenant, account, subscription and provisioning state.
- Dedicated tenant PostgreSQL database provisioning.
- Tenant schema initialization and READY gating.
- Authenticated tenant database routing and request-scoped tenant DB access.
- Admin User Management UI and API.
- Server-side seat-limit enforcement.
- Fixed V1 privilege matrix across protected application domains.
- Subscription plan catalog and lifecycle persistence.
- External checkout handoff and provider webhook processing.
- Webhook signature validation and idempotency.
- Subscription activation, renewal, payment failure/suspension, cancellation, expiry and reactivation.
- Account and Subscription UI.
- Production Playwright smoke workflow.
- Permanent production smoke tenant and cleanup-safe mutation smoke.
- Phase 9 security, UAT and go-live documentation.

### Recruitment

- Completed Requirement → Candidate × Requirement → Screening → Submission → Feedback → Interview → Selection → Offer → Joining → JOINED → Billing workflow.
- Kept Offer intentionally simple: made + accepted.
- Kept Joining intentionally simple: joining date + joined.
- Added operational Billing worklist for joined recruitment records.
- Preserved recruiter-assisted AI/resume functionality without autonomous hiring decisions.
- Maintained recruitment history and auditability.

### Security

- Enforced trusted authenticated tenant identity as the tenant security boundary.
- Enforced tenant DB routing on the V1 API.
- Added cross-tenant negative-test coverage.
- Kept tenant ownership checks as defense in depth.
- Enforced server-side V1 role privileges.
- Protected Admin User Management and tenant-admin accounts.
- Enforced subscription state before protected tenant access.
- Production JWT configuration now fails closed when `JWT_SECRET` is missing.
- Production systemd environment explicitly identifies the production runtime.
- Removed the obsolete second deployment-time migration loop.
- Restored schema definition 039 required by the production database and kept schema checksum integrity intact.

### CI / Deployment

- Production deployment gate loads persistent production database credentials for PostgreSQL-backed tests.
- Deployment gate runs formatting, build, vet and tests before service restart.
- Production smoke workflow validates required credentials without printing secrets.
- Production deployment verifies Nginx configuration, systemd service health, DB connectivity and API health.
- Final production smoke passed all four Playwright tests in 19.1 seconds.

### Documentation / Architecture

- Consolidated Phase 9 architecture and go-live documentation.
- Updated V1 product scope and architecture baseline.
- Preserved the product boundary: SkillSifter recruitment intelligence; HRMS remains a separate future product.
- Kept the modular Go application architecture; no microservice split.
- Kept payment/accounting responsibilities outside SkillSifter.
- Kept the fixed V1 role model; no permission-builder framework.

### Scope discipline

- No HRMS functionality.
- No microservice split.
- No accounting or financial ledger.
- No configurable permission-builder framework.
- No unnecessary approval hierarchy.
- No autonomous recruiting agents.
- No automated external sourcing/scraping.
- No unnecessary infrastructure expansion.

## [0.5.6] - 2026-09-13

### Removed
- Assignments frontend module.
- Business Dev frontend module; its distinguishing fields moved to Client.

### Changed
- Client gained partner name and contact person fields.
- Navbar navigation was made independently scrollable.

## [0.5.5] - 2026-09-08

### Added
- ADR 0009 Docker backend architecture and deployment strategy.
- Audit event read access.

### Fixed
- Resume AI frontend response-envelope handling.

### Removed
- Unused skills and candidate_skills tables.

## [0.5.3] - 2026-09-02

### Added
- Combined CP11 Production Quality Gate and CP12 Recruitment Workflow UAT / Go-Live Readiness checkpoint.

### Changed
- Separate backend and frontend GitHub Actions quality gates retained.
- Final product-level recruiter UAT remains focused and manual.

## [0.4.0] - 2026-08-31

### Added
- Recruitment assignment state machine.
- Tenant-aware authorization and audit enforcement.
- Candidate technical and language expertise.
- GitHub Actions CI.

## [0.3.0] - 2026-08-23

### Added
- Resume ingestion/search foundation.
- V1 architecture and product-scope baseline.

## [0.2.0] - 2026-08-02

### Added
- RBAC and backend/frontend test infrastructure.
- Full local test gate.

## [0.1.0] - 2026-07-30

Initial documented baseline.
