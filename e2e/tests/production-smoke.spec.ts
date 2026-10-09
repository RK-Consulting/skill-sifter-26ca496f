import { test, expect } from '@playwright/test';

const email = process.env.E2E_ADMIN_EMAIL;
const password = process.env.E2E_ADMIN_PASSWORD;
const apiOrigin = process.env.E2E_API_BASE_URL || 'https://api.skillsifter.in';
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


async function resetSmokeTenant(request: import('@playwright/test').APIRequestContext) {
  requireCredentials();

  const response = await request.post(`${apiOrigin}/api/e2e/reset`, {
    data: { email, password },
  });

  if (!response.ok()) {
    throw new Error(
      `Production smoke cleanup failed: HTTP ${response.status()} ${await response.text()}`,
    );
  }

  const payload = await response.json();
  expect(payload.success).toBe(true);
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

  test('mutating production smoke exercises the current recruitment lifecycle and cleans the dedicated tenant', async ({ page, request }) => {
    test.skip(!runMutations, 'Set E2E_RUN_MUTATIONS=true to run the production mutation smoke.');

    requireCredentials();
    await login(page);

    const suffix = Date.now().toString();
    const clientName = `Smoke Client ${suffix}`;
    const candidateName = `Smoke Candidate ${suffix}`;
    const candidateEmail = `smoke-${suffix}@example.invalid`;
    const requirementTitle = `Smoke Software Engineer ${suffix}`;
    const jobID = `SMOKE-${suffix}`;

    let cleanupError: unknown;

    try {
      // Client
      await page.goto('/clients/add');
      await expect(page.getByRole('heading', { name: 'Add Client' })).toBeVisible();
      await page.getByLabel('Client Name').fill(clientName);
      await page.getByLabel('Contact Person').fill('Production Smoke Contact');
      await page.getByLabel('Contact Email').fill(`smoke-contact-${suffix}@example.invalid`);
      await page.getByLabel('Status').selectOption('active');

      const clientResponsePromise = page.waitForResponse(
        response => response.url().includes('/api/v1/clients') && response.request().method() === 'POST',
      );
      await page.getByRole('button', { name: 'Save Client' }).click();
      const clientResponse = await clientResponsePromise;
      expect(clientResponse.ok()).toBeTruthy();
      const clientID = createdID(await clientResponse.json());
      expect(clientID).toBeGreaterThan(0);
      await expect(page).toHaveURL(/\/clients$/);
      await expect(page.getByText(clientName, { exact: true })).toBeVisible();

      // Requirement
      await page.goto('/requirements/add');
      await expect(page.getByLabel('Job ID *')).toBeVisible();
      await page.getByLabel('Job ID *').fill(jobID);
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
      const requirementID = createdID(await requirementResponse.json());
      expect(requirementID).toBeGreaterThan(0);
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
      await page.getByRole('main').getByRole('button', { name: 'Add Candidate', exact: true }).click();
      const candidateResponse = await candidateResponsePromise;
      expect(candidateResponse.ok()).toBeTruthy();
      const candidateID = createdID(await candidateResponse.json());
      expect(candidateID).toBeGreaterThan(0);
      await expect(page).not.toHaveURL(/\/candidates\/add$/);

      // Requirement × Candidate recruitment lifecycle
      await page.goto(`/recruitment/lifecycle?candidateId=${candidateID}&requirementId=${requirementID}`);
      await expect(page.getByRole('heading', { name: 'Recruitment Lifecycle' })).toBeVisible();
      await expect(page.getByText('1. Screening', { exact: true })).toBeVisible();

      await page.getByPlaceholder('Recruiter assessment (optional)').fill('Production smoke screening passed');
      await page.getByRole('button', { name: 'Complete Screening', exact: true }).click();
      await expect(page.getByText('Completed — Production smoke screening passed')).toBeVisible();

      await page.getByRole('button', { name: 'Submit to Client', exact: true }).click();
      await expect(page.getByText(`Submitted to ${clientName}`)).toBeVisible();

      await page.getByPlaceholder('Client feedback (optional)').fill('Production smoke client feedback');
      await page.getByRole('button', { name: 'Record Feedback', exact: true }).click();
      await expect(page.getByText('shortlist', { exact: true })).toBeVisible();

      const interviewDate = new Date(Date.now() + 24 * 60 * 60 * 1000);
      const localDateTime = new Date(interviewDate.getTime() - interviewDate.getTimezoneOffset() * 60000)
        .toISOString()
        .slice(0, 16);
      await page.locator('#interview-date').fill(localDateTime);
      await page.getByRole('button', { name: 'Schedule Interview', exact: true }).click();
      await expect(page.getByRole('button', { name: 'Complete Interview', exact: true })).toBeVisible();
      await page.getByPlaceholder('Interview outcome').fill('Production smoke interview completed');
      await page.getByRole('button', { name: 'Complete Interview', exact: true }).click();
      await expect(page.getByText('Production smoke interview completed')).toBeVisible();

      await page.getByRole('button', { name: 'Select Candidate', exact: true }).click();
      await expect(page.getByText('selected', { exact: true })).toBeVisible();

      await page.getByRole('button', { name: 'Create Offer', exact: true }).click();
      await expect(page.getByRole('button', { name: 'Accept Offer', exact: true })).toBeVisible();
      await page.getByRole('button', { name: 'Accept Offer', exact: true }).click();
      await expect(page.getByText('Accepted', { exact: true })).toBeVisible();

      const joiningDate = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);
      await page.locator('#joining-date').fill(joiningDate);
      await page.getByRole('button', { name: 'Record Joining', exact: true }).click();
      await expect(page.getByText('Joined', { exact: true })).toBeVisible();

      await page.locator('#billing-amount').fill('50000');
      await page.locator('#billing-currency').fill('INR');
      await page.locator('#invoice-reference').fill(`SMOKE-${suffix}`);
      await page.getByRole('button', { name: 'Create Billing', exact: true }).click();
      await expect(page.getByText('INR 50000', { exact: true })).toBeVisible();
      await expect(page.getByText(`SMOKE-${suffix}`, { exact: true })).toBeVisible();
    } finally {
      try {
        await resetSmokeTenant(request);
      } catch (error) {
        cleanupError = error;
      }
    }

    if (cleanupError) {
      throw cleanupError;
    }
  });
});
