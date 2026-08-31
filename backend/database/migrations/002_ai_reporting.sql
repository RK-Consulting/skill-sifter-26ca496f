-- SkillSifter reporting + local AI resume ingestion
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();

CREATE TABLE IF NOT EXISTS resumes (
    id SERIAL PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    candidate_id INTEGER REFERENCES candidates(id) ON DELETE SET NULL,
    file_name VARCHAR(500) NOT NULL,
    file_path TEXT NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    mime_type VARCHAR(120),
    extracted_text TEXT,
    parsing_status VARCHAR(30) NOT NULL DEFAULT 'pending',
    parser_model VARCHAR(120),
    parse_error TEXT,
    uploaded_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    parsed_at TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_resumes_company_hash ON resumes(company_name, file_hash);
CREATE INDEX IF NOT EXISTS idx_resumes_company ON resumes(company_name);
CREATE INDEX IF NOT EXISTS idx_resumes_status ON resumes(company_name, parsing_status);
CREATE INDEX IF NOT EXISTS idx_resumes_candidate ON resumes(candidate_id);

CREATE TABLE IF NOT EXISTS skills (
    id SERIAL PRIMARY KEY,
    name VARCHAR(160) NOT NULL UNIQUE,
    normalized_name VARCHAR(160) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS candidate_skills (
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    skill_id INTEGER NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    confidence NUMERIC(5,4),
    source VARCHAR(30) NOT NULL DEFAULT 'resume_ai',
    PRIMARY KEY(candidate_id, skill_id)
);
CREATE INDEX IF NOT EXISTS idx_candidate_skills_skill ON candidate_skills(skill_id);

CREATE TABLE IF NOT EXISTS activity_logs (
    id BIGSERIAL PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    actor_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(80) NOT NULL,
    entity_type VARCHAR(80) NOT NULL,
    entity_id VARCHAR(80),
    description TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_activity_company_time ON activity_logs(company_name, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_action ON activity_logs(company_name, action);
CREATE INDEX IF NOT EXISTS idx_activity_entity ON activity_logs(company_name, entity_type, entity_id);

CREATE TABLE IF NOT EXISTS resume_search_logs (
    id BIGSERIAL PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    actor_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    query_text TEXT NOT NULL,
    resumes_searched INTEGER NOT NULL DEFAULT 0,
    results_count INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_resume_search_company_time ON resume_search_logs(company_name, created_at DESC);

CREATE OR REPLACE FUNCTION skillsifter_activity_trigger() RETURNS trigger AS $$
DECLARE company TEXT; entity TEXT; entity_id_text TEXT; payload JSONB;
BEGIN
    company := COALESCE(NEW.company_name, OLD.company_name);
    entity := TG_TABLE_NAME; entity_id_text := COALESCE(NEW.id, OLD.id)::text;
    IF TG_OP = 'INSERT' THEN payload := to_jsonb(NEW);
    ELSIF TG_OP = 'UPDATE' THEN payload := jsonb_build_object('before', to_jsonb(OLD), 'after', to_jsonb(NEW));
    ELSE payload := to_jsonb(OLD); END IF;
    INSERT INTO activity_logs(company_name, action, entity_type, entity_id, description, metadata)
    VALUES(company, upper(TG_TABLE_NAME) || '_' || upper(TG_OP), entity, entity_id_text, upper(TG_OP) || ' on ' || entity, payload);
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_candidates_activity ON candidates;
CREATE TRIGGER trg_candidates_activity AFTER INSERT OR UPDATE OR DELETE ON candidates FOR EACH ROW EXECUTE FUNCTION skillsifter_activity_trigger();
DROP TRIGGER IF EXISTS trg_jobs_activity ON jobs;
CREATE TRIGGER trg_jobs_activity AFTER INSERT OR UPDATE OR DELETE ON jobs FOR EACH ROW EXECUTE FUNCTION skillsifter_activity_trigger();
DROP TRIGGER IF EXISTS trg_daily_jobs_activity ON daily_jobs;
CREATE TRIGGER trg_daily_jobs_activity AFTER INSERT OR UPDATE OR DELETE ON daily_jobs FOR EACH ROW EXECUTE FUNCTION skillsifter_activity_trigger();
DROP TRIGGER IF EXISTS trg_interviews_activity ON interviews;
CREATE TRIGGER trg_interviews_activity AFTER INSERT OR UPDATE OR DELETE ON interviews FOR EACH ROW EXECUTE FUNCTION skillsifter_activity_trigger();
DROP TRIGGER IF EXISTS trg_business_dev_activity ON business_dev;
CREATE TRIGGER trg_business_dev_activity AFTER INSERT OR UPDATE OR DELETE ON business_dev FOR EACH ROW EXECUTE FUNCTION skillsifter_activity_trigger();
DROP TRIGGER IF EXISTS trg_resumes_activity ON resumes;
CREATE TRIGGER trg_resumes_activity AFTER INSERT OR UPDATE OR DELETE ON resumes FOR EACH ROW EXECUTE FUNCTION skillsifter_activity_trigger();
