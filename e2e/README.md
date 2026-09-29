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

Use a disposable administrator account:

```bash
E2E_ADMIN_EMAIL="..." E2E_ADMIN_PASSWORD="..." npm run smoke
```

The default smoke suite is deliberately non-destructive. It checks public authentication pages, authenticated module access, Account & Subscription, Billing, and browser-level errors.

## Mutating UAT

The client/requirement/candidate creation flow is opt-in:

```bash
E2E_RUN_MUTATIONS=true \
E2E_ADMIN_EMAIL="..." \
E2E_ADMIN_PASSWORD="..." \
npm run smoke
```

Only use this against a disposable UAT tenant. It creates uniquely named dummy records.

## Reports

```bash
npm run report
```

Playwright retains screenshots, video and traces for failures. The HTML report is written to `playwright-report/`.

The production smoke workflow is manual by design. Production credentials are supplied through GitHub Actions secrets and are never committed to the repository.
