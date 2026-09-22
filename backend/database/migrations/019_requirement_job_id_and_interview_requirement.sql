-- 019_requirement_job_id_and_interview_requirement.sql
-- Job ID is the stable business/master index key for the recruitment demand
-- lifecycle. It is entered on Requirement and selected when an Interview is
-- scheduled. Candidate submission/assignment and Client do not use Job ID.
--
-- Existing Requirements and Interviews are preserved. The new relationships
-- remain nullable at the database level so historical rows that predate Job ID
-- can continue to exist. Application workflows require Job ID for new
-- Requirements and requirement_id for new/updated Interviews.

ALTER TABLE requirements
    ADD COLUMN IF NOT EXISTS job_id VARCHAR(100);

CREATE UNIQUE INDEX IF NOT EXISTS idx_requirements_tenant_job_id
    ON requirements(tenant_id, job_id)
    WHERE job_id IS NOT NULL;

ALTER TABLE interviews
    ADD COLUMN IF NOT EXISTS requirement_id INTEGER REFERENCES requirements(id);

CREATE INDEX IF NOT EXISTS idx_interviews_requirement
    ON interviews(requirement_id);
