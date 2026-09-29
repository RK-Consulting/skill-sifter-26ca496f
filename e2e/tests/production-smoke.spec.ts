import { test, expect } from '@playwright/test';

const email = process.env.E2E_ADMIN_EMAIL;
const password = process.env.E2E_ADMIN_PASSWORD;
const runMutations = process.env.E2E_RUN_MUTATIONS === 'true';

function requireCredentials() {
  if (!email || !password) {
    throw new Error(
      'Production smoke requires E2E_ADMIN_EMAIL and E2E_ADMIN_PASSWORD. ' +
      'Provide a disposable SkillSifter administrator account through environment variables.'
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

test.describe('SkillSifter Phase 9 production smoke', () => {
  test('public login and registration pages load', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: 'Login' })).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password')).toBeVisible();

    await page.goto('/register');
    await expect(page.getByRole('heading', { name: 'Create Tenant Account' })).toBeVisible();
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
    await expect(page.getByRole('heading', { name: 'Available Plans' })).toBeVisible();

    await page.goto('/billing');
    await expect(page.getByRole('heading', { name: 'Billing' })).toBeVisible();

    expect(errors, errors.join('\n')).toEqual([]);
  });

  test('mutating UAT can be explicitly enabled with disposable credentials', async ({ page }) => {
    test.skip(!runMutations, 'Set E2E_RUN_MUTATIONS=true only for a disposable UAT tenant.');

    await login(page);

    const suffix = Date.now().toString();
    const clientName = `E2E Client ${suffix}`;
    const candidateName = `E2E Candidate ${suffix}`;
    const candidateEmail = `e2e-${suffix}@example.invalid`;
    const requirementTitle = `E2E Software Engineer ${suffix}`;

    // Client
    await page.goto('/clients/add');
    await expect(page.getByRole('heading', { name: 'Add Client' })).toBeVisible();
    await page.getByLabel('Client Name').fill(clientName);
    await page.getByLabel('Contact Person').fill('E2E Contact');
    await page.getByLabel('Contact Email').fill(`contact-${suffix}@example.invalid`);
    await page.getByRole('button', { name: 'Save Client' }).click();
    await expect(page).toHaveURL(/\/clients$/);
    await expect(page.getByText(clientName, { exact: true })).toBeVisible();

    // Requirement
    await page.goto('/requirements/add');
    await expect(page.getByLabel('Job ID *')).toBeVisible();
    await page.getByLabel('Job ID *').fill(`E2E-${suffix}`);
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
    await page.getByLabel('Job Description').fill('E2E production smoke requirement.');
    await page.getByRole('button', { name: /Save|Create|Add Requirement/i }).click();
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
    await page.getByRole('button', { name: /Add Candidate|Save Candidate|Create Candidate/i }).click();
    await expect(page).not.toHaveURL(/\/candidates\/add$/);
  });
});
