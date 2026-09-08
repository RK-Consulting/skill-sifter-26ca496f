# SkillSifter v0.5.5

## Docker Architecture Decision, Resume AI Fix, and Repository Cleanup

SkillSifter v0.5.5 is a stabilization and cleanup release following a production incident on the DigitalOcean deployment. It establishes the target Docker/deployment architecture for future backend work (ADR 0009), fixes a broken Resume AI listing/search view, adds audit-event read access, corrects stale architecture documentation, and removes unused schema.

This release does not introduce new user-facing features. Backend containerization itself (per ADR 0009) is architecture and process, not yet implemented — that begins next on `v0.5.5-dev`.

## Highlights

**Deployment architecture decision (ADR 0009).**

- PostgreSQL is established as permanently independent of any container, on this host or any future one — the database's lifecycle must never be coupled to the application container's lifecycle.
- The backend is the only component intended to run in Docker going forward.
- `dev`-branch work must deploy via git-driven CD only; manual `docker run` / `docker compose up` on a server is explicitly disallowed, directly addressing the root cause of the incident this release follows.

**Resume AI fix.**

- The Resumes tab, search, and upload-results view were silently broken: the frontend expected list endpoints to return a bare array, but `resume-ai` endpoints return the standard `{success, message, data}` envelope used elsewhere in the API. Fixed by unwrapping `response.data.data` at all three call sites.

**Audit event read access.**

- Added `AuditEventRepository.GetByEntity`, so an assignment's audit history can be queried back out — previously the audit_events table supported writes only.

**Documentation corrections.**

- `docs/architecture.md` described a schema-initialization function (`ApplyMigrations`) that was renamed to `InitializeSchema` after v0.4.0, and a three-container `docker-compose` deployment model with PostgreSQL running inside Docker — both stale relative to the actual codebase and, in the deployment section's case, directly superseded by this release's own ADR 0009.

**Repository hygiene.**

- Standardized on a single `v<version>-dev` branch-naming convention.
- Removed multiple stale, differently-named, or fully-superseded development and feature branches, and two malformed release tags, that had accumulated outside that convention.

## Removed

- Dropped the unused `skills` and `candidate_skills` tables. Confirmed empty in production; no code path ever used them. `candidate_expertise` remains the single authoritative skills-storage mechanism for both curated and Resume-AI-extracted skill data.

## Testing & CI

No changes to the CI quality gate in this release. Existing backend and frontend regression coverage (formatting, build, vet, tests, lint) continues to gate every change, as in v0.5.3.

## Next

`v0.5.5-dev` begins the actual Docker/CI/CD implementation work described by ADR 0009: a backend `Dockerfile`-based build, CI image publishing, and a git-driven deployment mechanism for the `dev` line, isolated from production until proven stable.



See [CHANGELOG.md](CHANGELOG.md) for the itemized change history.
