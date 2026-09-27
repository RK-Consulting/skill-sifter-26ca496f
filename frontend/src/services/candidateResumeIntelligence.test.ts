import { describe, expect, it, vi, beforeEach } from 'vitest';
import { candidateService } from './api';
import api from './api';

vi.mock('./api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
  candidateService: {
    getAllCandidates: vi.fn(),
    getCandidateById: vi.fn(),
    createCandidate: vi.fn(),
    updateCandidate: vi.fn(),
    deleteCandidate: vi.fn(),
    uploadResume: vi.fn(),
    getResume: vi.fn(),
    getResumeIntelligence: vi.fn(),
  },
}));

describe('candidate Resume AI service contract', () => {
  beforeEach(() => vi.clearAllMocks());

  it('requests candidate Resume AI intelligence through the tenant-scoped candidate endpoint', async () => {
    const get = vi.mocked(api.get);
    get.mockResolvedValueOnce({ data: { success: true, data: {} } } as never);

    await candidateService.getResumeIntelligence(42);

    expect(get).toHaveBeenCalledWith('/candidates/42/resume-intelligence');
  });
});
