# Changelog

All notable changes to SkillSifter are documented in this file.

## [Unreleased] — Phase 9 SaaS / Go-Live

Phase 9 implementation is consolidated into main. Final production activation remains subject to the release gate.

### Added

- Tenant database provisioning and tenant schema migration support.
- Tenant provisioning status and READY gating.
- Tenant database routing for tenant-owned operations.
- Admin User Management UI and API integration.
- Server-side seat-limit enforcement.
- Subscription plans and lifecycle persistence.
- Checkout handoff and provider webhook processing.
- Activation, renewal, payment-failure, cancellation and expiry handling.
- Account and Subscription UI.
- Phase 9 release-gate documentation and GitHub Issue #99.

### Changed

- Completed V1 privilege enforcement across protected application domains.
- Strengthened tenant isolation and cross-tenant negative-test coverage.
- Kept tenant ownership checks as defense in depth.
- Separated control-plane authentication/subscription data from tenant recruitment data.
- Kept the recruitment workflow deliberately simple.
- Simplified Offer to made + accepted.
- Simplified Joining to joining date + joined.

### Security

- Server-side RBAC and tenant identity enforcement.
- Tenant database routing from authenticated tenant context.
- Cross-tenant access rejection.
- Protected admin user management.
- Subscription state enforcement.
- Authenticated/idempotent provider webhook handling.

### Release verification

- Implementation complete on main.
- Final CI, security, UAT and production-smoke verification remains required.

### Scope discipline

- No HRMS functionality.
- No microservice split.
- No accounting or financial ledger.
- No configurable permission-builder framework.
- No unnecessary Docker expansion.

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