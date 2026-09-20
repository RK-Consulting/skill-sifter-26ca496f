-- 009_requirements_replace_jobs.sql
-- Requirements are now the single authoritative recruitment-demand model.
-- This is additive first: existing legacy jobs data is preserved until an
-- explicit data-retention decision is made. The application no longer
-- exposes or writes the jobs resource.

ALTER TABLE requirements
    ADD COLUMN IF NOT EXISTS job_type VARCHAR(30),
    ADD COLUMN IF NOT EXISTS budget VARCHAR(255),
    ADD COLUMN IF NOT EXISTS certifications_required TEXT,
    ADD COLUMN IF NOT EXISTS notice_period VARCHAR(100),
    ADD COLUMN IF NOT EXISTS mandatory_requirements TEXT;

-- Normalize the previous lifecycle into the new Requirement lifecycle.
UPDATE requirements SET status = 'open' WHERE status = 'draft';
UPDATE requirements SET status = 'closed' WHERE status = 'filled';

ALTER TABLE requirements DROP CONSTRAINT IF EXISTS requirements_status_valid;
ALTER TABLE requirements
    ADD CONSTRAINT requirements_status_valid
    CHECK (status IN ('open', 'closed', 'on_hold', 'cancelled'));

ALTER TABLE requirements DROP CONSTRAINT IF EXISTS requirements_job_type_valid;
ALTER TABLE requirements
    ADD CONSTRAINT requirements_job_type_valid
    CHECK (job_type IS NULL OR job_type IN ('fulltime', 'contract'));

ALTER TABLE requirements DROP CONSTRAINT IF EXISTS requirements_work_arrangement_valid;
ALTER TABLE requirements
    ADD CONSTRAINT requirements_work_arrangement_valid
    CHECK (work_arrangement IS NULL OR work_arrangement IN ('hybrid', 'remote', 'office'));

CREATE INDEX IF NOT EXISTS idx_requirements_status ON requirements(status);


-- Legacy jobs data is intentionally discarded. Requirements are now the
-- sole recruitment-demand model; no historical jobs data is required.
DROP TABLE IF EXISTS jobs CASCADE;
