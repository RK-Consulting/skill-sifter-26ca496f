import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Function to set the JWT token in the request headers
const setAuthToken = (token: string | null) => {
  if (token) {
    api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  } else {
    delete api.defaults.headers.common['Authorization'];
  }
};

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

  // Logout
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
    return api.get('/candidates');
  },

  // Get a candidate by ID
  getCandidateById: async (id: number) => {
    return api.get(`/candidates/${id}`);
  },

  // Create a new candidate
  createCandidate: async (candidate: Record<string, unknown>) => {
    return api.post('/candidates', candidate);
  },

  // Update a candidate
  updateCandidate: async (id: number, candidate: Record<string, unknown>) => {
    return api.put(`/candidates/${id}`, candidate);
  },

  // Delete a candidate
  deleteCandidate: async (id: number) => {
    return api.delete(`/candidates/${id}`);
  },

  // Upload a resume file for a specific candidate. Deterministic — the
  // resume is associated with exactly this candidate, unlike the AI bulk
  // upload (resumeAIService.upload) which matches candidates by parsed
  // content.
  uploadResume: async (id: number, file: File) => {
    const form = new FormData();
    form.append('file', file, file.name);
    return api.post(`/candidates/${id}/resume`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },

  // Get the most recently uploaded resume for a candidate, if any.
  getResume: async (id: number) => {
    return api.get(`/candidates/${id}/resume`);
  },

  getResumeIntelligence: async (id: number) => {
    return api.get(`/candidates/${id}/resume-intelligence`);
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
  // Persistent candidate state is the control plane for recruitment capacity.
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
    return api.get(
      `/api/v1/candidates/${candidateId}/requirements/${requirementId}/submissions`,
    );
  },
  createSubmission: async (
    candidateId: number,
    requirementId: number,
    submission: Record<string, unknown>,
  ) => {
    return api.post(
      `/api/v1/candidates/${candidateId}/requirements/${requirementId}/submissions`,
      submission,
    );
  },
  getSelection: async (candidateId: number, requirementId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/selection`);
  },
  createSelection: async (
    candidateId: number,
    requirementId: number,
    selection: { decision: string; decisionNotes?: string; nextAction?: string },
  ) => {
    return api.post(
      `/api/v1/candidates/${candidateId}/requirements/${requirementId}/selection`,
      selection,
    );
  },
  getOffer: async (candidateId: number, requirementId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/offer`);
  },
  createOffer: async (candidateId: number, requirementId: number) => {
    return api.post(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/offer`);
  },
  updateOffer: async (candidateId: number, requirementId: number, accepted: boolean) => {
    return api.put(
      `/api/v1/candidates/${candidateId}/requirements/${requirementId}/offer`,
      { accepted },
    );
  },
  getJoining: async (candidateId: number, requirementId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/joining`);
  },
  createJoining: async (
    candidateId: number,
    requirementId: number,
    joining: { joiningDate?: string; joined: boolean },
  ) => {
    return api.post(
      `/api/v1/candidates/${candidateId}/requirements/${requirementId}/joining`,
      joining,
    );
  },
  updateJoining: async (
    candidateId: number,
    requirementId: number,
    joining: { joiningDate?: string; joined: boolean },
  ) => {
    return api.put(
      `/api/v1/candidates/${candidateId}/requirements/${requirementId}/joining`,
      joining,
    );
  },
  getBilling: async (candidateId: number, requirementId: number) => {
    return api.get(`/api/v1/candidates/${candidateId}/requirements/${requirementId}/billing`);
  },
  createBilling: async (
    candidateId: number,
    requirementId: number,
    billing: { amount: string; currency: string; invoiceReference?: string },
  ) => {
    return api.post(
      `/api/v1/candidates/${candidateId}/requirements/${requirementId}/billing`,
      billing,
    );
  },
};

export const businessDevService = {
  // Fix business-dev endpoint - removed /list which was causing 400 Bad Request
  getAllBusinessDevs: async () => {
    try {
      return await api.get('/business-dev');
    } catch (error) {
      console.error('Error fetching business dev contacts:', error);
      throw error;
    }
  },

  // Get a business development by ID
  getBusinessDevById: async (id: number) => {
    return api.get(`/business-dev/${id}`);
  },

  // Create a new business development
  createBusinessDev: async (businessDev: Record<string, unknown>) => {
    return api.post('/business-dev', businessDev);
  },

  // Update a business development
  updateBusinessDev: async (id: number, businessDev: Record<string, unknown>) => {
    return api.put(`/business-dev/${id}`, businessDev);
  },

  // Delete a business development
  deleteBusinessDev: async (id: number) => {
    return api.delete(`/business-dev/${id}`);
  },
};

export const companyService = {
  // Get all companies
  getAllCompanies: async () => {
    return api.get('/companies');
  },

  // Get a company by ID
  getCompanyById: async (id: string) => {
    return api.get(`/companies/${id}`);
  },

  // Create a new company
  createCompany: async (company: Record<string, unknown>) => {
    return api.post('/companies', company);
  },

  // Update a company
  updateCompany: async (id: string, company: Record<string, unknown>) => {
    return api.put(`/companies/${id}`, company);
  },

  // Delete a company
  deleteCompany: async (id: string) => {
    return api.delete(`/companies/${id}`);
  },
};

export const roleService = {
  // Get all roles
  getAllRoles: async () => {
    return api.get('/roles');
  },

  // Get a role by ID
  getRoleById: async (id: number) => {
    return api.get(`/roles/${id}`);
  },

  // Create a new role
  createRole: async (role: Record<string, unknown>) => {
    return api.post('/roles', role);
  },

  // Update a role
  updateRole: async (id: number, role: Record<string, unknown>) => {
    return api.put(`/roles/${id}`, role);
  },

  // Delete a role
  deleteRole: async (id: number) => {
    return api.delete(`/roles/${id}`);
  },
};

export const userService = {
  // Fix company-users endpoint - removed /list which was causing issues
  getAllUsers: async () => {
    try {
      return await api.get('/company-users');
    } catch (error) {
      console.error('Error fetching users:', error);
      throw error;
    }
  },
};

export const dailyJobService = {
  // Get all daily jobs
  getAllDailyJobs: async () => {
    return api.get('/daily-jobs');
  },

  // Get a daily job by ID
  getDailyJobById: async (id: number) => {
    return api.get(`/daily-jobs/${id}`);
  },

  // Create a new daily job
  createDailyJob: async (dailyJob: Record<string, unknown>) => {
    return api.post('/daily-jobs', dailyJob);
  },

  // Update a daily job
  updateDailyJob: async (id: number, dailyJob: Record<string, unknown>) => {
    return api.put(`/daily-jobs/${id}`, dailyJob);
  },

  // Delete a daily job
  deleteDailyJob: async (id: number) => {
    return api.delete(`/daily-jobs/${id}`);
  },
};

// Add report services to match backend handlers
export const reportService = {
  // Get hiring report data
  getHiringReport: async () => {
    return api.get('/reports/hiring');
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