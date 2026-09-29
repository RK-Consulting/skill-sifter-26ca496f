# SkillSifter Playwright E2E / Production Smoke

This is the browser-level Phase 9 verification suite for SkillSifter.

## Target

Default target:

`https://skillsifter.in`

Override with:

`PLAYWRIGHT_TEST_BASE_URL=https://<staging-host>`

## Setup

```bash
cd e2e
npm install
npx playwright install chromium
```

## Production smoke

The repository seeds one dedicated **permanent production-smoke account** through the database schema definitions:

- tenant: `e2e_smoke_tenant`
- company: `SkillSifter E2E Smoke`
- administrator: `e2e-admin@skillsifter.in`
- subscription: `e2e-smoke / ACTIVE / no expiry`
- tenant routing: `READY`

The account uses the normal login, RBAC, tenant-routing, and subscription checks. It does not bypass authentication or payment/subscription gates.

The smoke credentials are supplied through GitHub Actions secrets:

- `SKILLSIFTER_E2E_ADMIN_EMAIL`
- `SKILLSIFTER_E2E_ADMIN_PASSWORD`

For the seeded fixture, the email must be `e2e-admin@skillsifter.in`. The password must match the password represented by the bcrypt hash in `038_e2e_smoke_account.sql`.

The normal smoke suite is non-destructive. It checks public authentication pages, authenticated module access, Account & Subscription, Billing, and browser-level errors.

## Mutating production smoke

The client/requirement/candidate creation flow is opt-in:

```bash
E2E_RUN_MUTATIONS=true \
E2E_ADMIN_EMAIL="..." \
E2E_ADMIN_PASSWORD="..." \
npm run smoke
```

The mutation flow uses the permanent production-smoke tenant and creates uniquely named test records. **Only after the complete mutation flow succeeds**, the test deletes the candidate, requirement, and client created by that run, in dependency order.

The cleanup never deletes:

- the production-smoke tenant
- its tenant database
- the permanent administrator
- the platform user account
- the active subscription
- the subscription/plan configuration

If the mutation flow fails before cleanup begins, the created records are intentionally retained so the failure can be diagnosed.

## Reports

```bash
npm run report
```

Playwright retains screenshots, video and traces for failures. The HTML report is written to `playwright-report/`.

The production smoke workflow is manual by design. Production credentials are supplied through GitHub Actions secrets and are never committed to the repository.
