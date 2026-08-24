-- 006_tenant_identity.sql
-- Issue #33 / ADR 0001, Stage A + Stage B.
--
-- Stage A: companies.id is evaluated and retained as the authoritative,
-- immutable tenant identifier. It already exists (VARCHAR(255) PRIMARY KEY,
-- assigned once at company-registration time and never reassigned), so no
-- schema change is required for the companies table itself.
--
-- Stage B: introduce tenant_id into every tenant-owned table (users,
-- candidates, jobs, daily_jobs, interviews, business_dev), backfilled from
-- the existing company_name values. company_name is intentionally NOT
-- removed or stopped being written by application code in this migration —
-- it remains business/display data per ADR 0001. Retiring it as an
-- isolation key is Stage E, out of scope here.
--
-- This migration does not change company_name values, does not delete any
-- row, and does not repoint any row at a different tenant than the one its
-- company_name already implied.

-- Self-heal step: a tenant-owned row's company_name might not have a
-- matching companies row (e.g. data entered before the companies table was
-- consistently populated). Rather than leave such rows unbackfillable,
-- create the missing companies row using the same slug convention already
-- used at registration time (see handlers/auth_handlers.go RegisterUser:
-- "comp_" + lowercased, space-to-underscore company name). This prevents
-- existing data from becoming inaccessible during migration (ADR 0001
-- compatibility requirement).
INSERT INTO companies (id, name, created_at)
SELECT DISTINCT
    'comp_' || regexp_replace(lower(t.company_name), '\s+', '_', 'g') AS id,
    t.company_name,
    NOW()
FROM (
    SELECT company_name FROM users
    UNION SELECT company_name FROM candidates
    UNION SELECT company_name FROM jobs
    UNION SELECT company_name FROM daily_jobs
    UNION SELECT company_name FROM interviews
    UNION SELECT company_name FROM business_dev
) t
WHERE t.company_name IS NOT NULL
  AND t.company_name <> ''
  AND NOT EXISTS (SELECT 1 FROM companies c WHERE c.name = t.company_name)
ON CONFLICT (id) DO NOTHING;

-- Add tenant_id as nullable first so existing rows are never rejected
-- mid-backfill.
ALTER TABLE users        ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id);
ALTER TABLE candidates   ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id);
ALTER TABLE jobs         ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id);
ALTER TABLE daily_jobs   ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id);
ALTER TABLE interviews   ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id);
ALTER TABLE business_dev ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id);

-- Backfill from company_name -> companies.id. Only touches rows that don't
-- already have a tenant_id, so this migration is safe to design for
-- (in practice it only ever runs once, per the migration runner).
UPDATE users u        SET tenant_id = c.id FROM companies c WHERE u.company_name = c.name AND u.tenant_id IS NULL;
UPDATE candidates t    SET tenant_id = c.id FROM companies c WHERE t.company_name = c.name AND t.tenant_id IS NULL;
UPDATE jobs t          SET tenant_id = c.id FROM companies c WHERE t.company_name = c.name AND t.tenant_id IS NULL;
UPDATE daily_jobs t    SET tenant_id = c.id FROM companies c WHERE t.company_name = c.name AND t.tenant_id IS NULL;
UPDATE interviews t    SET tenant_id = c.id FROM companies c WHERE t.company_name = c.name AND t.tenant_id IS NULL;
UPDATE business_dev t  SET tenant_id = c.id FROM companies c WHERE t.company_name = c.name AND t.tenant_id IS NULL;

-- Safety gate: if any row could not be backfilled (which should only be
-- possible if company_name is NULL/empty on a row, already disallowed by
-- the existing NOT NULL constraints, but checked defensively rather than
-- assumed), stop this migration rather than silently enforcing NOT NULL
-- and letting Postgres produce a less diagnosable constraint-violation
-- error. Per ADR 0007, a failed migration is not recorded as applied and
-- production data is left untouched for the operator to resolve.
DO $$
DECLARE unresolved INTEGER;
BEGIN
    SELECT
        (SELECT COUNT(*) FROM users        WHERE tenant_id IS NULL) +
        (SELECT COUNT(*) FROM candidates   WHERE tenant_id IS NULL) +
        (SELECT COUNT(*) FROM jobs         WHERE tenant_id IS NULL) +
        (SELECT COUNT(*) FROM daily_jobs   WHERE tenant_id IS NULL) +
        (SELECT COUNT(*) FROM interviews   WHERE tenant_id IS NULL) +
        (SELECT COUNT(*) FROM business_dev WHERE tenant_id IS NULL)
    INTO unresolved;

    IF unresolved > 0 THEN
        RAISE EXCEPTION 'tenant_id backfill incomplete: % row(s) across tenant-owned tables have no matching companies row for their company_name. Migration 006_tenant_identity stopped; resolve the underlying data before re-running.', unresolved;
    END IF;
END $$;

ALTER TABLE users        ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE candidates   ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE jobs         ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE daily_jobs   ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE interviews   ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE business_dev ALTER COLUMN tenant_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_tenant        ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_candidates_tenant    ON candidates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_jobs_tenant          ON jobs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_daily_jobs_tenant    ON daily_jobs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_interviews_tenant    ON interviews(tenant_id);
CREATE INDEX IF NOT EXISTS idx_business_dev_tenant  ON business_dev(tenant_id);
