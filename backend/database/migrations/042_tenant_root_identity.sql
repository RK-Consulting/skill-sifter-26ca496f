-- 042_tenant_root_identity.sql
-- Make platform_tenants.tenant_id the single SaaS customer identity.
--
-- Record IDs such as users.id and platform_subscriptions.id remain record
-- identifiers. They do not identify the customer.
--
-- companies is retained as a compatibility projection for the current
-- application model. Its id must equal tenant_id and it is cascade-owned by
-- platform_tenants. New tenant-owned tables reference platform_tenants directly.

-- companies remains a compatibility table. Legacy fixtures and older
-- administrative paths may still insert it directly, so bridge those writes
-- into the canonical tenant root without making companies a second identity.
CREATE OR REPLACE FUNCTION ensure_platform_tenant_for_company()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO platform_tenants (
        tenant_id, company_name, account_status, provisioning_status
    )
    VALUES (
        NEW.id, NEW.name, 'ACTIVE', 'READY'
    )
    ON CONFLICT (tenant_id) DO UPDATE
        SET company_name = EXCLUDED.company_name,
            updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS companies_platform_tenant_bridge ON companies;

CREATE TRIGGER companies_platform_tenant_bridge
AFTER INSERT OR UPDATE OF name ON companies
FOR EACH ROW
EXECUTE FUNCTION ensure_platform_tenant_for_company();

-- Root deletion owns the compatibility company row as well. This is a
-- trigger rather than a foreign key to avoid a circular insert dependency.
CREATE OR REPLACE FUNCTION delete_compatibility_company_for_tenant()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM companies WHERE id = OLD.tenant_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS platform_tenant_delete_company ON platform_tenants;

CREATE TRIGGER platform_tenant_delete_company
AFTER DELETE ON platform_tenants
FOR EACH ROW
EXECUTE FUNCTION delete_compatibility_company_for_tenant();

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
