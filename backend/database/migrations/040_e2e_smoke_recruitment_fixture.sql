-- 040_e2e_smoke_recruitment_fixture.sql
-- Permanent tenant-scoped production smoke fixture.
--
-- This provides one complete recruitment transaction for the permanent
-- e2e_smoke_tenant so every recruitment tab has realistic data to render.
-- It is deliberately isolated to the E2E smoke tenant and is not customer data.
-- Jobs are not seeded: Requirements are the authoritative demand object.

DO $$
DECLARE
    v_tenant_id VARCHAR(255) := 'e2e_smoke_tenant';
    v_company_name VARCHAR(255) := 'SkillSifter E2E Smoke';
    v_user_id INTEGER;
    v_client_id INTEGER;
    v_candidate_id INTEGER;
    v_requirement_id INTEGER;
    v_submission_id INTEGER;
    v_interview_id INTEGER;
    v_selection_id INTEGER;
    v_offer_id INTEGER;
    v_joining_id INTEGER;
BEGIN
    SELECT id INTO v_user_id
    FROM users
    WHERE email = 'e2e-admin@skillsifter.in'
      AND tenant_id = v_tenant_id
    LIMIT 1;

    IF v_user_id IS NULL THEN
        RAISE EXCEPTION 'E2E smoke admin not found';
    END IF;

    INSERT INTO clients (
        name, status, contact_email, contact_phone, partner_name,
        contact_person, tenant_id
    )
    VALUES (
        'E2E Smoke Client',
        'active',
        'e2e-client@skillsifter.in',
        '9999999999',
        NULL,
        'E2E Smoke Contact',
        v_tenant_id
    )
    RETURNING id INTO v_client_id;

    INSERT INTO candidates (
        name, email, phone, position, location, experience,
        currentctc, expectedctc, noticeperiod, jobdescription,
        status, tenant_id, company_name
    )
    VALUES (
        'E2E Smoke Candidate',
        'e2e-candidate@skillsifter.in',
        '9999999998',
        'Software Engineer',
        'Bengaluru',
        '5 years',
        '1200000',
        '1600000',
        '30 days',
        'Permanent production smoke candidate.',
        'active',
        v_tenant_id,
        v_company_name
    )
    RETURNING id INTO v_candidate_id;

    INSERT INTO candidate_expertise (
        tenant_id, candidate_id, skill, category, proficiency_level
    )
    VALUES
        (v_tenant_id, v_candidate_id, 'Go', 'Backend', 'Advanced'),
        (v_tenant_id, v_candidate_id, 'PostgreSQL', 'Database', 'Advanced');

    INSERT INTO requirements (
        client_id, job_id, job_type, title, department, experience_required,
        budget, language_requirement, certifications_required, notice_period,
        work_arrangement, required_skills, mandatory_requirements,
        description, status, location, headcount, opened_date, tenant_id
    )
    VALUES (
        v_client_id,
        'E2E-SMOKE-001',
        'fulltime',
        'E2E Smoke Software Engineer',
        'Engineering',
        '5 years',
        '1800000',
        'English',
        '',
        '30 days',
        'hybrid',
        'Go, PostgreSQL, REST APIs',
        'Go, PostgreSQL, REST APIs',
        'Permanent production smoke requirement.',
        'open',
        'Bengaluru',
        1,
        CURRENT_DATE,
        v_tenant_id
    )
    RETURNING id INTO v_requirement_id;

    INSERT INTO recruitment_screenings (
        tenant_id, candidate_id, requirement_id, recruiter_user_id, status,
        current_ctc, expected_ctc, notice_period, current_location,
        willing_to_relocate, preferred_location, candidate_interest,
        relevant_experience, recruiter_assessment, notes
    )
    VALUES (
        v_tenant_id, v_candidate_id, v_requirement_id, v_user_id, 'completed',
        '1200000', '1600000', '30 days', 'Bengaluru',
        TRUE, 'Bengaluru', 'interested',
        '5 years', 'E2E smoke screening completed.',
        'Permanent smoke fixture.'
    );

    INSERT INTO recruitment_submissions (
        tenant_id, candidate_id, requirement_id, submitted_by_user_id,
        recipient_type, recipient_client_id, recipient_name,
        recipient_email, submission_context, recruiter_notes,
        candidate_snapshot, requirement_snapshot
    )
    VALUES (
        v_tenant_id, v_candidate_id, v_requirement_id, v_user_id,
        'client', v_client_id, 'E2E Smoke Client',
        'e2e-client@skillsifter.in',
        'E2E smoke submission.',
        'E2E smoke submission fixture.',
        jsonb_build_object('name', 'E2E Smoke Candidate', 'position', 'Software Engineer'),
        jsonb_build_object('jobId', 'E2E-SMOKE-001', 'title', 'E2E Smoke Software Engineer')
    )
    RETURNING id INTO v_submission_id;

    INSERT INTO recruitment_submission_feedback (
        tenant_id, submission_id, feedback_by_user_id,
        outcome, comments, next_action
    )
    VALUES (
        v_tenant_id, v_submission_id, v_user_id,
        'shortlist', 'E2E smoke client feedback.',
        'Proceed to interview.'
    );

    INSERT INTO interviews (
        candidate_id, candidate_name, requirement_id, position, round,
        interview_date, status, outcome, feedback, candidate_feedback,
        next_action, tenant_id, company_name
    )
    VALUES (
        v_candidate_id, 'E2E Smoke Candidate', v_requirement_id,
        'Software Engineer', 1,
        CURRENT_TIMESTAMP, 'completed', 'completed',
        'E2E smoke interview completed.',
        'E2E smoke candidate feedback.',
        'Proceed to selection.',
        v_tenant_id, v_company_name
    )
    RETURNING id INTO v_interview_id;

    INSERT INTO recruitment_selections (
        tenant_id, candidate_id, requirement_id, decision,
        decision_notes, next_action
    )
    VALUES (
        v_tenant_id, v_candidate_id, v_requirement_id, 'selected',
        'E2E smoke candidate selected.',
        'Make offer.'
    )
    RETURNING id INTO v_selection_id;

    INSERT INTO recruitment_offers (
        tenant_id, candidate_id, requirement_id, selection_id, accepted
    )
    VALUES (
        v_tenant_id, v_candidate_id, v_requirement_id, v_selection_id, TRUE
    )
    RETURNING id INTO v_offer_id;

    INSERT INTO recruitment_joinings (
        tenant_id, candidate_id, requirement_id, offer_id,
        joining_date, joined
    )
    VALUES (
        v_tenant_id, v_candidate_id, v_requirement_id, v_offer_id,
        CURRENT_DATE, TRUE
    )
    RETURNING id INTO v_joining_id;

    INSERT INTO recruitment_billings (
        tenant_id, candidate_id, requirement_id, client_id, joining_id,
        billing_date, amount, currency, invoice_reference
    )
    VALUES (
        v_tenant_id, v_candidate_id, v_requirement_id, v_client_id, v_joining_id,
        CURRENT_DATE, 1800000, 'INR', 'E2E-SMOKE-001'
    );
END $$;
