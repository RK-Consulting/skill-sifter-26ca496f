
-- Create companies/tenants table
CREATE TABLE IF NOT EXISTS companies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create roles table with predefined roles
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    permissions JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default roles
INSERT INTO roles (name, permissions, created_at) VALUES 
('admin', '["all"]', NOW()),
('manager', '["manage_candidates", "manage_jobs", "manage_interviews", "view_reports"]', NOW()),
('recruiter', '["view_candidates", "add_candidates", "view_jobs", "schedule_interviews"]', NOW()),
('team_leader', '["view_candidates", "add_candidates", "view_jobs", "manage_team"]', NOW())
ON CONFLICT (name) DO NOTHING;

-- Modify users table to include role and company
DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role_id INTEGER NOT NULL REFERENCES roles(id),
    company_id INTEGER NOT NULL REFERENCES companies(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Modify existing tables to add company_id

-- Modify candidates table
ALTER TABLE candidates 
ADD COLUMN IF NOT EXISTS company_id INTEGER REFERENCES companies(id);

-- Modify jobs table
ALTER TABLE jobs 
ADD COLUMN IF NOT EXISTS company_id INTEGER REFERENCES companies(id);

-- Modify daily_jobs table
ALTER TABLE daily_jobs 
ADD COLUMN IF NOT EXISTS company_id INTEGER REFERENCES companies(id);

-- Modify interviews table
ALTER TABLE interviews 
ADD COLUMN IF NOT EXISTS company_id INTEGER REFERENCES companies(id);

-- First-time setup company
INSERT INTO companies (name, created_at) 
VALUES ('Default Company', NOW())
ON CONFLICT DO NOTHING;

-- Update existing records to belong to the default company
UPDATE candidates SET company_id = 1 WHERE company_id IS NULL;
UPDATE jobs SET company_id = 1 WHERE company_id IS NULL;
UPDATE daily_jobs SET company_id = 1 WHERE company_id IS NULL;
UPDATE interviews SET company_id = 1 WHERE company_id IS NULL;

-- Make company_id required for all tables
ALTER TABLE candidates ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE jobs ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE daily_jobs ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE interviews ALTER COLUMN company_id SET NOT NULL;

-- Add any missing indexes
CREATE INDEX IF NOT EXISTS idx_candidates_company ON candidates(company_id);
CREATE INDEX IF NOT EXISTS idx_jobs_company ON jobs(company_id);
CREATE INDEX IF NOT EXISTS idx_daily_jobs_company ON daily_jobs(company_id);
CREATE INDEX IF NOT EXISTS idx_interviews_company ON interviews(company_id);
CREATE INDEX IF NOT EXISTS idx_users_company ON users(company_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role_id);
