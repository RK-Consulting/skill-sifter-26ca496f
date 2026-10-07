-- 042_tenant_root_identity.sql
-- Make platform_tenants.tenant_id the single SaaS customer identity.
--
-- Record IDs such as users.id and platform_subscriptions.id remain record
-- identifiers. They do not identify the customer.
--
-- companies is retained as a compatibility projection for the current
-- application model. Its id must equal tenant_id and it is cascade-owned by
-- platform_tenants. New tenant-owned tables reference platform_tenants directly.

-- The companies row is now a compatibility child of the tenant root.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'companies_tenant_fk'
    ) THEN
        ALTER TABLE companies
            ADD CONSTRAINT companies_tenant_fk
            FOREIGN KEY (id) REFERENCES platform_tenants(tenant_id)
            ON DELETE CASCADE;
    END IF;
END $$;

-- Move every existing tenant_id foreign key from companies to the tenant root
-- and make tenant deletion cascade through tenant-owned control/application
-- data.
DO $$
DECLARE
    r RECORD;
    constraint_name TEXT;
BEGIN
    FOR r IN
        SELECT
            c.conrelid::regclass AS table_name,
            c.conname
        FROM pg_constraint c
        WHERE c.contype = 'f'
          AND c.confrelid = 'companies'::regclass
          AND pg_get_constraintdef(c.oid) ILIKE '%tenant_id%'
    LOOP
        EXECUTE format(
            'ALTER TABLE %s DROP CONSTRAINT %I',
            r.table_name,
            r.conname
        );
    END LOOP;

    FOR r IN
        SELECT DISTINCT
            a.attrelid::regclass AS table_name,
            a.attname AS column_name
        FROM pg_attribute a
        JOIN pg_class cl ON cl.oid = a.attrelid
        JOIN pg_namespace n ON n.oid = cl.relnamespace
        WHERE a.attname = 'tenant_id'
          AND a.attnum > 0
          AND NOT a.attisdropped
          AND n.nspname = 'public'
          AND cl.relkind = 'r'
          AND cl.relname NOT IN ('platform_tenants', 'companies')
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

COMMENT ON TABLE platform_tenants IS
    'SaaS customer root. tenant_id is the single customer identity; other id columns identify records only.';

COMMENT ON COLUMN platform_tenants.tenant_id IS
    'Canonical and immutable SaaS customer identity used by control-plane and tenant-owned data.';
