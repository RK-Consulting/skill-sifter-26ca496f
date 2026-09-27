-- 025_recruitment_submissions.sql
-- Phase 3: formal candidate submission to a client or hiring manager.
CREATE TABLE IF NOT EXISTS recruitment_submissions (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    assignment_id INTEGER NOT NULL REFERENCES recruitment_assignments(id) ON DELETE CASCADE,
    submitted_by_user_id INTEGER NOT NULL REFERENCES users(id),
    recipient_type VARCHAR(50) NOT NULL,
    recipient_client_id INTEGER REFERENCES clients(id),
    recipient_user_id INTEGER REFERENCES users(id),
    recipient_name VARCHAR(255),
    recipient_email VARCHAR(255),
    submission_context TEXT,
    recruiter_notes TEXT,
    candidate_snapshot JSONB NOT NULL,
    requirement_snapshot JSONB NOT NULL,
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT recruitment_submissions_recipient_type_valid
        CHECK (recipient_type IN ('client', 'hiring_manager')),
    CONSTRAINT recruitment_submissions_recipient_valid
        CHECK (
            (recipient_type = 'client' AND recipient_client_id IS NOT NULL)
            OR (recipient_type = 'hiring_manager' AND recipient_user_id IS NOT NULL)
        )
);
CREATE INDEX IF NOT EXISTS idx_recruitment_submissions_tenant ON recruitment_submissions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_submissions_assignment ON recruitment_submissions(assignment_id, submitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_recruitment_submissions_client ON recruitment_submissions(recipient_client_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_submissions_recipient_user ON recruitment_submissions(recipient_user_id);
