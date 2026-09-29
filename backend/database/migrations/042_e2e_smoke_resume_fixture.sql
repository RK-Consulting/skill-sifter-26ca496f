-- 042_e2e_smoke_resume_fixture.sql
-- Permanent Resume AI fixture for production smoke.
-- The stored text file is part of the application tree so the real
-- repository/detail/download paths can be exercised without Ollama.

DO $$
DECLARE
    v_tenant_id VARCHAR(255) := 'e2e_smoke_tenant';
    v_company_name VARCHAR(255) := 'SkillSifter E2E Smoke';
    v_candidate_id INTEGER;
    v_resume_id INTEGER;
BEGIN
    SELECT id INTO v_candidate_id
    FROM candidates
    WHERE tenant_id = v_tenant_id
      AND email = 'e2e-candidate@skillsifter.in'
    LIMIT 1;

    IF v_candidate_id IS NULL THEN
        RAISE EXCEPTION 'E2E smoke candidate not found';
    END IF;

    SELECT id INTO v_resume_id
    FROM resumes
    WHERE tenant_id = v_tenant_id
      AND file_name = 'e2e-smoke-candidate.txt'
    LIMIT 1;

    IF v_resume_id IS NULL THEN
        INSERT INTO resumes (
            tenant_id, company_name, candidate_id, file_name, file_path,
            file_hash, mime_type, extracted_text, parsing_status,
            parser_model, uploaded_by, uploaded_at, parsed_at
        )
        SELECT
            v_tenant_id,
            v_company_name,
            v_candidate_id,
            'e2e-smoke-candidate.txt',
            './storage/resumes/e2e_smoke_tenant/e2e-smoke-candidate.txt',
            'deaa46b6bfe4a00481ea4b6fcdb1dbb364615e77fd09046e5cb29b14e4f1cb80',
            'text/plain',
            'E2E Smoke Candidate
Email: e2e-candidate@skillsifter.in
Phone: 9999999998
Location: Bengaluru
Current Title: Software Engineer
Experience: 5 years

Professional Summary:
Software Engineer with experience building backend systems and APIs.

Technical Skills:
Go, PostgreSQL, TypeScript, REST APIs

Education:
Bachelor of Engineering

Certifications:
None',
            'completed',
            'e2e-smoke-fixture',
            u.id,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        FROM users u
        WHERE u.id = (
            SELECT id FROM users
            WHERE email = 'e2e-admin@skillsifter.in'
              AND tenant_id = v_tenant_id
            LIMIT 1
        )
        RETURNING id INTO v_resume_id;
    END IF;
END $$;
