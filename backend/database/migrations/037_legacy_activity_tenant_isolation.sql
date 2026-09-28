-- 037_legacy_activity_tenant_isolation.sql
-- Move legacy reporting/audit records onto the authoritative tenant boundary.
-- company_name remains compatibility/display data only.

ALTER TABLE activity_logs DROP CONSTRAINT IF EXISTS activity_logs_tenant_id_fkey;
ALTER TABLE resume_search_logs DROP CONSTRAINT IF EXISTS resume_search_logs_tenant_id_fkey;

ALTER TABLE activity_logs
    ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);

ALTER TABLE resume_search_logs
    ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id);

UPDATE activity_logs a
SET tenant_id = c.id
FROM companies c
WHERE a.tenant_id IS NULL
  AND a.company_name = c.name;

UPDATE resume_search_logs r
SET tenant_id = c.id
FROM companies c
WHERE r.tenant_id IS NULL
  AND r.company_name = c.name;

CREATE INDEX IF NOT EXISTS idx_activity_tenant_time
    ON activity_logs(tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_resume_search_tenant_time
    ON resume_search_logs(tenant_id, created_at DESC);

CREATE OR REPLACE FUNCTION skillsifter_activity_trigger() RETURNS trigger AS $$
DECLARE
    company TEXT;
    tenant TEXT;
    entity TEXT;
    entity_id_text TEXT;
    payload JSONB;
BEGIN
    company := COALESCE(NEW.company_name, OLD.company_name);
    tenant := COALESCE(NEW.tenant_id, OLD.tenant_id);
    entity := TG_TABLE_NAME;
    entity_id_text := COALESCE(NEW.id, OLD.id)::text;

    IF TG_OP = 'INSERT' THEN
        payload := to_jsonb(NEW);
    ELSIF TG_OP = 'UPDATE' THEN
        payload := jsonb_build_object('before', to_jsonb(OLD), 'after', to_jsonb(NEW));
    ELSE
        payload := to_jsonb(OLD);
    END IF;

    INSERT INTO activity_logs(
        tenant_id,
        company_name,
        action,
        entity_type,
        entity_id,
        description,
        metadata
    )
    VALUES(
        tenant,
        company,
        upper(TG_TABLE_NAME) || '_' || upper(TG_OP),
        entity,
        entity_id_text,
        upper(TG_OP) || ' on ' || entity,
        payload
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
