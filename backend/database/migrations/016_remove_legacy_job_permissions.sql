-- 016_remove_legacy_job_permissions.sql
-- Remove legacy Jobs permissions from seeded roles without rewriting the
-- historical baseline migration. Requirements are the recruitment-demand
-- resource going forward.

UPDATE roles
SET permissions = (
    SELECT COALESCE(jsonb_agg(to_jsonb(value)), '[]'::jsonb)
    FROM jsonb_array_elements_text(permissions) AS p(value)
    WHERE value NOT IN ('manage_jobs', 'view_jobs')
)
WHERE permissions ?| ARRAY['manage_jobs', 'view_jobs'];