-- 017_resume_ai_tenant_isolation.sql
-- Resume records are tenant-owned. company_name remains display/compatibility
-- data, but must not be the isolation or duplicate boundary.

INSERT INTO companies (id, name, created_at)
SELECT DISTINCT
    'comp_' || regexp_replace(lower(r.company_name), '\\s+', '_', 'g'),
    r.company_name,
    NOW()
FROM resumes r
WHERE r.company_name IS NOT NULL
  AND r.company_name <> ''
  AND NOT EXISTS (SELECT 1 FROM companies c WHERE c.name = r.company_name)
ON CONFLICT (id) DO NOTHING;

ALTER TABLE resumes
    ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id);

UPDATE resumes r
SET tenant_id = c.id
FROM companies c
WHERE r.tenant_id IS NULL
  AND r.company_name = c.name;

DO $$
DECLARE unresolved INTEGER;
BEGIN
    SELECT COUNT(*) INTO unresolved
    FROM resumes
    WHERE tenant_id IS NULL;

    IF unresolved > 0 THEN
        RAISE EXCEPTION
            'resume tenant_id backfill incomplete: % resume row(s) have no matching companies row for company_name',
            unresolved;
    END IF;
END $$;

ALTER TABLE resumes
    ALTER COLUMN tenant_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_resumes_tenant
    ON resumes(tenant_id);

DROP INDEX IF EXISTS uq_resumes_company_hash;

CREATE UNIQUE INDEX IF NOT EXISTS uq_resumes_tenant_hash
    ON resumes(tenant_id, file_hash);

-- Jobs were retired in favor of Requirements. The historical 002 migration
-- created this trigger before the Jobs domain was removed. Remove only the
-- obsolete trigger through a forward migration; historical migrations stay
-- immutable.
DROP TRIGGER IF EXISTS trg_jobs_activity ON jobs;
