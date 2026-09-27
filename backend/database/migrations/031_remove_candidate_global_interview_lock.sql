-- 031_remove_candidate_global_interview_lock.sql
-- Interview concurrency is scoped to Candidate × Requirement. A candidate may
-- legitimately interview for multiple client requirements at the same time.
ALTER TABLE candidates DROP COLUMN IF EXISTS interview_locked;