# SkillSifter — Greenfield Database/Application Audit

## Audit status

Initial audit performed against the current `main` branch after establishing the Application/Database Architectural Invariants.

Scope: migration runner, schema definitions, control-plane registration/subscription lifecycle, tenant lifecycle, recruitment-domain relationships, PostgreSQL triggers/functions, and application concurrency handling.

## Executive finding

The current implementation is functionally close to the intended product, but the persistence layer still contains historical architecture embedded in both migrations and runtime schema initialization.

The correct action is a greenfield schema reset rather than another compatibility migration.

## RED — must be removed or redesigned

### 1. Historical migration chain

`backend/database/migrations/` currently contains the accumulated 001–045 evolution.

Problem:
- legacy Jobs schema
- companies tenancy
- compatibility bridges
- backfills
- cleanup migrations
- temporary smoke-account data
- historical subscription evolution
- old AI/activity schema
- migration ordering assumptions

Decision: replace the historical chain with clean control-plane and tenant-plane baselines.

### 2. Migration runner contains historical architecture

`backend/db/migrations.go` currently makes tenant schema selection depend on `version <= 34` and then performs runtime cleanup of `companies`.

This is unacceptable in the final design because schema ownership is encoded in historical migration numbers rather than explicit architecture.

Decision: redesign the runner around explicit control-plane and tenant-plane schema sets.

### 3. PostgreSQL business triggers/functions

The old migrations contain PostgreSQL trigger/function implementations.

Examples:
- `002_ai_reporting.sql` creates `skillsifter_activity_trigger()` and row-level activity triggers.
- `010_candidate_recruitment_engagement.sql` creates candidate engagement claim/release triggers.
- `011_cp10_service_owned_engagement.sql` subsequently removes the engagement triggers/functions because the application service owns that logic.

Decision: no business workflow triggers/functions in the final baseline. Application services own these transitions. PostgreSQL may still provide atomic primitives whose result is explicitly returned to Go.

### 4. Companies entity

Current code still contains migration/runtime references to `companies`, including historical foreign keys and compatibility cleanup.

Decision: `companies` does not exist in the final schema.

### 5. Cross-boundary relational FKs

Current control-plane and historical tenant schemas contain relationships such as:
- tenant -> platform_tenants
- platform user -> users
- subscription -> tenant
- domain records -> candidates/requirements/users/clients
- multiple historical tenant_id -> companies relationships

The user-design principle is that these relationships must not become hidden domain decision engines.

Decision: each relationship must be classified during the next audit as either:
1. structural persistence integrity that is intentionally retained, or
2. application-owned relationship that is removed from PostgreSQL and validated explicitly by Go.

No relationship will survive merely because it existed in a historical migration.

### 6. Database cascade as lifecycle authority

`CleanupExpiredTrials()` currently deletes `platform_tenants` and relies on cascade behavior to remove dependent control-plane records.

That makes PostgreSQL silently determine part of the application's deletion semantics.

Decision: application cleanup must explicitly perform the control-plane lifecycle operations. The final schema must not require cascades to express business deletion policy.

### 7. Schema seed/test data mixed with schema

`038_e2e_smoke_account.sql` inserts a smoke account into the schema installation path.

Decision: remove all test/customer/manual data from the production schema baseline. Test fixtures/bootstrap must own test data.

### 8. Historical commercial-plan migrations

`039_tbd_subscription_plan.sql` and `040_commercial_plans.sql` are operational seed/configuration evolution, not schema structure.

Decision: final schema and controlled plan/bootstrap data are separate concerns.

## AMBER — needs explicit redesign review

### 9. CHECK constraints

Current schema uses PostgreSQL CHECK constraints for state values and some field relationships.

These must be reviewed individually.

Rule: a CHECK that protects a simple physical invariant may be retained; a CHECK that represents a domain workflow decision should move to Go, with PostgreSQL used only for atomic persistence where needed.

### 10. UNIQUE constraints/indexes

Uniqueness is frequently useful as a concurrency arbiter, but the application must interpret the result.

Example:
- PostgreSQL prevents two simultaneous registrations from claiming the same permanent email.
- Go receives the conflict/result and returns the domain-level duplicate-registration response.

This is compatible with the architecture because PostgreSQL is arbitrating concurrency, not defining the domain meaning.

The permanent registration registry therefore needs a deliberate final decision on which physical uniqueness primitive is retained for concurrency safety.

### 11. Application pre-check followed by write

Several services first query for existence and then perform an INSERT/UPDATE.

This is safe only when the final write is itself concurrency-safe.

Audit target: replace race-prone check-then-write patterns with atomic operations where the operation can be contested concurrently, and explicitly translate the database result into domain errors.

## GREEN — architecture already aligned

### 12. Application-owned workflow services

Offer, Joining, and Billing services already perform meaningful workflow checks in Go before persistence.

Examples observed:
- Offer requires a selected recruitment decision.
- Joining requires an accepted offer.
- Joining requires a joining date when `joined=true`.
- Billing requires a joined recruitment record.

These are good examples of application-owned domain meaning.

### 13. Tenant routing

The current runtime already distinguishes control-plane database access from tenant database access through the database routing layer.

This should be preserved and simplified rather than replaced.

### 14. Permanent registration identity

The permanent registration registry is conceptually correct:
- email is the permanent identity
- it survives tenant deletion
- it stores minimal identity information
- it is written after successful email verification

The implementation should be retained in the clean baseline, with its concurrency semantics reviewed.

## Immediate implementation sequence

1. Freeze the invariants document.
2. Complete the schema object inventory.
3. Complete the application query/write inventory.
4. Classify every FK, CHECK, UNIQUE index, trigger/function, cascade, and DB-side mutation.
5. Design the final control-plane baseline.
6. Design the final tenant-plane baseline.
7. Replace the migration runner with explicit control/tenant schema initialization.
8. Move any remaining domain decisions from PostgreSQL into Go.
9. Introduce atomic DB operations where concurrency requires them and return explicit results to Go.
10. Rewrite tests against clean empty databases.
11. Wipe non-production development data and rebuild from the new baselines.
12. Run full CI and only then resume production deployment testing.

## Audit principle

**Do not patch the historical architecture. Extract the intended final system and rebuild it cleanly.**

## Implementation progress — final boundary pass

The greenfield reconstruction has now started.

### Completed
- Final control-plane architecture frozen in `docs/architecture/final-control-plane.md`.
- Final tenant-plane architecture frozen in `docs/architecture/final-tenant-plane.md`.
- Added `backend/database/control-plane/001_baseline.sql`.
- Added `backend/database/tenant-plane/001_baseline.sql`.
- Migration runner no longer uses historical numeric cutoffs. Control and tenant schemas are initialized from explicit schema roots with independent `schema_versions`.
- Tenant provisioning no longer mirrors users from the control database.
- Tenant-local login now resolves platform routing/subscription first, then authenticates against the tenant database.
- Admin user management now creates/updates/deletes tenant-local users and maintains only the control-plane routing record.
- Phone verification now belongs to the control-plane platform account, not tenant `users`.
- Razorpay subscription webhook idempotency now uses PostgreSQL `ON CONFLICT DO NOTHING` as the concurrency arbiter and records the event in the same transaction as subscription state.
- Subscription webhook writes the required subscription user limit.
- Interview persistence no longer requires legacy `job_id` or `company_name`.

### Remaining RED work
1. Remove remaining runtime tenant queries that select/insert `company_name`.
2. Remove remaining Resume AI compatibility columns/queries that depend on `company_name`.
3. Replace legacy activity/reporting reads with tenant `audit_events` and explicit application-generated events.
4. Complete atomic user-limit enforcement under concurrent user creation.
5. Complete atomic registration-email claim and verification replay handling.
6. Complete tenant provisioning retry/recovery without requiring a tenant user to exist before the tenant DB is ready.
7. Complete explicit tenant deletion sequence; no lifecycle meaning should depend on control-plane cascades.
8. Move remaining tests from the historical shared schema to clean control/tenant databases.
9. Remove legacy schema bootstrap assumptions and retire the historical migration chain from runtime execution.
10. Run the full CI/UAT gate only after the clean baseline and handler/query audit are complete.

### Current rule

Do not add compatibility migrations to make the historical shared schema look like the new architecture. The final baselines are the target; Go code and tests are being brought to those contracts.
