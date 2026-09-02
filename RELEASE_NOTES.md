# SkillSifter v0.5.3

## CP11 + CP12 — Production Quality and Go-Live Readiness

SkillSifter v0.5.3 closes the CP11 Production Quality Gate and CP12 Recruitment Workflow UAT / Go-Live Readiness checkpoints together. No v0.5.2 release is created.

This is a lean stabilization release. It does not add new platform infrastructure or speculative engineering. The focus is verification of the existing recruitment workflow, CI quality gate, tenant isolation, authorization, and regression safety.

## Highlights

**Production quality gate.**

- Backend CI validates formatting, repository structure, migration structure, schema definitions, build, vet, tests, and coverage reporting.
- Frontend CI validates dependency installation, lint, tests, and production build.
- Backend and frontend workflows remain separate and intentionally simple.
- Continuous deployment remains outside this release.

**Recruitment workflow readiness.**

The target workflow is:

```text
Candidate → Candidate expertise → Requirement → Assignment
→ Screening → Submission → Interview → Decision → Commercial / Joining
```

Existing automated regression coverage protects the most important backend workflow boundaries, including assignment state transitions, snapshots, audit integrity, tenant isolation, and authorization.

**Focused UAT.**

- Added a concise go-live readiness matrix for the recruiter workflow.
- Product-owner/recruiter UAT remains the final application-level validation before production cutover.
- UAT failures are treated as real product bugs and fixed narrowly, with regression coverage added when appropriate.

## Security

- Tenant isolation remains a release gate.
- Assignment actors must belong to the assignment tenant.
- Assignment audit actors remain tenant-scoped.
- Role authorization remains part of the recruitment workflow validation.

## Testing & CI

The release preserves the existing CI quality gate rather than introducing a heavier pipeline.

### Backend CI

- Go formatting
- repository/migration structure validation
- schema definition validation
- `go build ./...`
- `go vet ./...`
- `go test ./...` with coverage
- coverage summary publication

### Frontend CI

- `npm ci`
- lint
- tests
- production build

## Release Scope

- CP11 and CP12 are delivered as one release.
- No v0.5.2 intermediate release.
- No new observability platform, distributed tracing, CD system, or speculative performance framework.
- Production performance and reliability work will be driven by real usage evidence.

## Go-Live Gate

Before production cutover:

1. CI must be green.
2. Database schema validation must be green.
3. No known blocking regression may remain in the recruitment workflow.
4. Manual recruiter UAT must be completed against the running application.
5. Tenant isolation and role authorization must be explicitly exercised.

## Operating Principle

```text
LEAN → STABLE → LIVE → OBSERVE → FIX → OPTIMIZE → EXPAND
```

Only production evidence should justify additional performance, observability, infrastructure, or architectural work.

See [`docs/release/v0.5.3-go-live-readiness.md`](docs/release/v0.5.3-go-live-readiness.md) for the detailed checkpoint matrix.

---

# SkillSifter v0.4.0

## Recruitment Assignment State Machine

SkillSifter v0.4.0 is a significant backend maturity release focused on recruitment assignment workflow integrity, tenant isolation, candidate expertise, audit enforcement, and automated CI.

The release establishes recruitment assignments as a controlled domain workflow rather than a loosely coordinated set of handler operations. State transitions, authorization boundaries, snapshots, and audit behavior are now explicitly protected and covered by regression tests.

## Highlights

**Recruitment assignment state machine.**

- Formalized the recruitment assignment lifecycle and valid state transitions.
- Enforced state-transition rules through the assignment domain.
- Prevented invalid or unsupported assignment transitions.
- Strengthened assignment state and snapshot handling.
- Added regression coverage for transition behavior and state integrity.

**Tenant-aware assignment authorization.**

- Assignment actors are explicitly validated against the assignment tenant.
- Cross-tenant assignment actors are rejected.
- Tenant isolation is enforced at the assignment domain boundary.
- Prevents an actor belonging to another tenant from being used in assignment operations.

**Audit integrity.**

- Assignment audit records enforce tenant-aware actor relationships.
- Audit operations validate that the actor belongs to the same tenant as the assignment.
- Added regression coverage for cross-tenant audit actors.
- Strengthened the assignment history so audit events cannot be associated with an unauthorized tenant actor.

**Candidate expertise model.**

- Added generic candidate technical expertise support.
- Added candidate language expertise support.
- Added proficiency frameworks and proficiency levels.
- Added tenant-aware persistence for candidate expertise.
- Added uniqueness constraints and supporting indexes.
- Removed obsolete candidate expertise columns in favor of dedicated expertise tables.

**Candidate status.**

Candidate status is now explicitly constrained to:

- `active`
- `inactive`
- `blacklisted`
- `archived`

Invalid candidate status values are rejected at the database level.

## Database

Migration `009_candidate_expertise.sql` introduces the new candidate expertise model and removes the obsolete `jlptlanguage` and `skills` candidate columns.

New tables:

- `candidate_language_expertise`
- `candidate_expertise`

The new expertise tables are tenant-aware, reference the candidate and company, enforce uniqueness, and include indexes for tenant/candidate access.

## Security

- Assignment actors are tenant-scoped.
- Cross-tenant assignment actors are rejected.
- Assignment audit actors are tenant-scoped.
- Cross-tenant audit relationships are rejected.
- Candidate expertise persistence is tenant-scoped.

## Testing & CI

The release includes expanded automated regression coverage across:

- Assignment domain services
- Assignment state transitions
- Assignment snapshots
- Assignment handlers
- Candidate handlers
- Tenant isolation
- Assignment actor authorization
- Assignment audit enforcement

### Backend CI

GitHub Actions validates:

- Go formatting
- `go build ./...`
- `go vet ./...`
- `go test ./...`

### Frontend CI

GitHub Actions validates:

- dependency installation with `npm ci`
- ESLint
- Vitest
- production build

Backend and frontend CI are maintained as separate workflows so each side of the application has an independent, maintainable verification pipeline.

## Verification

The v0.4.0 implementation passed the local verification gate:

- Backend build: PASS
- Backend vet: PASS
- Backend tests: PASS
- Frontend lint: PASS
- Frontend tests: PASS
- Frontend production build: PASS
- GitHub Actions Backend CI: PASS
- GitHub Actions Frontend CI: PASS

## Release Scope

This release focuses on:

- recruitment assignment domain hardening
- controlled assignment state transitions
- tenant isolation
- assignment actor authorization
- audit integrity
- candidate expertise
- candidate status constraints
- backend and frontend CI foundations

Continuous deployment is intentionally **not** part of v0.4.0. CD will be introduced separately when the deployment workflow is ready.

## Upgrade Notes

Apply database migrations in the normal migration sequence before using the new candidate expertise functionality.

No special manual application-level migration procedure is required beyond the project's existing migration process.

There is no legacy compatibility or data migration for the obsolete candidate `jlptlanguage` and `skills` columns; the new expertise model is the authoritative representation.

## Release Reference

- Version: `v0.4.0`
- Release date: `2026-08-31`
- Release target: `main`
- Feature milestone: Recruitment Assignment State Machine

See [CHANGELOG.md](CHANGELOG.md) for the itemized change history.
