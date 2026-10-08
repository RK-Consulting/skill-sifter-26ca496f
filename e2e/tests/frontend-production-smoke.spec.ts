import { test, expect } from '@playwright/test';

const email = process.env.E2E_ADMIN_EMAIL;
const password = process.env.E2E_ADMIN_PASSWORD;

function requireCredentials() {
  if (!email || !password) {
    throw new Error('Frontend production smoke requires E2E_ADMIN_EMAIL and E2E_ADMIN_PASSWORD.');
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

test.describe('SkillSifter frontend production smoke', () => {
  test('public frontend loads and exposes the authentication entry points', async ({ page }) => {
    const consoleErrors: string[] = [];
    const pageErrors: string[] = [];
    page.on('console', message => {
      if (message.type() === 'error') consoleErrors.push(message.text());
    });
    page.on('pageerror', error => pageErrors.push(error.message));

    await page.goto('/');
    await expect(page).toHaveTitle(/SkillSifter/i);
    await expect(page.getByRole('heading').first()).toBeVisible();

    await page.goto('/login');
    await expect(page.getByRole('heading', { name: 'Login' })).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password')).toBeVisible();

    await page.goto('/register');
    await expect(page.getByRole('heading', { name: 'Start Using SkillSifter' })).toBeVisible();
    await expect(page.getByLabel('Administrator name')).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Company / Tenant Name')).toBeVisible();

    expect(pageErrors, pageErrors.join('\n')).toEqual([]);
    expect(consoleErrors, consoleErrors.join('\n')).toEqual([]);
  });

  test('authenticated frontend routes render and remain free of browser errors', async ({ page }) => {
    const consoleErrors: string[] = [];
    const pageErrors: string[] = [];
    page.on('console', message => {
      if (message.type() === 'error') consoleErrors.push(message.text());
    });
    page.on('pageerror', error => pageErrors.push(error.message));

    await login(page);

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

    expect(pageErrors, pageErrors.join('\n')).toEqual([]);
    expect(consoleErrors, consoleErrors.join('\n')).toEqual([]);
  });

  test('account and subscription frontend integrates with the live API', async ({ page }) => {
    await login(page);
    await page.goto('/account');

    await expect(page.getByRole('heading', { name: 'Account & Subscription' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Current Subscription' })).toBeVisible();
    await expect(page.getByText('Plan:', { exact: false })).toBeVisible();

    await page.goto('/billing');
    await expect(page.getByRole('heading', { name: 'Billing', exact: true })).toBeVisible();
  });
});
