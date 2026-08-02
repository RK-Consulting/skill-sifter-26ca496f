-- 004_candidates_created_at.sql
-- Adds a timestamp to candidates, needed for the Dashboard's Recent Activity
-- feed to include real candidate-related events (docs/architecture.md, the
-- Dashboard mock-data bug fix). Without this, candidate events cannot be
-- included in a real, timestamp-sorted activity feed at all.

ALTER TABLE candidates ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();