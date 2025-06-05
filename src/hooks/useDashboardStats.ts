
import { useQuery } from '@tanstack/react-query';
import { candidateService, jobService, dailyJobService, businessDevService, interviewService } from '@/services/api';

interface DashboardStats {
  totalCandidates: number;
  activeJobs: number;
  dailyTasks: number;
  businessContacts: number;
  totalInterviews: number;
  scheduledInterviews: number;
  completedInterviews: number;
  isLoading: boolean;
  error: Error | null;
}

export const useDashboardStats = (): DashboardStats => {
  // Fetch candidates
  const { 
    data: candidatesData, 
    isLoading: candidatesLoading,
    error: candidatesError 
  } = useQuery({
    queryKey: ['candidates'],
    queryFn: candidateService.getAllCandidates,
    retry: false, // Don't retry failed requests
  });

  // Fetch jobs
  const { 
    data: jobsData, 
    isLoading: jobsLoading,
    error: jobsError
  } = useQuery({
    queryKey: ['jobs'],
    queryFn: jobService.getAllJobs,
    retry: false,
  });

  // Fetch daily jobs
  const { 
    data: dailyJobsData, 
    isLoading: dailyJobsLoading,
    error: dailyJobsError
  } = useQuery({
    queryKey: ['dailyJobs'],
    queryFn: dailyJobService.getAllDailyJobs,
    retry: false,
  });

  // Fetch business contacts
  const { 
    data: businessData, 
    isLoading: businessLoading,
    error: businessError
  } = useQuery({
    queryKey: ['businessContacts'],
    queryFn: businessDevService.getAllBusinessDevs,
    retry: false,
  });

  // Fetch interviews
  const { 
    data: interviewsData, 
    isLoading: interviewsLoading,
    error: interviewsError
  } = useQuery({
    queryKey: ['interviews'],
    queryFn: interviewService.getAllInterviews,
    retry: false,
  });

  // Safely extract data with array check
  const safeGetArray = (data: any): any[] => {
    if (!data) return [];
    if (Array.isArray(data)) return data;
    if (data.data && Array.isArray(data.data)) return data.data;
    if (data.data && data.data.data && Array.isArray(data.data.data)) return data.data.data;
    console.warn('Expected array data but received:', data);
    return [];
  };

  // Extract data safely
  const candidatesArray = safeGetArray(candidatesData);
  const jobsArray = safeGetArray(jobsData);
  const dailyJobsArray = safeGetArray(dailyJobsData);
  const businessArray = safeGetArray(businessData);
  const interviewsArray = safeGetArray(interviewsData);
  
  // Calculate totals using the safe arrays
  const totalCandidates = candidatesArray.length;
  
  // Calculate active jobs (filter by status 'open' or 'Active')
  const activeJobs = jobsArray.filter(
    (job: any) => job.status === 'open' || job.status === 'Active'
  ).length;
  
  // Daily tasks count from backend
  const dailyTasks = dailyJobsArray.length;
  
  // Business contacts count from backend
  const businessContacts = businessArray.length;

  // Interview statistics with correct status matching
  const totalInterviews = interviewsArray.length;
  const scheduledInterviews = interviewsArray.filter((interview: any) => interview.status === 'scheduled').length;
  const completedInterviews = interviewsArray.filter((interview: any) => interview.status === 'completed').length;

  // Determine overall loading state - only if ALL are loading
  const isLoading = candidatesLoading && jobsLoading && dailyJobsLoading && businessLoading && interviewsLoading;
  
  // Only show error if ALL APIs failed, not just some
  const allFailed = candidatesError && jobsError && dailyJobsError && businessError && interviewsError;
  const error = allFailed ? (candidatesError || jobsError || dailyJobsError || businessError || interviewsError) : null;

  // Log individual errors for debugging without failing the entire dashboard
  if (candidatesError) console.warn('Candidates API failed:', candidatesError.message);
  if (jobsError) console.warn('Jobs API failed:', jobsError.message);
  if (dailyJobsError) console.warn('Daily Jobs API failed:', dailyJobsError.message);
  if (businessError) console.warn('Business Dev API failed:', businessError.message);
  if (interviewsError) console.warn('Interviews API failed:', interviewsError.message);

  return {
    totalCandidates,
    activeJobs,
    dailyTasks,
    businessContacts,
    totalInterviews,
    scheduledInterviews,
    completedInterviews,
    isLoading,
    error
  };
};
