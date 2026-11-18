-- Candidate job descriptions may contain complete role requirements and must not be truncated at 500 characters.
-- Keep this as a forward migration: the checksum-verified 001 baseline is immutable for existing tenant databases.
ALTER TABLE candidates
    ALTER COLUMN jobdescription TYPE TEXT;
