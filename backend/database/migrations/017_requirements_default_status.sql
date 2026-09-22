-- 017_requirements_default_status.sql
-- Align the database default with the authoritative Requirement lifecycle.
-- Migration 015 normalized existing rows and constrained valid statuses,
-- but the historical default remained 'draft'.

ALTER TABLE requirements
    ALTER COLUMN status SET DEFAULT 'open';
