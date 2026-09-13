# Changelog

All notable changes to SkillSifter are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project uses [Semantic Versioning](https://semver.org/).

## [0.5.6] - 2026-09-13

### Removed
- Assignments module (`Assignments.tsx`, `AddAssignment.tsx`) — redundant with Daily Tasks, which already covers candidate-requirement matching for internal recruiter workflows.
- Business Dev frontend module (`BusinessDev.tsx`, `AddBusinessDev.tsx`) — its two unique fields (`partnerName`, `contactPerson`) are now part of the Client model instead.

### Changed
- `Client` model and `client_handlers.go`: added optional `partner_name` and `contact_person` fields, absorbing Business Dev's distinguishing data onto the Client record.
- `Clients.tsx` / `AddClient.tsx`: updated to display and capture the new partner/contact fields.
- Navbar: nav-links section is now independently scrollable, keeping the logo and right-side actions (Add Candidate, Logout) pinned and visible as the number of tabs grows.

### Migration
- `013_client_partner_and_contact_person.sql` — adds `partner_name` and `contact_person` columns to `clients`. No data migration required (pre-release, no production client data to carry over).

### Known limitations
- The backend `BusinessDev` struct and its underlying table remain in place, now unused by the frontend — candidate for cleanup in a future release, similar to the `skills`/`candidate_skills` table removal in v0.5.5.



## [0.5.5] - 2026-09-08

### Added
- ADR 0009: Docker backend architecture and dev-branch deployment strategy — establishes PostgreSQL as permanently independent of any container, backend as the sole containerized component, and git-driven CD discipline for the `dev` line.
- Audit event read access (`AuditEventRepository.GetByEntity`) — the audit-events table previously supported writes only; this adds the ability to fetch an entity's audit history back out, e.g. for a future activity/audit-trail view.

### Fixed
- Resume AI: candidate resume list, search, and upload results were not rendering — the frontend expected a bare array from `resume-ai` endpoints, but the backend returns the standard `{success, message, data}` envelope. Frontend now unwraps `response.data.data` consistently across all three call sites.

### Changed
- `docs/architecture.md`: corrected a stale reference to `db.ApplyMigrations()` (renamed to `InitializeSchema()` after v0.4.0); rewrote the Deployment Architecture section, which still described the three-container `docker-compose` model with PostgreSQL running inside Docker — the exact pattern ADR 0009 supersedes.
- Backported explanatory documentation (tenant-scoping rationale, error-mapping decisions, route organization) into `assignment_handlers.go`, `client_handlers.go`, and `main.go` from earlier development-branch work that had richer comments than what shipped.

### Removed
- Dropped the unused `skills` and `candidate_skills` tables. Both were provisioned by an earlier design for normalized, confidence-scored skill tagging, but no code path ever read or wrote either table — Resume AI's actual extraction path writes into `candidate_expertise` instead. Confirmed empty in production before removal. `candidate_expertise` remains the single authoritative skills-storage mechanism.

### Process
- Repository branch and tag hygiene: established a consistent `v<version>-dev` naming convention, removed several stale and inconsistently-named development branches and two malformed release tags that had accumulated outside that convention.


## [0.5.3] - 2026-09-02

### Added
- Combined CP11 Production Quality Gate and CP12 Recruitment Workflow UAT / Go-Live Readiness checkpoint.
- Added a concise go-live readiness matrix covering the recruiter workflow from candidate through commercial/joining.

### Changed
- Confirmed the existing backend and frontend GitHub Actions workflows are sufficient for the release quality gate.
- Backend CI validates formatting, repository/migration structure, schema definitions, build, vet, tests, and coverage reporting.
- Frontend CI validates dependency installation, lint, tests, and production build.
- CP12 uses existing automated regression coverage plus focused manual recruiter UAT rather than introducing heavyweight browser-test infrastructure.
- Continuous deployment remains separate from this release.

### Security
- Release readiness explicitly requires tenant-isolation and role-authorization validation across the recruitment workflow.
- Assignment actor authorization and audit tenant enforcement remain covered by regression tests.

### Testing
- Existing backend regression coverage includes candidate, assignment, state-transition, snapshot, audit, authorization, and tenant-isolation paths.
- Existing frontend CI validates lint, tests, and production build.
- Manual recruiter UAT remains the final product-level gate before production cutover.

### Release Scope
- CP11 and CP12 are completed as one release checkpoint.
- No v0.5.2 release is created.
- No new observability, distributed infrastructure, CD system, or speculative performance framework is introduced.
- Production performance and reliability work will be driven by real usage evidence.

## [0.4.0] - 2026-08-31

### Added
- Recruitment assignment state-machine implementation with controlled lifecycle transitions.
- Tenant-aware assignment actor authorization.
- Assignment audit actor tenant enforcement.
- Generic candidate technical expertise model.
- Candidate language expertise model with proficiency framework and proficiency levels.
- GitHub Actions CI for backend and frontend.

### Changed
- Recruitment assignments now enforce valid state transitions through the assignment domain.
- Assignment state and snapshot handling has been strengthened around the recruitment workflow.
- Candidate status is now explicitly constrained to `active`, `inactive`, `blacklisted`, or `archived`.
- Candidate expertise has moved from obsolete candidate columns to dedicated expertise tables.
- Backend and frontend verification now runs automatically through GitHub Actions.

### Security
- Assignment actors are tenant-scoped.
- Cross-tenant assignment actors are rejected.
- Assignment audit actors are tenant-scoped.
- Cross-tenant audit relationships are rejected.

### Testing
- Expanded assignment state-machine regression coverage.
- Added tenant-isolation regression coverage.
- Added assignment actor authorization regression coverage.
- Added assignment audit enforcement regression coverage.
- Backend CI validates formatting, build, vet, and tests.
- Frontend CI validates lint, tests, and production build.

### Release Scope
- This release focuses on recruitment assignment domain hardening, tenant isolation, audit integrity, candidate expertise, and CI foundations.
- Continuous deployment is intentionally not part of this release.

## [0.3.0] - 2026-08-23

### Added
- Local, recruiter-assisted resume ingestion and search foundation, including resume metadata, extracted skills, activity records, and configurable local Ollama connectivity.
- A V1 architecture and product-scope baseline, Codex engineering rules, ADR process, and a v0.4 architecture-foundation backlog.

### Changed
- The repository's current baseline is designated v0.3.0. v0.4.0 is reserved for the architecture-foundation milestone and is not a feature release.

### Known limitations
- The v0.4 backlog records unresolved tenant identity, migration, RBAC, and domain-model decisions. No later-milestone product feature is implied by this baseline.

## [0.2.0] - 2026-08-02

### Added
- Role-based access control (RBAC): delete/edit hierarchy (Admin undeletable; Manager deletable only by Admin; Recruiter/Team Leader deletable by either), job ownership (`created_by_user_id`), manager routes (previously registered but empty)
- Backend test infrastructure: unit tests for auth/JWT logic and RBAC middleware, integration tests for candidate CRUD against a real database
- Frontend test infrastructure: Vitest setup, behavioral tests for the API service layer
- `infra/scripts/test.sh`: single command running the full backend (fmt/vet/build/test) and frontend (lint/test/build) gate
- `deploy.sh` now runs the backend test gate before touching the live service, aborting cleanly on any failure
- Real Recent Activity feed on the Dashboard (`GET /api/reports/activity`), replacing hardcoded mock data, built from real timestamps across candidates, jobs, business_dev, daily_jobs, and interviews
- `GET /api/company-users` endpoint (was missing entirely — blocked the Daily Tasks assignee dropdown)
- Skill alias reference table and job ownership migrations
- CORS wildcard support for Cloudflare Pages preview URLs

### Fixed
- **Critical**: Candidates CRUD (`GetCandidates`, `GetCandidateByID`, `AddCandidate`, `UpdateCandidate`) referenced database columns that did not exist in the actual schema — every call failed. Likely broken since an earlier schema redesign.
- **Critical**: JWT secret was still hardcoded in source despite an earlier fix attempt — the actual code change had never been applied. Rotated the secret after deploying the real fix.
- **Critical**: `GetJobs`/`GetJobByID` column-count mismatch after the RBAC migration added `created_by_user_id` — every job listing request failed silently.
- Daily Tasks assignment form couldn't be submitted (assignee dropdown was always empty due to the missing `/company-users` route)
- Daily Tasks list view crashed to a blank page on row click (route was never registered)
- `UpdateUser` and `DeleteUser` were non-functional stubs; both now perform real, hierarchy-enforcing operations
- 38 frontend lint errors (unsafe `any` typing throughout, empty interface declarations, a legacy `require()` import) and all associated warnings
- Repo-wide CRLF line endings in backend Go files, which silently broke `gofmt` enforcement
- Stray duplicate `.gitgnore` file and an accidentally committed `.env` containing a leaked secret value

### Changed
- Repository restructured: frontend moved into `frontend/`, symmetric with `backend/` (previously sat loose at repo root)
- `Candidate` model and handlers rewritten to match the actual deployed schema (location, experience, CTC, notice period, skills, etc.) instead of stale fields from an earlier design (status, source, resume URL, cover letter)

### Known limitations (not yet fixed, tracked for a future release)
- `GET /api/reports/sources` returns `500` — queries a `candidates.source` column that does not exist
- Interviews: display and scheduling not verified working
- Dashboard's Hiring Trend, Candidate Sources charts, and Recruitment Pipeline widget remain hardcoded/unwired
- CORS origins remain hardcoded in source rather than environment-configurable (the preview-URL wildcard added this release is a partial mitigation, not the full fix)
- No automated CI (GitHub Actions) — the test gate (`test.sh`) is real but run manually, not on push

## [0.1.0] - 2026-07-30 (approximate)

Initial documented baseline. README and architecture documentation added; JWT secret and compiled-binary tracking issues identified (fix landed in 0.2.0, see above).
