
import { useQuery } from '@tanstack/react-query';
import { candidateService, jobService, dailyJobService, businessDevService } from '@/services/api';

interface DashboardStats {
  totalCandidates: number;
  activeJobs: number;
  dailyTasks: number;
  businessContacts: number;
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

  // Calculate total candidates
  const totalCandidates = candidatesData?.data?.data?.length || 0;
  
  // Calculate active jobs (filter by status 'open')
  const activeJobs = jobsData?.data?.data?.filter(
    (job: any) => job.status === 'open'
  )?.length || 0;
  
  // Daily tasks count
  const dailyTasks = dailyJobsData?.data?.data?.length || 0;
  
  // Business contacts count
  const businessContacts = businessData?.data?.data?.length || 0;

  // Determine overall loading state
  const isLoading = candidatesLoading || jobsLoading || dailyJobsLoading || businessLoading;
  
  // Determine if there's any error
  const error = candidatesError || jobsError || dailyJobsError || businessError || null;

  return {
    totalCandidates,
    activeJobs,
    dailyTasks,
    businessContacts,
    isLoading,
    error
  };
};
