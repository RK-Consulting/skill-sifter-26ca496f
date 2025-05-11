
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
  });

  // Fetch jobs
  const { 
    data: jobsData, 
    isLoading: jobsLoading,
    error: jobsError
  } = useQuery({
    queryKey: ['jobs'],
    queryFn: jobService.getAllJobs,
  });

  // Fetch daily jobs
  const { 
    data: dailyJobsData, 
    isLoading: dailyJobsLoading,
    error: dailyJobsError
  } = useQuery({
    queryKey: ['dailyJobs'],
    queryFn: dailyJobService.getAllDailyJobs,
  });

  // Fetch business contacts
  const { 
    data: businessData, 
    isLoading: businessLoading,
    error: businessError
  } = useQuery({
    queryKey: ['businessContacts'],
    queryFn: businessDevService.getAllBusinessDevs,
  });

  // Fetch interviews
  const { 
    data: interviewsData, 
    isLoading: interviewsLoading,
    error: interviewsError
  } = useQuery({
    queryKey: ['interviews'],
    queryFn: interviewService.getAllInterviews,
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

  // Interview statistics
  const totalInterviews = interviewsArray.length;
  const scheduledInterviews = interviewsArray.filter((interview: any) => interview.status === 'Scheduled').length;
  const completedInterviews = interviewsArray.filter((interview: any) => interview.status === 'Completed').length;

  // Determine overall loading state
  const isLoading = candidatesLoading || jobsLoading || dailyJobsLoading || businessLoading || interviewsLoading;
  
  // Determine if there's any error
  const error = candidatesError || jobsError || dailyJobsError || businessError || interviewsError || null;

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
