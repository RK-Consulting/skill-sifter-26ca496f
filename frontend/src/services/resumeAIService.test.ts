import { describe, expect, it, vi, beforeEach } from 'vitest';
import { resumeAIService } from './resumeAIService';
import api from './api';

vi.mock('./api', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
  },
}));

describe('resumeAIService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('uploads selected files through the files multipart field', async () => {
    const post = vi.mocked(api.post);
    post.mockResolvedValueOnce({ data: { success: true } } as never);

    const first = new File(['resume-a'], 'resume-a.txt', { type: 'text/plain' });
    const second = new File(['resume-b'], 'resume-b.pdf', { type: 'application/pdf' });

    await resumeAIService.upload([first, second]);

    expect(post).toHaveBeenCalledTimes(1);
    const [url, form, config] = post.mock.calls[0];
    expect(url).toBe('/resume-ai/upload');
    expect(form).toBeInstanceOf(FormData);
    expect((form as FormData).getAll('files')).toHaveLength(2);
    expect(config).toEqual({ headers: { 'Content-Type': 'multipart/form-data' } });
  });

  it('sends the resume search query', async () => {
    const get = vi.mocked(api.get);
    get.mockResolvedValueOnce({ data: { success: true, data: [] } } as never);

    await resumeAIService.search('Go PostgreSQL');

    expect(get).toHaveBeenCalledWith('/resume-ai/search', { params: { q: 'Go PostgreSQL' } });
  });

  it('uses the resume repository and health endpoints', async () => {
    const get = vi.mocked(api.get);
    get.mockResolvedValue({ data: { success: true } } as never);

    await resumeAIService.list();
    await resumeAIService.health();

    expect(get).toHaveBeenNthCalledWith(1, '/resume-ai/resumes');
    expect(get).toHaveBeenNthCalledWith(2, '/resume-ai/health');
  });
});
