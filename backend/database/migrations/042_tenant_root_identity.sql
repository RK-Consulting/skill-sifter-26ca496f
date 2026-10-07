-- 042_tenant_root_identity.sql
-- Greenfield tenant model: platform_tenants.tenant_id is the single
-- customer identity. companies is not a customer entity and is removed.

DO $$
DECLARE
    r RECORD;
    constraint_name TEXT;
BEGIN
    FOR r IN
        SELECT c.conrelid::regclass AS table_name, c.conname
        FROM pg_constraint c
        WHERE c.contype = 'f'
          AND c.confrelid = 'companies'::regclass
          AND pg_get_constraintdef(c.oid) ILIKE '%tenant_id%'
    LOOP
        EXECUTE format('ALTER TABLE %s DROP CONSTRAINT %I', r.table_name, r.conname);
    END LOOP;

    FOR r IN
        SELECT DISTINCT a.attrelid::regclass AS table_name
        FROM pg_attribute a
        JOIN pg_class cl ON cl.oid = a.attrelid
        JOIN pg_namespace n ON n.oid = cl.relnamespace
        WHERE a.attname = 'tenant_id'
          AND a.attnum > 0
          AND NOT a.attisdropped
          AND n.nspname = 'public'
          AND cl.relkind = 'r'
          AND cl.relname <> 'platform_tenants'
          AND NOT EXISTS (
              SELECT 1
              FROM pg_constraint c
              WHERE c.conrelid = a.attrelid
                AND c.contype = 'f'
                AND a.attnum = ANY(c.conkey)
                AND c.confrelid = 'platform_tenants'::regclass
          )
    LOOP
        constraint_name := left(
            'fk_' || replace(r.table_name::text, '"', '') || '_tenant',
            63
        );

        EXECUTE format(
            'ALTER TABLE %s ADD CONSTRAINT %I FOREIGN KEY (tenant_id) REFERENCES platform_tenants(tenant_id) ON DELETE CASCADE',
            r.table_name,
            constraint_name
        );
    END LOOP;
END $$;

DROP TABLE IF EXISTS companies CASCADE;

COMMENT ON TABLE platform_tenants IS
    'SaaS customer root. tenant_id is the single customer identity; other id columns identify records only.';

COMMENT ON COLUMN platform_tenants.tenant_id IS
    'Canonical immutable SaaS customer identity.';
