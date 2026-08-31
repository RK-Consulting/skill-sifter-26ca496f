# SkillSifter v0.4 Quality Gates

## Purpose

This document defines the minimum engineering checks required before feature work is accepted into the repository. These are acceptance criteria, not a claim that every check is currently automated in GitHub Actions.

## Gate status

| Gate | Required | Current status |
|---|---|---|
| Backend formatting | Yes | Local/CI check required |
| Backend unit tests | Yes | Local/CI check required |
| Backend static analysis | Yes | Local/CI check required |
| Frontend lint | Yes | Local/CI check required |
| Frontend build | Yes | Local/CI check required |
| Database migrations | Yes when schema changes | Migration validation required |
| Integration/API tests | Yes for affected backend behavior | Required for affected changes |
| Docker build | Yes for infrastructure/container changes | Required when affected |
| GitHub Actions | Yes when workflow exists | Must pass before merge |
| Deployment | Not a v0.4 acceptance gate | Deployment verification is separate |

## Backend gate

For backend changes:

1. Run `gofmt` on changed Go files.
2. Run `go test ./...` from `backend/`.
3. Run `go vet ./...` from `backend/`.
4. Changes that affect API behavior must include or update relevant handler/service tests.
5. Security-sensitive changes must include negative authorization or tenant-isolation coverage where applicable.

## Frontend gate

For frontend changes:

1. Run the repository's configured lint command.
2. Run the production build.
3. Changes to API contracts must be checked against the API/domain conventions in ADR 0008.
4. UI changes affecting existing workflows must preserve the current supported behavior unless the issue explicitly changes it.

## Database and migration gate

For schema or migration changes:

- A migration must be deterministic and reviewable.
- Existing data must remain compatible unless the issue explicitly authorizes a breaking migration.
- Migration order and rollback/compatibility implications must be documented where applicable.
- Application startup must not silently discard existing data.
- Tenant-scoped changes must preserve the isolation guarantees established by ADR 0001.

## Build and integration gate

A change is not considered complete if it only compiles in isolation but breaks the application composition.

When affected, validate:

- backend build;
- frontend production build;
- Docker image/build configuration;
- API integration behavior;
- database initialization/migration behavior.

## GitHub Actions gate

GitHub Actions is an acceptance gate only when the relevant workflow is actually implemented and running. Documentation must not describe CI as active merely because a workflow is planned.

When CI is implemented, required checks should cover at minimum:

- backend test and static analysis;
- frontend lint and build;
- affected integration checks;
- repository/build integrity checks.

Deployment workflows are separate from pull-request quality gates and must not be treated as proof that application tests passed unless the workflow explicitly performs those checks.

## Pull request acceptance criteria

A future PR should be accepted only when:

- the linked issue's acceptance criteria are satisfied;
- required local checks pass;
- applicable automated checks pass;
- no unrelated behavior or files are changed without justification;
- API/database/security compatibility is preserved;
- documentation is updated when an architectural or operational convention changes;
- migrations are included when schema changes require them;
- tests cover new behavior and important regressions;
- the PR description identifies known limitations or intentionally deferred work.

## CI truthfulness rule

The repository must distinguish between **required checks** and **implemented checks**. A check may be required by policy before it is automated. Until a GitHub Actions workflow actually runs that check, documentation should describe it as a required local/manual gate rather than an active CI gate.

## Scope

These gates establish the v0.4 engineering baseline. They do not authorize unrelated refactoring, API breaks, schema redesign, or deployment changes.