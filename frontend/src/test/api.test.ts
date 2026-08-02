import { describe, it, expect, vi } from 'vitest';

// Capture the mock instance so we can assert on calls made through it.
const mockPost = vi.fn(() => Promise.resolve({ data: { success: true } }));
const mockGet = vi.fn(() => Promise.resolve({ data: { success: true, data: [] } }));

vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      post: mockPost,
      get: mockGet,
      put: vi.fn(() => Promise.resolve({ data: { success: true } })),
      delete: vi.fn(() => Promise.resolve({ data: { success: true } })),
      defaults: { headers: { common: {} } },
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() },
      },
    })),
  },
}));

describe('candidateService.createCandidate', () => {
  it('POSTs to /candidates with the exact payload it was given (no silent field renaming/dropping)', async () => {
    const { candidateService } = await import('@/services/api');

    // Matches backend/models/models.go's Candidate struct JSON tags after the
    // schema-mismatch fix (docs/architecture.md). If a field name below needs
    // updating, check the backend Candidate struct was updated to match too —
    // that exact drift is what caused the production bug this test guards
    // against.
    const payload = {
      name: 'Test Candidate',
      email: 'test@example.com',
      phone: '9999999999',
      position: 'Software Engineer',
      location: 'Bangalore',
      experience: '5 years',
      currentCTC: '10 LPA',
      expectedCTC: '15 LPA',
      noticePeriod: '30 days',
      jlptLanguage: 'N/A',
      skills: 'Go, React',
      newJD: 'Job description text',
    };

    await candidateService.createCandidate(payload);

    expect(mockPost).toHaveBeenCalledWith('/candidates', payload);
  });

  it('getAllCandidates GETs /candidates', async () => {
    const { candidateService } = await import('@/services/api');
    await candidateService.getAllCandidates();
    expect(mockGet).toHaveBeenCalledWith('/candidates');
  });
});