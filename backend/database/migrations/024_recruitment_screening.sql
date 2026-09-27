-- 024_recruitment_screening.sql
-- Recruitment workflow Phase 2: recruiter screening evidence.
-- Phase 2: recruiter screening and candidate enrichment evidence.
--
-- Screening answers belong to the candidate-requirement transaction, not
-- Candidate master. Multiple screening records preserve every interaction.

CREATE TABLE IF NOT EXISTS recruitment_screenings (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    assignment_id INTEGER NOT NULL REFERENCES recruitment_assignments(id) ON DELETE CASCADE,
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
    screened_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recruitment_screenings_tenant
    ON recruitment_screenings(tenant_id);

CREATE INDEX IF NOT EXISTS idx_recruitment_screenings_assignment
    ON recruitment_screenings(assignment_id, screened_at DESC);

CREATE INDEX IF NOT EXISTS idx_recruitment_screenings_recruiter
    ON recruitment_screenings(recruiter_user_id);
