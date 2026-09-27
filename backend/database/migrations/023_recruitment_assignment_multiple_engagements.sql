-- 023_recruitment_assignment_multiple_engagements.sql
-- Recruitment workflow correction.
--
-- A candidate may participate in multiple recruitment assignments at the
-- same time. The durable transaction boundary is the candidate/requirement
-- pair, which remains unique in recruitment_assignments.
--
-- The previous active_recruitment_engagements flag encoded a single-active-
-- engagement rule that conflicts with the accepted Recruitment Transaction
-- Architecture. Assignment lifecycle state belongs to each assignment, not
-- to a candidate-wide boolean.

ALTER TABLE candidates
    DROP COLUMN IF EXISTS active_recruitment_engagements;

