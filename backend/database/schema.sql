
-- Create database (run this separately)
-- CREATE DATABASE skillsifter;

-- Connect to the database
-- \c skillsifter

-- Users Table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Candidates Table
CREATE TABLE IF NOT EXISTS candidates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    position VARCHAR(255),
    status VARCHAR(50) DEFAULT 'New',
    date_applied TIMESTAMP NOT NULL DEFAULT NOW(),
    resume_url TEXT,
    cover_letter TEXT,
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Jobs Table
CREATE TABLE IF NOT EXISTS jobs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    department VARCHAR(100),
    location VARCHAR(100),
    status VARCHAR(50) DEFAULT 'Open',
    date_posted TIMESTAMP NOT NULL DEFAULT NOW(),
    description TEXT,
    requirements TEXT,
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Daily Jobs Table
CREATE TABLE IF NOT EXISTS daily_jobs (
    id SERIAL PRIMARY KEY,
    jd_no INTEGER NOT NULL,
    instructions TEXT NOT NULL,
    assigned_user INTEGER NOT NULL REFERENCES users(id),
    assigned_date TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Interviews Table
CREATE TABLE IF NOT EXISTS interviews (
    id SERIAL PRIMARY KEY,
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    candidate_name VARCHAR(255) NOT NULL,
    position VARCHAR(255) NOT NULL,
    interview_date TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'Scheduled',
    feedback TEXT,
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Business Contacts Table
CREATE TABLE IF NOT EXISTS business_contacts (
    id SERIAL PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    industry VARCHAR(100),
    notes TEXT,
    last_contact_date TIMESTAMP,
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Triggers for updating last_modified timestamp
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_modified = NOW();
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

-- Create triggers for each table
CREATE TRIGGER update_users_modtime
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_candidates_modtime
BEFORE UPDATE ON candidates
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_jobs_modtime
BEFORE UPDATE ON jobs
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_daily_jobs_modtime
BEFORE UPDATE ON daily_jobs
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_interviews_modtime
BEFORE UPDATE ON interviews
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_business_contacts_modtime
BEFORE UPDATE ON business_contacts
FOR EACH ROW EXECUTE FUNCTION update_modified_column();
