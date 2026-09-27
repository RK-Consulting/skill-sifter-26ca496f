-- 022_resume_ai_candidate_intelligence.sql
-- RAI-03: structured candidate intelligence derived from Resume AI.
-- Historical schema definitions remain immutable; all RAI-03 changes are forward-only.

ALTER TABLE candidate_language_expertise
    ADD COLUMN IF NOT EXISTS source_resume_id INTEGER REFERENCES resumes(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_candidate_language_expertise_source_resume
    ON candidate_language_expertise(source_resume_id);

CREATE TABLE IF NOT EXISTS candidate_professional_profiles (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    source_resume_id INTEGER REFERENCES resumes(id) ON DELETE SET NULL,
    current_title VARCHAR(255),
    professional_summary TEXT,
    location VARCHAR(255),
    total_experience VARCHAR(100),
    relevant_experience VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT candidate_professional_profiles_candidate_unique
        UNIQUE (tenant_id, candidate_id)
);

CREATE INDEX IF NOT EXISTS idx_candidate_professional_profiles_tenant_candidate
    ON candidate_professional_profiles(tenant_id, candidate_id);
CREATE INDEX IF NOT EXISTS idx_candidate_professional_profiles_source_resume
    ON candidate_professional_profiles(source_resume_id);

CREATE TABLE IF NOT EXISTS candidate_employment_history (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    source_resume_id INTEGER NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    employer VARCHAR(255) NOT NULL,
    job_title VARCHAR(255),
    start_date DATE,
    end_date DATE,
    start_year INTEGER,
    end_year INTEGER,
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT candidate_employment_history_year_valid
        CHECK (
            (start_year IS NULL OR start_year BETWEEN 1900 AND 2200)
            AND (end_year IS NULL OR end_year BETWEEN 1900 AND 2200)
        )
);

CREATE INDEX IF NOT EXISTS idx_candidate_employment_history_tenant_candidate
    ON candidate_employment_history(tenant_id, candidate_id);
CREATE INDEX IF NOT EXISTS idx_candidate_employment_history_source_resume
    ON candidate_employment_history(source_resume_id);

CREATE TABLE IF NOT EXISTS candidate_education (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    source_resume_id INTEGER NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    institution VARCHAR(255) NOT NULL,
    degree VARCHAR(255),
    field_of_study VARCHAR(255),
    start_date DATE,
    end_date DATE,
    start_year INTEGER,
    end_year INTEGER,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT candidate_education_year_valid
        CHECK (
            (start_year IS NULL OR start_year BETWEEN 1900 AND 2200)
            AND (end_year IS NULL OR end_year BETWEEN 1900 AND 2200)
        )
);

CREATE INDEX IF NOT EXISTS idx_candidate_education_tenant_candidate
    ON candidate_education(tenant_id, candidate_id);
CREATE INDEX IF NOT EXISTS idx_candidate_education_source_resume
    ON candidate_education(source_resume_id);

CREATE TABLE IF NOT EXISTS candidate_certifications (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    source_resume_id INTEGER NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    issuer VARCHAR(255),
    issue_date DATE,
    expiry_date DATE,
    issue_year INTEGER,
    expiry_year INTEGER,
    credential_reference VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT candidate_certifications_year_valid
        CHECK (
            (issue_year IS NULL OR issue_year BETWEEN 1900 AND 2200)
            AND (expiry_year IS NULL OR expiry_year BETWEEN 1900 AND 2200)
        )
);

CREATE INDEX IF NOT EXISTS idx_candidate_certifications_tenant_candidate
    ON candidate_certifications(tenant_id, candidate_id);
CREATE INDEX IF NOT EXISTS idx_candidate_certifications_source_resume
    ON candidate_certifications(source_resume_id);

CREATE TABLE IF NOT EXISTS candidate_projects (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    source_resume_id INTEGER NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    project_name VARCHAR(255) NOT NULL,
    description TEXT,
    role VARCHAR(255),
    technologies TEXT[] NOT NULL DEFAULT '{}',
    start_date DATE,
    end_date DATE,
    start_year INTEGER,
    end_year INTEGER,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT candidate_projects_year_valid
        CHECK (
            (start_year IS NULL OR start_year BETWEEN 1900 AND 2200)
            AND (end_year IS NULL OR end_year BETWEEN 1900 AND 2200)
        )
);

CREATE INDEX IF NOT EXISTS idx_candidate_projects_tenant_candidate
    ON candidate_projects(tenant_id, candidate_id);
CREATE INDEX IF NOT EXISTS idx_candidate_projects_source_resume
    ON candidate_projects(source_resume_id);
