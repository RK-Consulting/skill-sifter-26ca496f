
import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL;

console.log('API URL configured as:', API_URL);

// Create axios instance with base URL
const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000, // 10 seconds timeout
  withCredentials: true, // Important for CORS with credentials
});

// Add request interceptor to include authorization header if token exists
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
      
      // Log JWT payload for debugging (development only)
      try {
        const parts = token.split('.');
        if (parts.length === 3) {
          const payload = JSON.parse(atob(parts[1]));
          console.debug("[JWT Request]", payload);
        }
      } catch (e) {
        console.error("Error parsing JWT token:", e);
      }
    }
    console.log(`API Request: ${config.method?.toUpperCase()} ${config.url}`);
    return config;
  },
  (error) => Promise.reject(error)
);

// Add response interceptor to handle token expiration and log detailed errors
api.interceptors.response.use(
  (response) => {
    console.log(`API Response: ${response.status} ${response.config.url}`);
    return response;
  },
  (error) => {
    // Log detailed error information for debugging
    console.error('API Error:', error.message);
    
    if (error.response) {
      console.error(`Status: ${error.response.status}, URL: ${error.config?.url}`);
      console.error('Response data:', error.response.data);
    }
    
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

// Business development services
export const businessDevService = {
  getAllBusinessDevs: () => api.get('/business-dev'),
  getBusinessDevById: (id: number) => api.get(`/business-dev/${id}`),
  createBusinessDev: (businessDevData: any) => api.post('/business-dev', businessDevData),
  updateBusinessDev: (id: number, businessDevData: any) => api.put(`/business-dev/${id}`, businessDevData),
  deleteBusinessDev: (id: number) => api.delete(`/business-dev/${id}`),
};

// Interviews services
export const interviewService = {
  getAllInterviews: () => api.get('/interviews'),
  getInterviewById: (id: number) => api.get(`/interviews/${id}`),
  scheduleInterview: (interviewData: any) => api.post('/interviews', interviewData),
  updateInterview: (id: number, interviewData: any) => api.put(`/interviews/${id}`, interviewData),
  deleteInterview: (id: number) => api.delete(`/interviews/${id}`),
};

// Improved health check function that tries multiple endpoints and paths
const checkApiHealth = async () => {
  const endpoints = [
    '/health-check',
    '/api/health-check',
    '/ping',
    '/api/ping'
  ];
  
  for (const endpoint of endpoints) {
    try {
      console.log(`Trying health check at: ${API_URL}${endpoint}`);
      await api.get(endpoint);
      console.log(`✅ API connection successful via ${endpoint}`);
      return true;
    } catch (error: any) {
      console.log(`❌ Failed with ${endpoint}:`, error.message);
      // Continue to next endpoint
    }
  }
  
  // If all attempts failed
  console.error('❌ All health check attempts failed. API may be unreachable.');
  return false;
};

// Run health check on startup
checkApiHealth();

export default api;
