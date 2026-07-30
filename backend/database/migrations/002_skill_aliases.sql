-- 002_skill_aliases.sql
-- Reference table for skill abbreviation/full-form matching (see docs/architecture.md §11, Decision 7).
-- Used at query time to expand a search term (e.g. "ML") to all its known
-- equivalent forms (e.g. "Machine Learning") before matching against
-- candidates.skills. Not used at extraction time — extraction stays freeform.
--
-- This table is expected to grow over time as new terminology enters common
-- usage. Add new terms via additional migration files (003_..., 004_..., etc.),
-- not by editing this file after it has shipped.

CREATE TABLE IF NOT EXISTS skill_aliases (
    id SERIAL PRIMARY KEY,
    canonical_term VARCHAR(255) NOT NULL,
    alias VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(alias)
);

CREATE INDEX IF NOT EXISTS idx_skill_aliases_canonical ON skill_aliases(canonical_term);
CREATE INDEX IF NOT EXISTS idx_skill_aliases_alias ON skill_aliases(alias);

-- Starter seed list. Not exhaustive — extend as real usage surfaces gaps.
INSERT INTO skill_aliases (canonical_term, alias, category) VALUES
    -- Languages
    ('JavaScript', 'JS', 'language'),
    ('TypeScript', 'TS', 'language'),
    ('Python', 'Py', 'language'),
    ('Golang', 'Go', 'language'),
    ('C Sharp', 'C#', 'language'),
    ('C Plus Plus', 'C++', 'language'),

    -- Infra / Cloud
    ('Kubernetes', 'K8s', 'infra'),
    ('Amazon Web Services', 'AWS', 'infra'),
    ('Google Cloud Platform', 'GCP', 'infra'),
    ('Microsoft Azure', 'Azure', 'infra'),
    ('Continuous Integration Continuous Deployment', 'CI/CD', 'infra'),
    ('Infrastructure as Code', 'IaC', 'infra'),
    ('Docker Compose', 'Compose', 'infra'),

    -- AI / ML
    ('Machine Learning', 'ML', 'ai_ml'),
    ('Artificial Intelligence', 'AI', 'ai_ml'),
    ('Natural Language Processing', 'NLP', 'ai_ml'),
    ('Large Language Model', 'LLM', 'ai_ml'),
    ('Retrieval Augmented Generation', 'RAG', 'ai_ml'),
    ('Deep Learning', 'DL', 'ai_ml'),
    ('Computer Vision', 'CV', 'ai_ml'),

    -- Databases
    ('Database', 'DB', 'database'),
    ('Relational Database Management System', 'RDBMS', 'database'),
    ('PostgreSQL', 'Postgres', 'database'),
    ('Structured Query Language', 'SQL', 'database'),

    -- Practices / Roles
    ('Quality Assurance', 'QA', 'practice'),
    ('User Interface', 'UI', 'practice'),
    ('User Experience', 'UX', 'practice'),
    ('Object Oriented Programming', 'OOP', 'practice'),
    ('Test Driven Development', 'TDD', 'practice'),
    ('Minimum Viable Product', 'MVP', 'practice'),
    ('Proof of Concept', 'POC', 'practice'),

    -- Web / Frameworks / APIs
    ('Application Programming Interface', 'API', 'web'),
    ('Representational State Transfer', 'REST', 'web'),
    ('Single Page Application', 'SPA', 'web'),
    ('Node.js', 'NodeJS', 'web'),

    -- Business / Misc tech
    ('Human Resources', 'HR', 'business'),
    ('Customer Relationship Management', 'CRM', 'business'),
    ('Enterprise Resource Planning', 'ERP', 'business'),
    ('Software as a Service', 'SaaS', 'business')
ON CONFLICT (alias) DO NOTHING;