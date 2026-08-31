# Changelog

All notable changes to SkillSifter are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project uses [Semantic Versioning](https://semver.org/).

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
