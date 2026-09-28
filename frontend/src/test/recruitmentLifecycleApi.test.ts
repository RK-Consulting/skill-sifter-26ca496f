import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockGet = vi.fn(() => Promise.resolve({ data: { success: true, data: {} } }));
const mockPost = vi.fn(() => Promise.resolve({ data: { success: true, data: {} } }));
const mockPut = vi.fn(() => Promise.resolve({ data: { success: true, data: {} } }));

vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      get: mockGet,
      post: mockPost,
      put: mockPut,
      delete: vi.fn(),
      defaults: { headers: { common: {} } },
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() },
      },
    })),
  },
}));

describe('candidateRecruitmentService lifecycle APIs', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('uses the Candidate × Requirement offer endpoint for create and accept', async () => {
    const { candidateRecruitmentService } = await import('@/services/api');

    await candidateRecruitmentService.createOffer(12, 34);
    await candidateRecruitmentService.updateOffer(12, 34, true);

    expect(mockPost).toHaveBeenCalledWith('/api/v1/candidates/12/requirements/34/offer');
    expect(mockPut).toHaveBeenCalledWith(
      '/api/v1/candidates/12/requirements/34/offer',
      { accepted: true },
    );
  });

  it('uses the Candidate × Requirement joining endpoint', async () => {
    const { candidateRecruitmentService } = await import('@/services/api');
    const joining = { joiningDate: '2026-10-01', joined: true };

    await candidateRecruitmentService.createJoining(12, 34, joining);
    await candidateRecruitmentService.updateJoining(12, 34, joining);

    expect(mockPost).toHaveBeenCalledWith(
      '/api/v1/candidates/12/requirements/34/joining',
      joining,
    );
    expect(mockPut).toHaveBeenCalledWith(
      '/api/v1/candidates/12/requirements/34/joining',
      joining,
    );
  });

  it('uses the tenant billing worklist endpoint', async () => {
    const { candidateRecruitmentService } = await import('@/services/api');

    await candidateRecruitmentService.getBillingWorklist();

    expect(mockGet).toHaveBeenCalledWith('/api/v1/billing');
  });

  it('uses the Candidate × Requirement billing GET endpoint', async () => {
    const { candidateRecruitmentService } = await import('@/services/api');

    await candidateRecruitmentService.getBilling(12, 34);

    expect(mockGet).toHaveBeenCalledWith(
      '/api/v1/candidates/12/requirements/34/billing',
    );
  });

  it('uses the Candidate × Requirement billing POST endpoint with explicit commercial fields', async () => {
    const { candidateRecruitmentService } = await import('@/services/api');
    const billing = {
      amount: '50000',
      currency: 'INR',
      invoiceReference: 'INV-001',
    };

    await candidateRecruitmentService.createBilling(12, 34, billing);

    expect(mockPost).toHaveBeenCalledWith(
      '/api/v1/candidates/12/requirements/34/billing',
      billing,
    );
  });
  it('uses screening and submission endpoints', async () => {
    const { candidateRecruitmentService } = await import('@/services/api');

    await candidateRecruitmentService.getScreenings(12);
    await candidateRecruitmentService.createScreening(12, { requirementId: 34, recruiterUserId: 7 });
    await candidateRecruitmentService.updateScreening(12, 56, 'completed');
    await candidateRecruitmentService.getSubmissions(12, 34);
    await candidateRecruitmentService.createSubmission(12, 34, { recipientType: 'client', recipientClientId: 9 });

    expect(mockGet).toHaveBeenCalledWith('/api/v1/candidates/12/screenings');
    expect(mockPost).toHaveBeenCalledWith('/api/v1/candidates/12/screenings', {
      requirementId: 34,
      recruiterUserId: 7,
    });
    expect(mockPut).toHaveBeenCalledWith('/api/v1/candidates/12/screenings/56', { status: 'completed' });
    expect(mockGet).toHaveBeenCalledWith('/api/v1/candidates/12/requirements/34/submissions');
    expect(mockPost).toHaveBeenCalledWith('/api/v1/candidates/12/requirements/34/submissions', {
      recipientType: 'client',
      recipientClientId: 9,
    });
  });

  it('uses submission feedback endpoints', async () => {
    const { candidateRecruitmentService } = await import('@/services/api');
    const feedback = { outcome: 'shortlist', comments: 'Proceed to interview' };

    await candidateRecruitmentService.getSubmissionFeedback(88);
    await candidateRecruitmentService.createSubmissionFeedback(88, feedback);

    expect(mockGet).toHaveBeenCalledWith('/api/v1/submissions/88/feedback');
    expect(mockPost).toHaveBeenCalledWith('/api/v1/submissions/88/feedback', feedback);
  });

});
