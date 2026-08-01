# SkillSifter v0.2.0

This release hardens the v0.1.0 baseline: real role-based access control, a working test suite for the first time, and — critically — several core features (Candidates, Jobs, Daily Tasks) that were silently broken and are now fixed and verified.

## Highlights

**Candidates, Jobs, and Daily Tasks are now genuinely functional.** Each had a different root cause — a stale schema mismatch, a scan-count bug introduced by this release's own RBAC migration, and a missing API route respectively — but the practical effect was the same: core parts of the app appeared to work but silently failed. All three are now fixed and confirmed working end to end.

**Role-based access control is real, not just data.** Previously, the `roles` table defined permissions that were never actually checked in code. This release adds a real delete/edit hierarchy (an Admin can never be removed, by anyone; a Manager can only be removed by an Admin) and scopes job management to Manager/Admin roles, matching the permissions the system always claimed to enforce.

**The Dashboard's Recent Activity is real data.** It previously showed the same three fabricated entries to every user, regardless of what had actually happened in their account. It now shows real, timestamp-sorted events from across candidates, jobs, business development, daily tasks, and interviews.

**There is now a real, working test gate.** Zero automated tests existed before this release. `infra/scripts/test.sh` runs the full backend (format, vet, build, test) and frontend (lint, test, build) suite in one command, and `deploy.sh` now runs this gate automatically before ever touching the live service — a broken deploy is refused, not shipped.

## Security

- The JWT signing secret was still hardcoded in source, despite having been "fixed" in an earlier commit — the actual code change had never been applied. This is now genuinely fixed, and the exposed secret has been rotated.
- A `.env` file containing a leaked secret value had been accidentally committed to the repository; it has been removed and `.gitignore` corrected.

## Known limitations in this release

Being upfront about what's still broken or unfinished, rather than implying full coverage:

- The candidate-sources report (`/api/reports/sources`) currently errors — it references a database column that doesn't exist.
- Interview display and scheduling have not yet been verified working.
- The Dashboard's Hiring Trend chart, Candidate Sources chart, and Recruitment Pipeline widget are not yet wired to real data.
- CORS origins are still largely hardcoded in source, not environment-configurable.
- There is no automated CI (e.g. GitHub Actions) — the test gate is real, but currently run manually rather than triggered automatically on push.

See [CHANGELOG.md](CHANGELOG.md) for the complete, itemized list of changes.

## Upgrade notes

This release includes three new database migrations (skill aliases, job ownership, candidates timestamp), applied automatically by `deploy.sh` in order. No manual database steps are required.