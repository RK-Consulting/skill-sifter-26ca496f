-- 030_remove_assignment_runtime_paths.sql
-- Complete the recruitment workflow migration away from recruitment_assignments.
-- Historical assignment data is retained by the earlier migrations, but no
-- current recruitment record depends on the Assignment entity.

ALTER TABLE recruitment_screenings
    DROP CONSTRAINT IF EXISTS recruitment_screenings_assignment_id_fkey;

ALTER TABLE recruitment_screenings
    DROP COLUMN IF EXISTS assignment_id;

ALTER TABLE recruitment_submissions
    ADD COLUMN IF NOT EXISTS candidate_id INTEGER REFERENCES candidates(id),
    ADD COLUMN IF NOT EXISTS requirement_id INTEGER REFERENCES requirements(id);

UPDATE recruitment_submissions s
SET candidate_id = a.candidate_id,
    requirement_id = a.requirement_id
FROM recruitment_assignments a
WHERE s.assignment_id = a.id
  AND (s.candidate_id IS NULL OR s.requirement_id IS NULL);

ALTER TABLE recruitment_submissions
    ALTER COLUMN candidate_id SET NOT NULL,
    ALTER COLUMN requirement_id SET NOT NULL;

ALTER TABLE recruitment_submissions
    DROP CONSTRAINT IF EXISTS recruitment_submissions_assignment_id_fkey;

ALTER TABLE recruitment_submissions
    DROP COLUMN IF EXISTS assignment_id;

CREATE UNIQUE INDEX IF NOT EXISTS recruitment_submissions_candidate_requirement_unique
    ON recruitment_submissions(tenant_id, candidate_id, requirement_id);

CREATE INDEX IF NOT EXISTS idx_recruitment_submissions_candidate
    ON recruitment_submissions(tenant_id, candidate_id, submitted_at DESC);

CREATE INDEX IF NOT EXISTS idx_recruitment_submissions_candidate_requirement
    ON recruitment_submissions(tenant_id, candidate_id, requirement_id, submitted_at DESC);

ALTER TABLE recruitment_selections
    DROP CONSTRAINT IF EXISTS recruitment_selections_assignment_id_fkey;

ALTER TABLE recruitment_selections
    DROP COLUMN IF EXISTS assignment_id;

-- Screening and submission records are now independently scoped to the
-- Candidate × Requirement relationship. Assignment is no longer a runtime
-- dependency.
