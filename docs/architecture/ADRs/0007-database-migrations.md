# ADR 0007: Database Migration Strategy

**Status:** Accepted  
**Decider:** Product owner  
**Related issue:** #23

## Context

The current repository contains a migration directory and a Go `ApplyMigrations` function, but the implementation currently applies a single reporting/resume migration rather than deterministically executing the complete migration set. The repository also contains duplicate numeric migration prefixes (`002_*`), creating an ordering ambiguity. V1 requires deterministic fresh-install and upgrade behaviour across local, test, container, and production environments.

## Decision

SkillSifter V1 will use a **single application-owned Go migration runner** as the authoritative migration execution path. The runner will execute ordered SQL migration files from `backend/database/migrations/` and maintain durable migration tracking in PostgreSQL.

No new external migration framework is required by this architecture decision. The existing Go migration mechanism is the starting point, but it must be evolved from the current one-off migration execution into the deterministic runner defined below.

## Migration Source of Truth

- `backend/database/migrations/` is the authoritative ordered migration set.
- Migration filenames use a unique numeric sequence followed by a descriptive name, for example `001_baseline.sql`, `002_ai_reporting.sql`.
- Sequence numbers must be unique. Two migration files with the same sequence are an error and must stop execution rather than being selected arbitrarily.
- Existing duplicate `002_*` files must be reconciled during implementation; the runner must not silently choose between them.
- Migration files are immutable after they have been applied to a deployed database. A correction requires a new migration.

## Migration Tracking

A dedicated migration-history table records at minimum:

- migration version/sequence
- migration name
- applied timestamp
- checksum/content identity
- execution status where required for diagnostics

The runner compares the repository migration set with recorded history and refuses unsafe divergence, including a changed checksum for an already-applied migration.

## Fresh Installation

A fresh database is initialized by running the complete migration sequence in ascending order. `schema.sql` is **not** a second competing schema-authority path.

`schema.sql` remains a documented reference/bootstrap artifact during the transition and may be used for inspection or controlled bootstrap compatibility, but the migration sequence is the authoritative definition of database evolution. Future schema changes must be represented by migrations.

## Upgrade

An existing database runs only migrations that are not recorded as successfully applied, strictly in sequence order. Previously applied migrations are never re-executed merely because the application restarts.

Application startup must fail safely if a required migration cannot be applied. The service must not continue as though the database were current.

## Transaction and Failure Behaviour

Each migration should execute atomically where PostgreSQL semantics permit. A failed migration must not be recorded as successfully applied. The runner stops at the first failed migration and reports the exact migration/version requiring operator attention.

The runner must not silently skip a failed migration or continue with later migrations.

## Rollback and Recovery

V1 does not require automatic down-migrations. Recovery is operational and controlled:

1. stop application startup on migration failure;
2. preserve the failing migration and database diagnostics;
3. restore from a verified backup when destructive recovery is required, or apply a corrective forward migration when safe;
4. resume the ordered migration sequence after the database is restored to a valid state.

Production deployment procedures must take a database backup before migrations that may be destructive or materially alter data.

## Startup and Deployment Consistency

The same migration runner and migration set must be used for local startup, test environments, container deployment, and production deployment. There must not be separate migration logic for Docker, native execution, or ad-hoc deployment scripts.

Application startup may invoke the runner, but deployment tooling must not apply a different subset or ordering.

## Testing Requirements

The implementation must test at minimum:

- fresh database → all migrations applied in order;
- upgrade database → only pending migrations applied;
- restart after successful migration → no duplicate application;
- duplicate migration sequence → deterministic failure;
- changed checksum of an applied migration → deterministic failure;
- failed migration → no success record and later migrations do not execute;
- migration recovery from a restored/valid database;
- schema/migration parity for supported fresh-install scenarios;
- concurrent startup behaviour so multiple application instances cannot corrupt migration history or apply the same migration unsafely.

## Implementation Boundary

This ADR defines migration architecture only. It does not authorize redesigning business tables, tenant schema, recruitment assignment schema, audit schema, or other domain models. Those changes must arrive through separately scoped implementation issues and migrations.

## Consequences

### Positive

- One deterministic migration path across environments.
- Fresh installs and upgrades use the same ordered migration set.
- Applied migrations are tracked and protected against silent modification.
- Migration failures stop startup instead of leaving the application on an unknown schema version.
- The architecture does not require introducing another major technology dependency.

### Trade-offs

- The existing migration implementation must be strengthened substantially.
- Duplicate migration numbering must be reconciled before the runner can be authoritative.
- Production deployments need reliable database backup/recovery procedures.
- Automatic down-migrations are intentionally excluded from V1.
