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
  login: async (credentials: any) => {
    return api.post('/auth/login', credentials);
  },

  // Register
  register: async (credentials: any) => {
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
  createCandidate: async (candidate: any) => {
    return api.post('/candidates', candidate);
  },

  // Update a candidate
  updateCandidate: async (id: number, candidate: any) => {
    return api.put(`/candidates/${id}`, candidate);
  },

  // Delete a candidate
  deleteCandidate: async (id: number) => {
    return api.delete(`/candidates/${id}`);
  },
};

export const jobService = {
  // Get all jobs
  getAllJobs: async () => {
    return api.get('/jobs');
  },

  // Get a job by ID
  getJobById: async (id: number) => {
    return api.get(`/jobs/${id}`);
  },

  // Create a new job
  createJob: async (job: any) => {
    // Log job data format
    console.log('Sending job data to API:', job);
    return api.post('/jobs', job);
  },

  // Update a job
  updateJob: async (id: number, job: any) => {
    return api.put(`/jobs/${id}`, job);
  },

  // Delete a job
  deleteJob: async (id: number) => {
    return api.delete(`/jobs/${id}`);
  },
};

export const interviewService = {
  // Get all interviews
  getAllInterviews: async () => {
    return api.get('/interviews');
  },

  // Get an interview by ID
  getInterviewById: async (id: number) => {
    return api.get(`/interviews/${id}`);
  },

  // Create a new interview
  createInterview: async (interview: any) => {
    console.log('Interview scheduled:', interview);
    return api.post('/interviews', interview);
  },

  // Update an interview
  updateInterview: async (id: number, interview: any) => {
    return api.put(`/interviews/${id}`, interview);
  },

  // Delete an interview
  deleteInterview: async (id: number) => {
    return api.delete(`/interviews/${id}`);
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
  createBusinessDev: async (businessDev: any) => {
    return api.post('/business-dev', businessDev);
  },

  // Update a business development
  updateBusinessDev: async (id: number, businessDev: any) => {
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
  createCompany: async (company: any) => {
    return api.post('/companies', company);
  },

  // Update a company
  updateCompany: async (id: string, company: any) => {
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
  createRole: async (role: any) => {
    return api.post('/roles', role);
  },

  // Update a role
  updateRole: async (id: number, role: any) => {
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
  createDailyJob: async (dailyJob: any) => {
    return api.post('/daily-jobs', dailyJob);
  },

  // Update a daily job
  updateDailyJob: async (id: number, dailyJob: any) => {
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
};

export default api;
