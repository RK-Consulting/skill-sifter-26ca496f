-- 009_candidate_expertise.sql
-- CP7B: Candidate Status & Generic Expertise
-- New Candidate expertise model.
-- No legacy compatibility or data migration.

-- ============================================================
-- 1. Remove obsolete Candidate columns
-- ============================================================

ALTER TABLE candidates DROP COLUMN IF EXISTS jlptlanguage;
ALTER TABLE candidates DROP COLUMN IF EXISTS skills;

-- ============================================================
-- 2. Candidate status constraint
-- ============================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'candidates'
          AND constraint_name = 'candidates_status_valid'
    ) THEN
        ALTER TABLE candidates
            ADD CONSTRAINT candidates_status_valid
            CHECK (status IN (
                'active',
                'inactive',
                'blacklisted',
                'archived'
            ));
    END IF;
END $$;

-- ============================================================
-- 3. Candidate Language Expertise
-- ============================================================

CREATE TABLE IF NOT EXISTS candidate_language_expertise (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    language VARCHAR(100) NOT NULL,
    proficiency_framework VARCHAR(50) NOT NULL,
    proficiency_level VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT candidate_language_expertise_unique
        UNIQUE (
            candidate_id,
            language,
            proficiency_framework,
            proficiency_level
        )
);

CREATE INDEX IF NOT EXISTS idx_candidate_language_expertise_tenant_candidate
    ON candidate_language_expertise(tenant_id, candidate_id);

-- ============================================================
-- 4. Candidate Technical Expertise
-- ============================================================

CREATE TABLE IF NOT EXISTS candidate_expertise (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    skill VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL,
    proficiency_level VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT candidate_expertise_unique
        UNIQUE (
            candidate_id,
            skill,
            category
        )
);

CREATE INDEX IF NOT EXISTS idx_candidate_expertise_tenant_candidate
    ON candidate_expertise(tenant_id, candidate_id);
