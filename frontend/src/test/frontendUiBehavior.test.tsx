// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

class ResizeObserverMock {
  observe = vi.fn();
  unobserve = vi.fn();
  disconnect = vi.fn();
}

window.ResizeObserver = ResizeObserverMock as unknown as typeof ResizeObserver;
Element.prototype.scrollIntoView = vi.fn();

import { MemoryRouter, Route, Routes } from 'react-router-dom';

const mocks = vi.hoisted(() => {
  const service = () => {
    const methods = new Map<string | symbol, ReturnType<typeof vi.fn>>();
    return new Proxy(
      {},
      {
        get: (_target, property: string | symbol) => {
          if (!methods.has(property)) {
            methods.set(property, vi.fn().mockResolvedValue({ data: { success: true, data: [] } }));
          }
          return methods.get(property);
        },
      },
    );
  };

  return {
    authService: service(),
    candidateService: service(),
    candidateRecruitmentService: service(),
    clientService: service(),
    requirementService: service(),
    dailyJobService: service(),
    interviewService: service(),
    reportService: service(),
    resumeAIService: service(),
    businessDevService: service(),
    companyService: service(),
    roleService: service(),
    userService: service(),
    subscriptionService: service(),
    toast: {
      success: vi.fn(),
      error: vi.fn(),
    },
  };
});

vi.mock('@/services/api', () => mocks);
vi.mock('@/services/resumeAIService', () => ({ resumeAIService: mocks.resumeAIService }));
vi.mock('sonner', () => ({ toast: mocks.toast, Toaster: () => null }));

import Login from '@/pages/Login';
import Register from '@/pages/Register';
import Index from '@/pages/Index';
import Candidates from '@/pages/Candidates';
import AddCandidate from '@/pages/AddCandidate';
import CandidateProfile from '@/pages/CandidateProfile';
import DailyJobs from '@/pages/DailyJobs';
import AddDailyJob from '@/pages/AddDailyJob';
import Interviews from '@/pages/Interviews';
import InterviewDetails from '@/pages/InterviewDetails';
import ScheduleInterview from '@/pages/ScheduleInterview';
import Clients from '@/pages/Clients';
import AddClient from '@/pages/AddClient';
import Requirements from '@/pages/Requirements';
import AddRequirement from '@/pages/AddRequirement';
import Reports from '@/pages/Reports';
import ResumeAI from '@/pages/ResumeAI';
import RecruitmentLifecycle from '@/pages/RecruitmentLifecycle';
import Billing from '@/pages/Billing';
import AdminUsers from '@/pages/AdminUsers';
import Account from '@/pages/Account';
import NotFound from '@/pages/NotFound';
import Navbar from '@/components/layout/Navbar';

const renderPage = (ui: React.ReactElement, initialEntries = ['/']) => {
  const route = initialEntries[0];
  const path = route.startsWith('/candidates/') ? '/candidates/:id'
    : route.startsWith('/interviews/') ? '/interviews/:id'
    : route;
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={initialEntries}>
        <Routes>
          <Route path={path} element={ui} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
};

const setLoggedIn = () => {
  localStorage.setItem('token', 'test-token');
  localStorage.setItem(
    'user',
    JSON.stringify({ username: 'Test User', email: 'test@example.com', isLoggedIn: true }),
  );
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
});

describe('frontend UI smoke coverage', () => {
  it.each([
    ['Login', <Login />, 'Login'],
    ['Register', <Register />, 'Create Tenant Account'],
    ['Dashboard', <Index />, 'SkillSifter ATS'],
    ['Candidates', <Candidates />, 'Candidates'],
    ['Add Candidate', <AddCandidate />, 'Add New Candidate'],
    ['Candidate Profile', <CandidateProfile />, 'Resume AI source', '/candidates/1'],
    ['Daily Tasks', <DailyJobs />, 'Daily Job Assignments'],
    ['Add Daily Task', <AddDailyJob />, 'Add Daily Job Assignment'],
    ['Interviews', <Interviews />, 'Interviews'],
    ['Interview Details', <InterviewDetails />, 'Interview Details', '/interviews/1'],
    ['Schedule Interview', <ScheduleInterview />, 'Schedule Interview'],
    ['Clients', <Clients />, 'Clients'],
    ['Add Client', <AddClient />, 'Add Client'],
    ['Requirements', <Requirements />, 'Requirements'],
    ['Add Requirement', <AddRequirement />, 'Add Requirement'],
    ['Reports', <Reports />, 'Reports & Activity'],
    ['Resume AI', <ResumeAI />, 'Resume AI'],
    ['Recruitment Lifecycle', <RecruitmentLifecycle />, 'Recruitment Lifecycle'],
    ['Billing', <Billing />, 'Billing'],
    ['Not Found', <NotFound />, '404'],
  ])('%s renders its primary UI', async (_name, page, heading, route = '/') => {
    if (_name !== 'Login') setLoggedIn();
    renderPage(page, [route]);
    expect(await screen.findByRole('heading', { name: heading, exact: false })).toBeInTheDocument();
  });
});

describe('authentication UI behavior', () => {
  it('keeps an empty login form from submitting', async () => {
    renderPage(<Login />, ['/login']);
    await screen.findByRole('heading', { name: 'Login' });

    fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    await waitFor(() => {
      expect(mocks.authService.login).not.toHaveBeenCalled();
    });
  });

  it('logs in with the supported demo credentials and navigates to the dashboard', async () => {
    renderPage(<Login />, ['/login']);

    mocks.authService.login.mockResolvedValueOnce({
      data: {
        success: true,
        data: {
          token: 'jwt-token',
          user: {
            id: 1,
            username: 'Admin User',
            email: 'admin@example.com',
            role: 'admin',
            tenantId: 'tenant_demo',
            companyName: 'Demo Company',
          },
          subscriptionStatus: 'ACTIVE',
          planCode: 'legacy',
        },
      },
    });

    fireEvent.change(await screen.findByPlaceholderText('Enter your email'), {
      target: { value: 'admin@example.com' },
    });
    fireEvent.change(await screen.findByLabelText('Password'), {
      target: { value: 'password123' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    await waitFor(() => {
      expect(localStorage.getItem('token')).toBe('jwt-token');
      expect(JSON.parse(localStorage.getItem('user') || '{}').username).toBe('Admin User');
    });
    expect(mocks.toast.success).toHaveBeenCalledWith('Login successful');
  });

  it('rejects mismatched registration passwords', async () => {
    renderPage(<Register />, ['/register']);

    fireEvent.change(await screen.findByPlaceholderText('Enter your name'), {
      target: { value: 'tester' },
    });
    fireEvent.change(await screen.findByPlaceholderText('Enter your email'), {
      target: { value: 'tester@example.com' },
    });
    fireEvent.change(await screen.findByPlaceholderText('Create a password'), {
      target: { value: 'password123' },
    });
    fireEvent.change(await screen.findByPlaceholderText('Confirm your password'), {
      target: { value: 'different123' },
    });
    fireEvent.change(await screen.findByPlaceholderText('Enter your company name'), {
      target: { value: 'Test Company' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Create Tenant Account' }));

    expect(await screen.findByText("Passwords don't match")).toBeInTheDocument();
    expect(mocks.authService.register).not.toHaveBeenCalled();
  });
});

describe('navigation UI behavior', () => {
  it('shows Billing in the authenticated navigation', () => {
    setLoggedIn();
    renderPage(<Navbar />);

    expect(screen.getByRole('link', { name: 'Billing', exact: true })).toHaveAttribute('href', '/billing');
  });

  it('shows Login and Register instead of protected navigation when logged out', () => {
    renderPage(<Navbar />);

    expect(screen.getAllByRole('link', { name: 'Login', exact: true }).some((link) => link.getAttribute('href') === '/login')).toBe(true);
    expect(screen.getByRole('link', { name: 'Register' })).toHaveAttribute('href', '/register');
    expect(screen.queryByRole('link', { name: 'Billing', exact: true })).not.toBeInTheDocument();
  });

  it('logs out from the user menu', async () => {
    setLoggedIn();
    renderPage(<Navbar />);

    const userMenuButton = screen.getAllByRole('button').find(
      (button) => button.getAttribute('aria-haspopup') === 'menu',
    );
    expect(userMenuButton).toBeDefined();
    fireEvent.keyDown(userMenuButton!, { key: 'Enter', code: 'Enter' });
    const logoutItem = await screen.findByRole('menuitem', { name: 'Logout', exact: true });
    expect(logoutItem).toBeInTheDocument();

    fireEvent.click(logoutItem);

    expect(localStorage.getItem('token')).toBeNull();
    expect(localStorage.getItem('user')).toBeNull();
  });
});

describe('recruitment lifecycle UI behavior', () => {
  it('shows the lifecycle stages and blocks downstream actions until prerequisites exist', async () => {
    setLoggedIn();

    mocks.candidateService.getAllCandidates.mockResolvedValueOnce({
      data: { data: [{ id: 12, name: 'Alice Candidate' }] },
    });
    mocks.requirementService.getAllRequirements.mockResolvedValueOnce({
      data: { data: [{ id: 34, title: 'Senior Go Developer', status: 'Open' }] },
    });

    renderPage(<RecruitmentLifecycle />);

    expect(await screen.findByText('Recruitment Lifecycle')).toBeInTheDocument();
    expect(screen.getByText('Select candidate')).toBeInTheDocument();
    expect(screen.getByText('Select candidate')).toBeInTheDocument();
    expect(screen.getByText('Select requirement')).toBeInTheDocument();
  });
});

describe('Admin user management UI behavior', () => {
  it('shows tenant users, seat usage, and protects the admin account', async () => {
    setLoggedIn();

    mocks.authService.getCurrentAccount.mockImplementation(async () => ({
      data: { data: { role: 'admin', userCount: 2, userLimit: 5, companyName: 'Demo Company' } },
    }));
    mocks.userService.getAllUsers.mockImplementation(async () => ({
      data: {
        data: [
          { id: 1, username: 'Admin User', email: 'admin@example.com', role: 'admin' },
          { id: 2, username: 'Recruiter One', email: 'recruiter@example.com', role: 'recruiter' },
        ],
      },
    }));

    renderPage(<AdminUsers />, ['/admin/users']);

    expect(await screen.findByText('2 of 5 users in use')).toBeInTheDocument();
    await waitFor(() => expect(mocks.userService.getAllUsers).toHaveBeenCalled());
    expect(await screen.findByText('Admin User')).toBeInTheDocument();
    expect(screen.getByText('Tenant administrator')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add User' })).toBeEnabled();
  });

  it('creates an additional tenant user with a fixed operational role', async () => {
    setLoggedIn();

    mocks.authService.getCurrentAccount.mockResolvedValueOnce({
      data: { data: { role: 'admin', userCount: 1, userLimit: 3 } },
    });
    mocks.userService.getAllUsers.mockResolvedValueOnce({
      data: { data: [{ id: 1, username: 'Admin User', email: 'admin@example.com', role: 'admin' }] },
    });
    mocks.userService.createUser.mockResolvedValueOnce({
      data: { success: true, data: { id: 2 } },
    });

    renderPage(<AdminUsers />, ['/admin/users']);

    fireEvent.click(await screen.findByRole('button', { name: 'Add User' }));
    fireEvent.change(screen.getByPlaceholderText('User name'), { target: { value: 'New Recruiter' } });
    fireEvent.change(screen.getByPlaceholderText('user@example.com'), { target: { value: 'new@example.com' } });
    fireEvent.change(screen.getByPlaceholderText('Initial password'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: 'Create User' }));

    await waitFor(() =>
      expect(mocks.userService.createUser).toHaveBeenCalledWith({
        username: 'New Recruiter',
        email: 'new@example.com',
        password: 'password123',
        role: 'recruiter',
      }),
    );
  });
});

describe('Account subscription UI behavior', () => {
  it('shows account details and configured subscription plans', async () => {
    setLoggedIn();
    mocks.authService.getCurrentAccount.mockResolvedValueOnce({
      data: { data: { role: 'admin', companyName: 'Demo Company', userCount: 1, userLimit: 5, accountStatus: 'ACTIVE' } },
    });
    mocks.subscriptionService.getSubscription.mockResolvedValueOnce({
      data: { data: { planCode: 'starter', planName: 'Starter', status: 'ACTIVE', provider: 'razorpay', userLimit: 5 } },
    });
    mocks.subscriptionService.getPlans.mockResolvedValueOnce({
      data: { data: [{ code: 'starter', name: 'Starter', amountMinor: 99900, currency: 'INR', billingInterval: 1, billingPeriod: 'month', userLimit: 5 }] },
    });

    renderPage(<Account />, ['/account']);

    expect(await screen.findByRole('heading', { name: 'Account & Subscription' })).toBeInTheDocument();
    expect(await screen.findByText('Demo Company')).toBeInTheDocument();
    expect(await screen.findByText('Starter')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Choose Plan' })).toBeInTheDocument();
  });
});

describe('Billing UI behavior', () => {
  const prepareBillingWorklist = async (billing: unknown = null) => {
    mocks.candidateRecruitmentService.getBillingWorklist.mockResolvedValueOnce({
      data: {
        data: [{
          candidateId: 12,
          candidateName: 'Alice Candidate',
          requirementId: 34,
          requirementJobId: 'REQ-34',
          requirementTitle: 'Senior Go Developer',
          clientId: 56,
          clientName: 'Acme Client',
          joiningId: 78,
          joiningDate: '2026-10-01',
          billingId: billing ? 77 : undefined,
          billingDate: billing ? '2026-09-28T00:00:00.000Z' : undefined,
          amount: billing ? '50000.00' : undefined,
          currency: billing ? 'INR' : undefined,
          invoiceReference: billing ? 'INV-001' : undefined,
          billed: !!billing,
        }],
      },
    });

    renderPage(<Billing />);

    expect(await screen.findByText('Candidate Billing Worklist')).toBeInTheDocument();
    expect(screen.getByText('Alice Candidate')).toBeInTheDocument();
  };

  it('shows joined candidates as billing worklist entries', async () => {
    await prepareBillingWorklist();
    expect(screen.getByText('Pending')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Create Billing' })).toBeInTheDocument();
  });

  it('allows billing creation from a pending worklist entry', async () => {
    await prepareBillingWorklist();

    fireEvent.change(screen.getByLabelText('Amount'), {
      target: { value: '50000' },
    });
    fireEvent.change(screen.getByPlaceholderText('Optional'), {
      target: { value: 'INV-001' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Create Billing' }));

    await waitFor(() =>
      expect(mocks.candidateRecruitmentService.createBilling).toHaveBeenCalledWith(12, 34, {
        amount: '50000',
        currency: 'INR',
        invoiceReference: 'INV-001',
      }),
    );
  });

  it('displays an existing billing record as billed', async () => {
    await prepareBillingWorklist({
      id: 77,
      amount: '50000',
      currency: 'INR',
      invoiceReference: 'INV-001',
    });

    expect(await screen.findByText('INR 50000.00')).toBeInTheDocument();
    expect(screen.getByText('INV-001')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Create Billing' })).not.toBeInTheDocument();
  });
});
