
-- Create candidates table
CREATE TABLE IF NOT EXISTS candidates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    position VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    date_applied TIMESTAMP NOT NULL DEFAULT NOW(),
    resume_url TEXT,
    cover_letter TEXT,
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    department VARCHAR(100) NOT NULL,
    location VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    date_posted TIMESTAMP NOT NULL DEFAULT NOW(),
    description TEXT,
    requirements TEXT,
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create daily_jobs table
CREATE TABLE IF NOT EXISTS daily_jobs (
    id SERIAL PRIMARY KEY,
    jd_no INTEGER NOT NULL,
    instructions TEXT NOT NULL,
    assigned_user INTEGER NOT NULL,
    assigned_date TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create interviews table
CREATE TABLE IF NOT EXISTS interviews (
    id SERIAL PRIMARY KEY,
    candidate_id INTEGER NOT NULL,
    candidate_name VARCHAR(255) NOT NULL,
    position VARCHAR(255) NOT NULL,
    interview_date TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL,
    feedback VARCHAR(50) NOT NULL,
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert sample data for daily_jobs
INSERT INTO daily_jobs (jd_no, instructions, assigned_user, assigned_date)
VALUES 
(1001, 'Source candidates for Senior Java Developer position', 1, NOW() - INTERVAL '2 days'),
(1002, 'Review resumes for Frontend Developer candidates', 2, NOW() - INTERVAL '1 day'),
(1003, 'Prepare interview questions for DevOps Engineer', 1, NOW()),
(1004, 'Follow up with candidates from yesterday interviews', 3, NOW() - INTERVAL '4 hours'),
(1005, 'Update job descriptions for open positions', 2, NOW() - INTERVAL '3 days');

-- Insert sample data for interviews
INSERT INTO interviews (candidate_id, candidate_name, position, interview_date, status, feedback)
VALUES 
(1, 'John Smith', 'Senior Java Developer', NOW() + INTERVAL '1 day', 'Scheduled', 'Pending'),
(2, 'Jane Doe', 'Frontend Developer', NOW() - INTERVAL '2 days', 'Completed', 'Selected'),
(3, 'Michael Johnson', 'DevOps Engineer', NOW() - INTERVAL '1 day', 'Completed', 'Rejected'),
(4, 'Emily Williams', 'UX Designer', NOW() + INTERVAL '3 days', 'Scheduled', 'Pending'),
(5, 'Robert Brown', 'Product Manager', NOW(), 'Scheduled', 'Pending');
