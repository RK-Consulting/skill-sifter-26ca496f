-- 041_e2e_smoke_operational_fixture.sql
-- Complete the permanent E2E smoke tenant with the remaining operational
-- surfaces. This does not recreate the removed Jobs domain.
--
-- Daily Tasks is a legacy operational-assignment table and remains separate
-- from Requirements. Business Development is seeded so Reports has coverage.

DO $$
DECLARE
    v_tenant_id VARCHAR(255) := 'e2e_smoke_tenant';
    v_company_name VARCHAR(255) := 'SkillSifter E2E Smoke';
    v_user_id INTEGER;
BEGIN
    SELECT id INTO v_user_id
    FROM users
    WHERE email = 'e2e-admin@skillsifter.in'
      AND tenant_id = v_tenant_id
    LIMIT 1;

    IF v_user_id IS NULL THEN
        RAISE EXCEPTION 'E2E smoke admin not found';
    END IF;

    INSERT INTO daily_jobs (
        jd_no, instructions, assigned_user, assigned_date,
        last_modified, tenant_id, company_name
    )
    VALUES (
        9001,
        'E2E smoke: review E2E Smoke Software Engineer requirement.',
        v_user_id,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        v_tenant_id,
        v_company_name
    );

    INSERT INTO business_dev (
        client_name, partner_name, contact_person, contact_number,
        contact_email, created_at, last_modified, tenant_id, company_name
    )
    VALUES (
        'E2E Smoke Prospect',
        'E2E Smoke Partner',
        'E2E Smoke Contact',
        '9999999997',
        'e2e-business-dev@skillsifter.in',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        v_tenant_id,
        v_company_name
    );
END $$;
