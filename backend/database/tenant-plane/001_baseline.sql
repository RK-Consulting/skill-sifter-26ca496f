-- SkillSifter final tenant-plane baseline.
-- One database = one customer.
-- No control-plane references. No companies table. No legacy Jobs/Assignment tables.
-- Business workflow remains application-owned; PostgreSQL supplies structural
-- integrity and concurrency arbitration.

CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    permissions JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, permissions) VALUES
    ('admin', '["all"]'),
    ('manager', '["manage_candidates","manage_requirements","manage_interviews","view_reports"]'),
    ('recruiter', '["view_candidates","add_candidates","view_requirements","schedule_interviews"]'),
    ('team_leader', '["view_candidates","add_candidates","view_requirements","manage_team"]')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL REFERENCES roles(name),
    tenant_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_role ON users(role);

CREATE TABLE clients (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'prospect',
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    partner_name VARCHAR(255),
    contact_person VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT clients_status_valid CHECK (status IN ('prospect','active','inactive'))
);

CREATE INDEX idx_clients_tenant ON clients(tenant_id);

CREATE TABLE requirements (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    client_id INTEGER NOT NULL REFERENCES clients(id),
    job_type VARCHAR(30) NOT NULL,
    title VARCHAR(255) NOT NULL,
    department VARCHAR(100),
    experience_required VARCHAR(100),
    budget VARCHAR(255),
    language_requirement VARCHAR(255),
    certifications_required TEXT,
    notice_period VARCHAR(100),
    work_arrangement VARCHAR(30),
    mandatory_requirements TEXT,
    description TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'open',
    location VARCHAR(255),
    headcount INTEGER NOT NULL DEFAULT 1,
    opened_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT requirements_status_valid CHECK (status IN ('open','closed','on_hold','cancelled')),
    CONSTRAINT requirements_job_type_valid CHECK (job_type IN ('fulltime','contract')),
    CONSTRAINT requirements_work_arrangement_valid CHECK (work_arrangement IN ('hybrid','remote','office')),
    CONSTRAINT requirements_headcount_positive CHECK (headcount > 0)
);

CREATE INDEX idx_requirements_tenant ON requirements(tenant_id);
CREATE INDEX idx_requirements_client ON requirements(client_id);
CREATE INDEX idx_requirements_status ON requirements(status);

CREATE TABLE candidates (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    position VARCHAR(100),
    location VARCHAR(255),
    experience VARCHAR(100),
    currentctc VARCHAR(100),
    expectedctc VARCHAR(100),
    noticeperiod VARCHAR(100),
    jobdescription VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    pipeline_stage VARCHAR(50) NOT NULL DEFAULT 'new',
    screening_count INTEGER NOT NULL DEFAULT 0,
    screening_limit INTEGER NOT NULL DEFAULT 3,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT candidates_status_valid CHECK (status IN ('active','inactive','blacklisted','archived')),
    CONSTRAINT candidates_screening_count_valid CHECK (screening_count >= 0),
    CONSTRAINT candidates_screening_limit_valid CHECK (screening_limit > 0)
);

CREATE INDEX idx_candidates_tenant ON candidates(tenant_id);
CREATE INDEX idx_candidates_email ON candidates(tenant_id, email);

CREATE TABLE candidate_language_expertise (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    language VARCHAR(100) NOT NULL,
    proficiency_framework VARCHAR(50) NOT NULL,
    proficiency_level VARCHAR(50) NOT NULL,
    source_resume_id INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT candidate_language_expertise_unique UNIQUE(candidate_id, language, proficiency_framework, proficiency_level)
);

CREATE INDEX idx_candidate_language_expertise_tenant_candidate
    ON candidate_language_expertise(tenant_id, candidate_id);
CREATE INDEX idx_candidate_language_expertise_source_resume
    ON candidate_language_expertise(source_resume_id);

CREATE TABLE candidate_expertise (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    skill VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL,
    proficiency_level VARCHAR(50) NOT NULL,
    source_resume_id INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT candidate_expertise_unique UNIQUE(candidate_id, skill, category)
);

CREATE INDEX idx_candidate_expertise_tenant_candidate
    ON candidate_expertise(tenant_id, candidate_id);

CREATE TABLE resumes (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
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

CREATE UNIQUE INDEX uq_resumes_tenant_hash ON resumes(tenant_id, file_hash);
CREATE INDEX idx_resumes_tenant ON resumes(tenant_id);
CREATE INDEX idx_resumes_status ON resumes(tenant_id, parsing_status);
CREATE INDEX idx_resumes_candidate ON resumes(candidate_id);

ALTER TABLE candidate_language_expertise
    ADD CONSTRAINT candidate_language_expertise_source_resume_fk
    FOREIGN KEY (source_resume_id) REFERENCES resumes(id) ON DELETE SET NULL;

ALTER TABLE candidate_expertise
    ADD CONSTRAINT candidate_expertise_source_resume_fk
    FOREIGN KEY (source_resume_id) REFERENCES resumes(id) ON DELETE SET NULL;

CREATE TABLE resume_search_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    actor_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    query_text TEXT NOT NULL,
    resumes_searched INTEGER NOT NULL DEFAULT 0,
    results_count INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_resume_search_tenant_time ON resume_search_logs(tenant_id, created_at DESC);

CREATE TABLE candidate_professional_profiles (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    source_resume_id INTEGER,
    current_title VARCHAR(255),
    professional_summary TEXT,
    location VARCHAR(255),
    total_experience VARCHAR(100),
    relevant_experience VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT candidate_professional_profiles_candidate_unique UNIQUE(tenant_id, candidate_id)
);

CREATE INDEX idx_candidate_professional_profiles_source_resume
    ON candidate_professional_profiles(source_resume_id);

CREATE TABLE candidate_employment_history (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
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
    CONSTRAINT candidate_employment_history_year_valid CHECK (
        (start_year IS NULL OR start_year BETWEEN 1900 AND 2200) AND
        (end_year IS NULL OR end_year BETWEEN 1900 AND 2200)
    )
);

CREATE INDEX idx_candidate_employment_history_candidate
    ON candidate_employment_history(tenant_id, candidate_id);

CREATE TABLE candidate_education (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
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
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_candidate_education_candidate ON candidate_education(tenant_id, candidate_id);

CREATE TABLE candidate_certifications (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
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
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_candidate_certifications_candidate ON candidate_certifications(tenant_id, candidate_id);

CREATE TABLE candidate_projects (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
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
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_candidate_projects_candidate ON candidate_projects(tenant_id, candidate_id);

CREATE TABLE interviews (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    candidate_name VARCHAR(255) NOT NULL,
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    position VARCHAR(100),
    round INTEGER NOT NULL DEFAULT 1,
    interview_date TIMESTAMP NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'scheduled',
    outcome VARCHAR(100),
    feedback TEXT,
    candidate_feedback TEXT,
    next_action TEXT,
    last_modified TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT interviews_round_positive CHECK (round > 0),
    CONSTRAINT interviews_status_valid CHECK (status IN ('scheduled','completed','cancelled','rescheduled','no_show'))
);

CREATE INDEX idx_interviews_tenant_requirement_round
    ON interviews(tenant_id, requirement_id, round, interview_date DESC);
CREATE INDEX idx_interviews_tenant_candidate
    ON interviews(tenant_id, candidate_id, interview_date DESC);

CREATE UNIQUE INDEX uq_interviews_active_candidate_requirement
    ON interviews(tenant_id, candidate_id, requirement_id)
    WHERE status IN ('scheduled','rescheduled');

CREATE TABLE recruitment_screenings (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    recruiter_user_id INTEGER NOT NULL REFERENCES users(id),
    current_ctc VARCHAR(255),
    expected_ctc VARCHAR(255),
    notice_period VARCHAR(255),
    last_working_day DATE,
    current_location VARCHAR(255),
    willing_to_relocate BOOLEAN,
    preferred_location VARCHAR(255),
    reason_for_change TEXT,
    offers_in_hand TEXT,
    candidate_interest VARCHAR(50),
    availability_date DATE,
    relevant_experience TEXT,
    recruiter_assessment TEXT,
    notes TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'completed',
    screened_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT recruitment_screening_status_valid CHECK (status IN ('active','completed','rejected','withdrawn'))
);

CREATE INDEX idx_recruitment_screenings_candidate
    ON recruitment_screenings(tenant_id, candidate_id, screened_at DESC);
CREATE INDEX idx_recruitment_screenings_candidate_requirement
    ON recruitment_screenings(tenant_id, candidate_id, requirement_id, screened_at DESC);

CREATE TABLE recruitment_submissions (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    submitted_by_user_id INTEGER NOT NULL REFERENCES users(id),
    recipient_type VARCHAR(50) NOT NULL,
    recipient_client_id INTEGER REFERENCES clients(id),
    recipient_name VARCHAR(255),
    recipient_email VARCHAR(255),
    submission_context TEXT,
    recruiter_notes TEXT,
    candidate_snapshot JSONB NOT NULL,
    requirement_snapshot JSONB NOT NULL,
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT recruitment_submissions_recipient_type_valid CHECK (recipient_type = 'client'),
    CONSTRAINT recruitment_submissions_recipient_valid CHECK (recipient_client_id IS NOT NULL)
);

CREATE UNIQUE INDEX recruitment_submissions_candidate_requirement_unique
    ON recruitment_submissions(tenant_id, candidate_id, requirement_id);
CREATE INDEX idx_recruitment_submissions_candidate
    ON recruitment_submissions(tenant_id, candidate_id, submitted_at DESC);
CREATE INDEX idx_recruitment_submissions_client
    ON recruitment_submissions(recipient_client_id);

CREATE TABLE recruitment_submission_feedback (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    submission_id INTEGER NOT NULL REFERENCES recruitment_submissions(id) ON DELETE CASCADE,
    feedback_by_user_id INTEGER NOT NULL REFERENCES users(id),
    outcome VARCHAR(50) NOT NULL,
    reason_code VARCHAR(100),
    comments TEXT,
    next_action TEXT,
    feedback_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT recruitment_submission_feedback_outcome_valid CHECK (outcome IN ('shortlist','hold','reject'))
);

CREATE INDEX idx_recruitment_submission_feedback_submission
    ON recruitment_submission_feedback(submission_id, feedback_at DESC);

CREATE TABLE recruitment_selections (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    decision VARCHAR(20) NOT NULL,
    decision_notes TEXT,
    next_action TEXT,
    decided_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT recruitment_selection_decision_valid CHECK (decision IN ('selected','rejected')),
    CONSTRAINT recruitment_selection_candidate_requirement_unique UNIQUE(candidate_id, requirement_id)
);

CREATE INDEX idx_recruitment_selections_tenant_decided
    ON recruitment_selections(tenant_id, decided_at DESC);

CREATE TABLE recruitment_offers (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    selection_id INTEGER NOT NULL REFERENCES recruitment_selections(id),
    accepted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_recruitment_offers_tenant_pair
    ON recruitment_offers(tenant_id, candidate_id, requirement_id);

CREATE TABLE recruitment_joinings (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    offer_id INTEGER NOT NULL REFERENCES recruitment_offers(id),
    joining_date DATE,
    joined BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT recruitment_joining_date_required CHECK (NOT joined OR joining_date IS NOT NULL)
);

CREATE UNIQUE INDEX uq_recruitment_joinings_tenant_pair
    ON recruitment_joinings(tenant_id, candidate_id, requirement_id);

CREATE TABLE recruitment_billings (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    client_id INTEGER NOT NULL REFERENCES clients(id),
    joining_id INTEGER NOT NULL REFERENCES recruitment_joinings(id),
    billing_date DATE NOT NULL,
    amount NUMERIC(14,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    invoice_reference VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT recruitment_billings_amount_positive CHECK (amount > 0),
    CONSTRAINT recruitment_billings_currency_valid CHECK (currency = upper(currency))
);

CREATE UNIQUE INDEX uq_recruitment_billings_tenant_pair
    ON recruitment_billings(tenant_id, candidate_id, requirement_id);

CREATE INDEX idx_recruitment_billings_client
    ON recruitment_billings(tenant_id, client_id);

CREATE TABLE audit_events (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    actor_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id INTEGER NOT NULL,
    action VARCHAR(100) NOT NULL,
    occurred_at TIMESTAMP NOT NULL DEFAULT NOW(),
    correlation_id VARCHAR(100),
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB
);

CREATE INDEX idx_audit_events_tenant ON audit_events(tenant_id);
CREATE INDEX idx_audit_events_entity ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_events_correlation ON audit_events(correlation_id);
