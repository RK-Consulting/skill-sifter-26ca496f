
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

  // Calculate total candidates from API response
  const totalCandidates = candidatesData?.data?.length || 0;
  
  // Calculate active jobs (filter by status 'open' or 'Active')
  const activeJobs = jobsData?.data?.filter(
    (job: any) => job.status === 'open' || job.status === 'Active'
  )?.length || 0;
  
  // Daily tasks count from backend
  const dailyTasks = dailyJobsData?.data?.length || 0;
  
  // Business contacts count from backend
  const businessContacts = businessData?.data?.length || 0;

  // Interview statistics
  const interviews = interviewsData?.data || [];
  const totalInterviews = interviews.length || 0;
  const scheduledInterviews = interviews.filter((interview: any) => interview.status === 'Scheduled').length || 0;
  const completedInterviews = interviews.filter((interview: any) => interview.status === 'Completed').length || 0;

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
