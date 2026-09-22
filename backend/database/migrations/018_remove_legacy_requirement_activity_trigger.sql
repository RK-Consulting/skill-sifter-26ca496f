-- 018_remove_legacy_requirement_activity_trigger.sql
-- Requirements use the tenant_id/client_id domain and must not invoke the
-- legacy activity trigger from migration 002, which expects company_name.
--
-- The legacy trigger function remains in place for older tables that still
-- use company_name. Only the Requirements trigger is removed.

DROP TRIGGER IF EXISTS trg_requirements_activity ON requirements;
