import { test, expect } from '@playwright/test';

const email = process.env.E2E_ADMIN_EMAIL;
const password = process.env.E2E_ADMIN_PASSWORD;
const runMutations = process.env.E2E_RUN_MUTATIONS === 'true';

function requireCredentials() {
  if (!email || !password) {
    throw new Error(
      'Production smoke requires E2E_ADMIN_EMAIL and E2E_ADMIN_PASSWORD. ' +
      'The configured account is the permanent SkillSifter production-smoke administrator.'
    );
  }
}

async function login(page: import('@playwright/test').Page) {
  requireCredentials();
  await page.goto('/login');
  await expect(page.getByRole('heading', { name: 'Login' })).toBeVisible();
  await page.getByLabel('Email').fill(email!);
  await page.getByLabel('Password').fill(password!);
  await page.getByRole('button', { name: 'Login' }).click();
  await expect(page).not.toHaveURL(/\/login$/);
}

async function deleteResource(
  page: import('@playwright/test').Page,
  path: string,
  token: string,
  apiOrigin: string,
  label: string,
) {
  const response = await page.request.delete(new URL(path, apiOrigin).toString(), {
    headers: { Authorization: `Bearer ${token}` },
  });

  if (!response.ok()) {
    throw new Error(`Production smoke cleanup failed for ${label}: HTTP ${response.status()}`);
  }
}

function createdID(payload: { data?: { id?: number } }) {
  const id = payload.data?.id;
  if (!id) {
    throw new Error('Production smoke mutation response did not contain a created record ID.');
  }
  return id;
}

test.describe('SkillSifter Phase 9 production smoke', () => {
  test('public login and registration pages load', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: 'Login' })).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password')).toBeVisible();

    await page.goto('/register');
    await expect(page.getByRole('heading', { name: 'Start Using SkillSifter' })).toBeVisible();
    await expect(page.getByLabel('Administrator name')).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Company / Tenant Name')).toBeVisible();
  });

  test('authenticated application smoke across Phase 9 modules', async ({ page }) => {
    await login(page);

    await expect(page).toHaveURL(/\/$/);
    await expect(page.getByRole('heading').first()).toBeVisible();

    const routes = [
      ['/clients', 'Clients'],
      ['/requirements', 'Requirements'],
      ['/candidates', 'Candidates'],
      ['/recruitment/lifecycle', 'Recruitment Lifecycle'],
      ['/billing', 'Billing'],
      ['/account', 'Account & Subscription'],
      ['/admin/users', 'User Management'],
    ] as const;

    for (const [route, heading] of routes) {
      await page.goto(route);
      await expect(page.getByRole('heading', { name: heading, exact: true })).toBeVisible();
    }
  });

  test('subscription/account surfaces load without browser errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('pageerror', error => errors.push(error.message));

    await login(page);
    await page.goto('/account');
    await expect(page.getByRole('heading', { name: 'Account & Subscription' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Current Subscription' })).toBeVisible();
    await expect(page.getByText('Plan:', { exact: false })).toBeVisible();

    await page.goto('/billing');
    await expect(page.getByRole('heading', { name: 'Billing', exact: true })).toBeVisible();

    expect(errors, errors.join('\n')).toEqual([]);
  });

  test('mutating production smoke uses the permanent account and clears only its test data', async ({ page }) => {
    test.skip(!runMutations, 'Set E2E_RUN_MUTATIONS=true to run the opt-in production smoke mutation/cleanup flow.');

    await login(page);

    const suffix = Date.now().toString();
    const clientName = `Smoke Client ${suffix}`;
    const candidateName = `Smoke Candidate ${suffix}`;
    const candidateEmail = `smoke-${suffix}@example.invalid`;
    const requirementTitle = `Smoke Software Engineer ${suffix}`;

    let clientID: number | undefined;
    let requirementID: number | undefined;
    let candidateID: number | undefined;
    let apiOrigin: string | undefined;

    // Client
    await page.goto('/clients/add');
    await expect(page.getByRole('heading', { name: 'Add Client' })).toBeVisible();
    await page.getByLabel('Client Name').fill(clientName);
    await page.getByLabel('Contact Person').fill('Production Smoke Contact');
    await page.getByLabel('Status').selectOption('active');
    await page.getByLabel('Contact Email').fill(`smoke-contact-${suffix}@example.invalid`);

    const clientResponsePromise = page.waitForResponse(
      response => response.url().includes('/api/v1/clients') && response.request().method() === 'POST',
    );
    await page.getByRole('button', { name: 'Save Client' }).click();
    const clientResponse = await clientResponsePromise;
    expect(clientResponse.ok()).toBeTruthy();
    apiOrigin = new URL(clientResponse.url()).origin;
    clientID = createdID(await clientResponse.json());

    await expect(page).toHaveURL(/\/clients$/);
    await expect(page.getByText(clientName, { exact: true })).toBeVisible();

    // Requirement
    await page.goto('/requirements/add');
    await expect(page.getByLabel('Job ID *')).toBeVisible();
    await page.getByLabel('Job ID *').fill(`SMOKE-${suffix}`);
    const clientSelect = page.getByLabel('Client *');
    await expect(clientSelect).toBeVisible();
    await clientSelect.selectOption({ label: clientName });
    await page.getByLabel('Job Title *').fill(requirementTitle);
    await page.getByLabel('Department').fill('Engineering');
    await page.getByLabel('Experience Required').fill('5 years');
    await page.getByLabel('Budget').fill('1800000');
    await page.getByLabel('Notice Period').fill('30 days');
    await page.getByLabel('Job Location').fill('Bengaluru');
    await page.getByLabel('No. of Open Positions *').fill('1');
    await page.getByLabel('Mandatory Requirements').fill('Go, PostgreSQL, REST APIs');
    await page.getByLabel('Job Description').fill('Production smoke requirement.');

    const requirementResponsePromise = page.waitForResponse(
      response => response.url().includes('/api/v1/requirements') && response.request().method() === 'POST',
    );
    await page.getByRole('button', { name: /Save|Create|Add Requirement/i }).click();
    const requirementResponse = await requirementResponsePromise;
    expect(requirementResponse.ok()).toBeTruthy();
    requirementID = createdID(await requirementResponse.json());

    await expect(page).toHaveURL(/\/requirements$/);

    // Candidate
    await page.goto('/candidates/add');
    await expect(page.getByRole('heading', { name: 'Add New Candidate' })).toBeVisible();
    await page.getByLabel('Full Name').fill(candidateName);
    await page.getByLabel('Email').fill(candidateEmail);
    await page.getByLabel('Phone Number').fill('9999999999');
    await page.getByLabel('Role/Position').fill('Software Engineer');
    await page.getByLabel('Location').fill('Bengaluru');
    await page.getByLabel('Experience').fill('5');
    await page.getByLabel('Current CTC').fill('1200000');
    await page.getByLabel('Expected CTC').fill('1600000');
    await page.getByLabel('Notice Period').fill('30');
    await page.getByLabel('Client Name').fill(clientName);
    await page.getByLabel('Skills').fill('Go, PostgreSQL, TypeScript');

    const candidateResponsePromise = page.waitForResponse(
      response => response.url().includes('/api/candidates') && response.request().method() === 'POST',
    );
    await page.getByRole('button', { name: /Add Candidate|Save Candidate|Create Candidate/i }).click();
    const candidateResponse = await candidateResponsePromise;
    expect(candidateResponse.ok()).toBeTruthy();
    candidateID = createdID(await candidateResponse.json());

    await expect(page).not.toHaveURL(/\/candidates\/add$/);

    // Cleanup runs only after the complete mutation flow has succeeded.
    // If any mutation fails above, the test data is intentionally preserved
    // for debugging. Cleanup never touches the tenant, administrator,
    // subscription, or tenant database.
    const token = await page.evaluate(() => localStorage.getItem('token'));
    if (!token) {
      throw new Error('Production smoke cleanup could not read the authenticated token.');
    }

    if (!apiOrigin) {
      throw new Error('Production smoke cleanup could not determine the API origin.');
    }

    await deleteResource(page, `/api/candidates/${candidateID}`, token, apiOrigin, 'candidate');
    await deleteResource(page, `/api/v1/requirements/${requirementID}`, token, apiOrigin, 'requirement');
    await deleteResource(page, `/api/v1/clients/${clientID}`, token, apiOrigin, 'client');
  });
});
