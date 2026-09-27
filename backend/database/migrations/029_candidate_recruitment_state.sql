-- 029_candidate_recruitment_state.sql
-- Recruitment architecture correction:
-- current recruitment control state lives on candidates; history remains in
-- screening/interview/selection records. Recruitment Assignment is no longer
-- required by the new workflow APIs.

ALTER TABLE candidates
    ADD COLUMN IF NOT EXISTS screening_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS screening_limit INTEGER NOT NULL DEFAULT 3,
    ADD COLUMN IF NOT EXISTS interview_locked BOOLEAN NOT NULL DEFAULT FALSE;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'candidates' AND constraint_name = 'candidates_screening_count_valid'
    ) THEN
        ALTER TABLE candidates ADD CONSTRAINT candidates_screening_count_valid
            CHECK (screening_count >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'candidates' AND constraint_name = 'candidates_screening_limit_valid'
    ) THEN
        ALTER TABLE candidates ADD CONSTRAINT candidates_screening_limit_valid
            CHECK (screening_limit > 0);
    END IF;
END $$;

-- Preserve current state for existing data before the new gates become
-- authoritative. Existing assignments remain available as historical
-- compatibility data; new workflow code does not depend on them.
UPDATE candidates c
SET screening_count = COALESCE((
    SELECT COUNT(*)
    FROM recruitment_assignments a
    WHERE a.candidate_id = c.id
      AND a.tenant_id = c.tenant_id
      AND a.status = 'screening'
), 0);

UPDATE candidates c
SET interview_locked = EXISTS (
    SELECT 1
    FROM interviews i
    WHERE i.candidate_id = c.id
      AND i.tenant_id = c.tenant_id
      AND i.status IN ('scheduled', 'rescheduled')
);

ALTER TABLE recruitment_screenings
    ADD COLUMN IF NOT EXISTS candidate_id INTEGER REFERENCES candidates(id),
    ADD COLUMN IF NOT EXISTS requirement_id INTEGER REFERENCES requirements(id),
    ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'completed';

UPDATE recruitment_screenings s
SET candidate_id = a.candidate_id,
    requirement_id = a.requirement_id,
    status = CASE WHEN a.status = 'screening' THEN 'active' ELSE 'completed' END
FROM recruitment_assignments a
WHERE s.assignment_id = a.id
  AND (s.candidate_id IS NULL OR s.requirement_id IS NULL);

ALTER TABLE recruitment_screenings
    ALTER COLUMN candidate_id SET NOT NULL,
    ALTER COLUMN requirement_id SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'recruitment_screenings'
          AND constraint_name = 'recruitment_screening_status_valid'
    ) THEN
        ALTER TABLE recruitment_screenings ADD CONSTRAINT recruitment_screening_status_valid
            CHECK (status IN ('active', 'completed', 'rejected', 'withdrawn'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_recruitment_screenings_candidate
    ON recruitment_screenings(tenant_id, candidate_id, screened_at DESC);

CREATE INDEX IF NOT EXISTS idx_recruitment_screenings_candidate_requirement
    ON recruitment_screenings(tenant_id, candidate_id, requirement_id, screened_at DESC);

ALTER TABLE recruitment_selections
    ADD COLUMN IF NOT EXISTS candidate_id INTEGER REFERENCES candidates(id),
    ADD COLUMN IF NOT EXISTS requirement_id INTEGER REFERENCES requirements(id);

UPDATE recruitment_selections s
SET candidate_id = a.candidate_id,
    requirement_id = a.requirement_id
FROM recruitment_assignments a
WHERE s.assignment_id = a.id
  AND (s.candidate_id IS NULL OR s.requirement_id IS NULL);

ALTER TABLE recruitment_selections
    ALTER COLUMN candidate_id SET NOT NULL,
    ALTER COLUMN requirement_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS recruitment_selections_candidate_requirement_unique
    ON recruitment_selections(candidate_id, requirement_id);

CREATE INDEX IF NOT EXISTS idx_recruitment_selections_candidate
    ON recruitment_selections(tenant_id, candidate_id, decided_at DESC);
