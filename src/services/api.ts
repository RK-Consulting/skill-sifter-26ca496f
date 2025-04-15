
import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

console.log('API URL configured as:', API_URL);

// Create axios instance with base URL
const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000, // 10 seconds timeout
});

// Add request interceptor to include authorization header if token exists
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Add response interceptor to handle token expiration
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error.message);
    
    if (error.response && error.response.status === 401) {
      // Token has expired or is invalid
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      // Redirect to login page if not already there
      if (window.location.pathname !== '/login') {
        window.location.href = '/login';
      }
    }
    
    // Add more specific error information
    if (!error.response && error.code === 'ERR_NETWORK') {
      error.serverDown = true;
      console.error('Server connection failed. Is the backend running?');
    }
    
    return Promise.reject(error);
  }
);

// Authentication services
export const authService = {
  register: (userData: any) => api.post('/auth/register', userData),
  login: (credentials: any) => api.post('/auth/login', credentials),
};

// User management services (admin only)
export const userService = {
  getAllUsers: () => api.get('/admin/users'),
  createUser: (userData: any) => api.post('/admin/users', userData),
  updateUser: (id: number, userData: any) => api.put(`/admin/users/${id}`, userData),
  deleteUser: (id: number) => api.delete(`/admin/users/${id}`),
};

// Role services (admin only)
export const roleService = {
  getAllRoles: () => api.get('/admin/roles'),
};

// Candidates services
export const candidateService = {
  getAllCandidates: () => api.get('/candidates'),
  getCandidateById: (id: number) => api.get(`/candidates/${id}`),
  createCandidate: (candidateData: any) => api.post('/candidates', candidateData),
  updateCandidate: (id: number, candidateData: any) => api.put(`/candidates/${id}`, candidateData),
  deleteCandidate: (id: number) => api.delete(`/candidates/${id}`),
};

// Jobs services
export const jobService = {
  getAllJobs: () => api.get('/jobs'),
  getJobById: (id: number) => api.get(`/jobs/${id}`),
  createJob: (jobData: any) => api.post('/jobs', jobData),
  updateJob: (id: number, jobData: any) => api.put(`/jobs/${id}`, jobData),
  deleteJob: (id: number) => api.delete(`/jobs/${id}`),
};

// Daily jobs services
export const dailyJobService = {
  getAllDailyJobs: () => api.get('/daily-jobs'),
  getDailyJobById: (id: number) => api.get(`/daily-jobs/${id}`),
  createDailyJob: (dailyJobData: any) => api.post('/daily-jobs', dailyJobData),
  updateDailyJob: (id: number, dailyJobData: any) => api.put(`/daily-jobs/${id}`, dailyJobData),
  deleteDailyJob: (id: number) => api.delete(`/daily-jobs/${id}`),
};

// Interviews services
export const interviewService = {
  getAllInterviews: () => api.get('/interviews'),
  getInterviewById: (id: number) => api.get(`/interviews/${id}`),
  scheduleInterview: (interviewData: any) => api.post('/interviews', interviewData),
  updateInterview: (id: number, interviewData: any) => api.put(`/interviews/${id}`, interviewData),
  deleteInterview: (id: number) => api.delete(`/interviews/${id}`),
};

// Check API connectivity
api.get('/health-check')
  .then(() => console.log('✅ API connection successful'))
  .catch(error => {
    if (error.code === 'ERR_NETWORK') {
      console.error('❌ Cannot connect to API server. Please make sure the backend is running.');
    } else {
      console.error('❌ API health check failed:', error.message);
    }
  });

export default api;
