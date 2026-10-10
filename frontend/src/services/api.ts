import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add a request interceptor to include the JWT token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    // Ensure all requests include /api/ prefix
    if (config.url && !config.url.startsWith('/api/') && !config.url.startsWith('api/')) {
      config.url = `/api${config.url.startsWith('/') ? config.url : `/${config.url}`}`;
    }
    
    console.log(`Sending request to: ${config.baseURL}${config.url}`);
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add a response interceptor for better error handling
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    console.error('API Error:', error.message);
    if (error.response) {
      console.error('Status:', error.response.status, 'URL:', error.config?.url);
      console.error('Response data:', error.response.data);
      
      // If unauthorized, handle authentication errors
      if (error.response.status === 401) {
        console.log('Unauthorized access, redirecting to login');
        // We'll just log it but not immediately redirect to prevent disrupting user experience
        // localStorage.removeItem('token');
        // localStorage.removeItem('user');
        // window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export const authService = {
  // Login
  login: async (credentials: Record<string, unknown>) => {
    return api.post('/auth/login', credentials);
  },

  // Register
  register: async (credentials: Record<string, unknown>) => {
    return api.post('/auth/register', credentials);
  },

  verifyEmail: async (registrationId: number, code: string) => {
    return api.post('/auth/register/verify-email', { registrationId, code });
  },

  // Logout
  getCurrentAccount: async () => {
    return api.get('/account');
  },

  logout: async () => {
    return api.post('/auth/logout');
  },

  // Forgot Password
  forgotPassword: async (email: string) => {
    return api.post('/auth/forgot-password', { email });
  },

  // Reset Password
  resetPassword: async (token: string, newPassword: string) => {
    return api.post(`/auth/reset-password/${token}`, { newPassword });
  },
};

export const candidateService = {
  // Get all candidates
  getAllCandidates: async () => {
    return api.get('/api/v1/candidates');
  },

  // Get a candidate by ID
  getCandidateById: async (id: number) => {
    return api.get(`/api/v1/candidates/${id}`);
  },

  // Create a new candidate
  createCandidate: async (candidate: Record<string, unknown>) => {
    return api.post('/api/v1/candidates', candidate);
  },

  // Update a candidate
  updateCandidate: async (id: number, candidate: Record<string, unknown>) => {
    return api.put(`/api/v1/candidates/${id}`, candidate);
  },

  // Delete a candidate
  deleteCandidate: async (id: number) => {
    return api.delete(`/api/v1/candidates/${id}`);
  },

  // Upload a resume file for a specific candidate. Deterministic — the
  // resume is associated with exactly this candidate, unlike the AI bulk
  // upload (resumeAIService.upload) which matches candidates by parsed
  // content.
  uploadResume: async (id: number, file: File) => {
    const form = new FormData();
    form.append('file', file, file.name);
    return api.post(`/api/v1/candidates/${id}/resume`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },

  // Get the most recently uploaded resume for a candidate, if any.
  getResume: async (id: number) => {
    return api.get(`/api/v1/candidates/${id}/resume`);
  },

  getResumeIntelligence: async (id: number) => {
    return api.get(`/api/v1/candidates/${id}/resume-intelligence`);
  },
};

export const interviewService = {
  // Phase 5 V1 interview workflow. The request interceptor prefixes /api.
  getAllInterviews: async () => {
    return api.get('/v1/interviews');
  },

  getInterviewById: async (id: number) => {
    return api.get(`/v1/interviews/${id}`);
  },

  createInterview: async (interview: Record<string, unknown>) => {
    return api.post('/v1/interviews', interview);
  },

  updateInterview: async (id: number, interview: Record<string, unknown>) => {
    return api.put(`/v1/interviews/${id}`, interview);
  },


};

export const candidateRecruitmentService = {
  getScreenings: async (candidateId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/screenings`);
  },
  createScreening: async (candidateId: number, screening: Record<string, unknown>) => {
    return api.post(`/api/v1/candidates/${candidateId}/screenings`, screening);
  },
  updateScreening: async (candidateId: number, screeningId: number, status: string) => {
    return api.put(`/api/v1/candidates/${candidateId}/screenings/${screeningId}`, { status });
  },
  getInterviews: async (candidateId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/interviews`);
  },
  getSubmissions: async (candidateId: number, requirementId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/submissions`);
  },
  createSubmission: async (candidateId: number, requirementId: number, submission: Record<string, unknown>) => {
    return api.post(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/submissions`, submission);
  },
  getSubmissionFeedback: async (submissionId: number) => {
    return api.get(`/api/v1/submissions/${submissionId}/feedback`);
  },
  createSubmissionFeedback: async (
    submissionId: number,
    feedback: { outcome: string; reasonCode?: string; comments?: string; nextAction?: string },
  ) => {
    return api.post(`/api/v1/submissions/${submissionId}/feedback`, feedback);
  },
  getSelection: async (candidateId: number, requirementId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/selection`);
  },
  createSelection: async (candidateId: number, requirementId: number, selection: { decision: string; decisionNotes?: string; nextAction?: string }) => {
    return api.post(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/selection`, selection);
  },
  getOffer: async (candidateId: number, requirementId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/offer`);
  },
  createOffer: async (candidateId: number, requirementId: number) => {
    return api.post(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/offer`);
  },
  updateOffer: async (candidateId: number, requirementId: number, accepted: boolean) => {
    return api.put(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/offer`, { accepted });
  },
  getJoining: async (candidateId: number, requirementId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/joining`);
  },
  createJoining: async (candidateId: number, requirementId: number, joining: { joiningDate?: string; joined: boolean }) => {
    return api.post(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/joining`, joining);
  },
  updateJoining: async (candidateId: number, requirementId: number, joining: { joiningDate?: string; joined: boolean }) => {
    return api.put(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/joining`, joining);
  },
  getBillingWorklist: async () => {
    return api.get('/api/v1/billing');
  },
  getBilling: async (candidateId: number, requirementId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/billing`);
  },
  createBilling: async (candidateId: number, requirementId: number, billing: { amount: string; currency: string; invoiceReference?: string }) => {
    return api.post(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/billing`, billing);
  },
};

export const subscriptionService = {
  getPlans: async () => api.get('/account/plans'),
  getSubscription: async () => api.get('/account/subscription'),
  checkout: async (planCode: string) => api.post('/account/subscription/checkout', { planCode }),
  sendPhoneVerification: async (phone: string) => api.post('/account/subscription/phone/send', { phone }),
  verifyPhoneVerification: async (code: string) => api.post('/account/subscription/phone/verify', { code }),
  cancel: async () => api.post('/account/subscription/cancel'),
};

export const userService = {
  getAllUsers: async () => {
    return api.get('/admin/users');
  },

  createUser: async (user: { username: string; email: string; password: string; role: string }) => {
    return api.post('/admin/users', user);
  },

  updateUser: async (id: number, user: { username?: string; email?: string; role?: string }) => {
    return api.put(`/admin/users/${id}`, user);
  },

  deleteUser: async (id: number) => {
    return api.delete(`/admin/users/${id}`);
  },
};

// Add report services to match backend handlers
export const reportService = {
  // Get hiring report data
  getHiringReport: async () => {
    return api.get('/reports/hiring');
  },

  // Get recruitment pipeline report data
  getPipelineReport: async () => {
    return api.get('/reports/pipeline');
  },

  // Get source report data
  getSourceReport: async () => {
    return api.get('/reports/sources');
  },

  // Get recent activity feed
  getRecentActivity: async () => {
    return api.get('/reports/activity');
  },
};

// --- V1 domain services (ADR 0008: /api/v1) ---
// These call /api/v1/... directly (not the bare resource path) because
// the request interceptor above only auto-prepends /api/ when the URL is
// missing it entirely; passing the full /api/v1/... path here avoids that
// rewrite so it isn't collapsed back to /api/....

export const clientService = {
  // Paginated (ADR 0008). page/limit are 1-indexed; status is optional.
  getAllClients: async (params?: { page?: number; limit?: number; status?: string }) => {
    return api.get('/api/v1/clients', { params });
  },
  getClientById: async (id: number) => {
    return api.get(`/api/v1/clients/${id}`);
  },
  createClient: async (client: Record<string, unknown>) => {
    return api.post('/api/v1/clients', client);
  },
  updateClient: async (id: number, client: Record<string, unknown>) => {
    return api.put(`/api/v1/clients/${id}`, client);
  },
  deleteClient: async (id: number) => {
    return api.delete(`/api/v1/clients/${id}`);
  },
};

export const requirementService = {
  // Not paginated — GetRequirements returns the full tenant list.
  getAllRequirements: async () => {
    return api.get('/api/v1/requirements');
  },
  getRequirementById: async (id: number) => {
    return api.get(`/api/v1/requirements/${id}`);
  },
  createRequirement: async (requirement: Record<string, unknown>) => {
    return api.post('/api/v1/requirements', requirement);
  },
  updateRequirement: async (id: number, requirement: Record<string, unknown>) => {
    return api.put(`/api/v1/requirements/${id}`, requirement);
  },
  deleteRequirement: async (id: number) => {
    return api.delete(`/api/v1/requirements/${id}`);
  },
};

export default api;