
import axios from 'axios';

// Create axios instance with base URL
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
  headers: {
    'Content-Type': 'application/json',
  },
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

// Authentication services
export const authService = {
  register: (userData: any) => api.post('/auth/register', userData),
  login: (credentials: any) => api.post('/auth/login', credentials),
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

export default api;
