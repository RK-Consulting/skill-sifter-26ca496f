-- 003_job_ownership.sql
-- Adds job ownership, needed to enforce RBAC (docs/architecture.md section 13.4):
-- jobs are a manager-owned artifact per the existing roles.permissions design
-- (manage_jobs belongs to manager/admin only). This also gives auto-JD-matching
-- settings (section 13.6) a well-defined per-user scope to act on.

ALTER TABLE jobs ADD COLUMN IF NOT EXISTS created_by_user_id INTEGER REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_jobs_created_by ON jobs(created_by_user_id);