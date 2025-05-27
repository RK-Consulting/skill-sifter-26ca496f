
-- Create companies/tenants table
CREATE TABLE IF NOT EXISTS companies (
    id VARCHAR(255) PRIMARY KEY,
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

-- Create users table with role and company name (changed from company_id)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(100) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create candidates table
CREATE TABLE IF NOT EXISTS candidates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    position VARCHAR(100),
    status VARCHAR(50) DEFAULT 'applied',
    source VARCHAR(100),
    date_applied TIMESTAMP DEFAULT NOW(),
    resume_url TEXT,
    cover_letter TEXT,
    last_modified TIMESTAMP DEFAULT NOW(),
    company_name VARCHAR(255) NOT NULL
);

-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    department VARCHAR(100),
    location VARCHAR(100),
    status VARCHAR(50) DEFAULT 'open',
    date_posted TIMESTAMP DEFAULT NOW(),
    description TEXT,
    requirements TEXT,
    last_modified TIMESTAMP DEFAULT NOW(),
    company_name VARCHAR(255) NOT NULL
);

-- Create daily_jobs table
CREATE TABLE IF NOT EXISTS daily_jobs (
    id SERIAL PRIMARY KEY,
    jd_no INTEGER NOT NULL,
    instructions TEXT,
    assigned_user INTEGER REFERENCES users(id),
    assigned_date TIMESTAMP DEFAULT NOW(),
    last_modified TIMESTAMP DEFAULT NOW(),
    company_name VARCHAR(255) NOT NULL
);

-- Create interviews table
CREATE TABLE IF NOT EXISTS interviews (
    id SERIAL PRIMARY KEY,
    candidate_id INTEGER REFERENCES candidates(id),
    candidate_name VARCHAR(255) NOT NULL,
    position VARCHAR(100),
    interview_date TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'scheduled',
    feedback TEXT,
    last_modified TIMESTAMP DEFAULT NOW(),
    company_name VARCHAR(255) NOT NULL
);

-- Create business_dev table (recreate to ensure all fields are present)
DROP TABLE IF EXISTS business_dev;
CREATE TABLE business_dev (
    id SERIAL PRIMARY KEY,
    client_name VARCHAR(255) NOT NULL,
    partner_name VARCHAR(255),
    contact_person VARCHAR(255) NOT NULL,
    contact_number VARCHAR(50),
    contact_email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_modified TIMESTAMP DEFAULT NOW(),
    company_name VARCHAR(255) NOT NULL
);

-- Drop unused index
DROP INDEX IF EXISTS idx_users_role;
-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_candidates_company ON candidates(company_name);
CREATE INDEX IF NOT EXISTS idx_jobs_company ON jobs(company_name);
CREATE INDEX IF NOT EXISTS idx_daily_jobs_company ON daily_jobs(company_name);
CREATE INDEX IF NOT EXISTS idx_interviews_company ON interviews(company_name);
CREATE INDEX IF NOT EXISTS idx_users_company ON users(company_name);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_business_dev_company ON business_dev(company_name);
