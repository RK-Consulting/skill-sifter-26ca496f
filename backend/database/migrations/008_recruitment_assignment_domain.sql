-- 008_recruitment_assignment_domain.sql
-- Issue #35 / ADR 0003, ADR 0004, ADR 0006.
--
-- Three additive, independent pieces:
--   1. candidates.status (ADR 0004) - required so assignment creation can
--      enforce candidate eligibility (ADR 0004 section 2).
--   2. recruitment_assignments (ADR 0003) - the candidate<->requirement
--      transaction, with embedded submission-snapshot columns (ADR 0003
--      section 6) rather than a separate snapshot table, since exactly one
--      snapshot is taken per assignment (at formal submission), not a
--      repeated history.
--   3. audit_events (ADR 0006) - a new, independent append-only event
--      model. This is NOT an extension of the legacy activity_logs
--      mechanism (ADR 0006 section 6 explicitly forbids that); it is a
--      separate table with its own tenant scoping and no full-row
--      serialization.
--
-- Nothing here modifies jobs, interviews, daily_jobs, clients, or the
-- legacy activity_logs table/triggers.

-- ============================================================
-- 1. Candidate status (ADR 0004)
-- ============================================================

ALTER TABLE candidates ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'active';

-- Backfill defensively for any row that predates the DEFAULT (Postgres
-- applies the DEFAULT to new rows automatically once the column exists,
-- and also backfills existing rows with the DEFAULT value as part of
-- adding a NOT NULL column with a DEFAULT in modern Postgres, but this
-- UPDATE is kept as an explicit, visible safety net rather than relying
-- solely on that behavior).
UPDATE candidates SET status = 'active' WHERE status IS NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'candidates' AND constraint_name = 'candidates_status_valid'
    ) THEN
        ALTER TABLE candidates ADD CONSTRAINT candidates_status_valid
            CHECK (status IN ('active', 'inactive', 'blacklisted', 'archived'));
    END IF;
END $$;

-- ============================================================
-- 2. Recruitment Assignment (ADR 0003)
-- ============================================================

CREATE TABLE IF NOT EXISTS recruitment_assignments (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_by_user_id INTEGER NOT NULL REFERENCES users(id),
    owner_user_id INTEGER NOT NULL REFERENCES users(id),

    -- Submission snapshots (ADR 0003 section 6). Populated once, at formal
    -- submission (draft/screening -> submitted). NULL before that point.
    -- Immutability is enforced at the application layer: no handler ever
    -- issues an UPDATE that touches these columns after they are first
    -- set.
    candidate_snapshot JSONB,
    requirement_snapshot JSONB,
    snapshot_created_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT recruitment_assignments_status_valid CHECK (
        status IN ('draft', 'screening', 'submitted', 'interviewing', 'offered', 'joined', 'rejected', 'withdrawn')
    ),

    -- ADR 0003 section 2: a candidate must not have multiple simultaneous
    -- assignment records for the same requirement. No re-attempt/versioning
    -- mechanism is implemented in this issue, so this is a hard, permanent
    -- uniqueness constraint for now, matching the ADR's explicit statement
    -- that reopening requires "an explicit product decision/versioning
    -- mechanism" not built here.
    CONSTRAINT recruitment_assignments_candidate_requirement_unique UNIQUE (candidate_id, requirement_id)
);

CREATE INDEX IF NOT EXISTS idx_recruitment_assignments_tenant ON recruitment_assignments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_assignments_candidate ON recruitment_assignments(candidate_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_assignments_requirement ON recruitment_assignments(requirement_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_assignments_owner ON recruitment_assignments(owner_user_id);

-- Defensive note (same pattern as 007's clients/requirements comment):
-- Postgres cannot express a cross-table CHECK that candidate.tenant_id,
-- requirement.tenant_id, owner_user.tenant_id, and created_by_user.tenant_id
-- all equal this row's tenant_id. That is enforced at the application layer
-- in the assignment handlers, which is where the authenticated tenant_id
-- is actually known and authoritative.

-- ============================================================
-- 3. Audit events (ADR 0006)
-- ============================================================

CREATE TABLE IF NOT EXISTS audit_events (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    actor_user_id INTEGER NOT NULL REFERENCES users(id),
    entity_type VARCHAR(100) NOT NULL,
    entity_id INTEGER NOT NULL,
    action VARCHAR(100) NOT NULL,
    occurred_at TIMESTAMP NOT NULL DEFAULT NOW(),
    correlation_id VARCHAR(100),
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB
);

CREATE INDEX IF NOT EXISTS idx_audit_events_tenant ON audit_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_entity ON audit_events(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_correlation ON audit_events(correlation_id);

-- Immutability (ADR 0006 section 5) is enforced at the application layer:
-- no handler issues UPDATE or DELETE against this table. This matches the
-- existing repository convention (no DB-level permission separation is
-- used anywhere else in this schema either).
