import { useQuery } from '@tanstack/react-query';
import { candidateService, dailyJobService, businessDevService, interviewService, requirementService } from '@/services/api';

interface DashboardStats {
  totalCandidates: number;
  activeRequirements: number;
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

  // Fetch requirements — the authoritative recruitment-demand resource.
  const {
    data: requirementsData,
    isLoading: requirementsLoading,
    error: requirementsError
  } = useQuery({
    queryKey: ['requirements'],
    queryFn: requirementService.getAllRequirements,
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

  // Minimal shapes for the fields this hook actually reads. Other fields
  // returned by the API are ignored here, hence the index signature.
  interface InterviewRecord {
    status?: string;
    [key: string]: unknown;
  }

  // Safely extract data with array check. API responses sometimes come back
  // as a bare array, or wrapped once (`{ data: [...] }`), or wrapped twice
  // (`{ data: { data: [...] } }`) depending on the endpoint — this defends
  // against all three shapes without assuming which one a given call used.
  const safeGetArray = <T,>(data: unknown): T[] => {
    if (!data) return [];
    if (Array.isArray(data)) return data as T[];
    if (typeof data === 'object' && data !== null && 'data' in data) {
      const inner = (data as { data: unknown }).data;
      if (Array.isArray(inner)) return inner as T[];
      if (typeof inner === 'object' && inner !== null && 'data' in inner) {
        const innerInner = (inner as { data: unknown }).data;
        if (Array.isArray(innerInner)) return innerInner as T[];
      }
    }
    console.warn('Expected array data but received:', data);
    return [];
  };

  // Extract data safely
  const candidatesArray = safeGetArray<Record<string, unknown>>(candidatesData);
  const requirementsArray = safeGetArray<{ status?: string }>(requirementsData);
  const dailyJobsArray = safeGetArray<Record<string, unknown>>(dailyJobsData);
  const businessArray = safeGetArray<Record<string, unknown>>(businessData);
  const interviewsArray = safeGetArray<InterviewRecord>(interviewsData);
  
  // Calculate totals using the safe arrays
  const totalCandidates = candidatesArray.length;
  
  // Calculate active requirements.
  const activeRequirements = requirementsArray.filter(
    (requirement) => requirement.status === 'open'
  ).length;
  
  // Daily tasks count from backend
  const dailyTasks = dailyJobsArray.length;
  
  // Business contacts count from backend
  const businessContacts = businessArray.length;

  // Interview statistics with correct status matching
  const totalInterviews = interviewsArray.length;
  const scheduledInterviews = interviewsArray.filter((interview) => interview.status === 'scheduled').length;
  const completedInterviews = interviewsArray.filter((interview) => interview.status === 'completed').length;

  // Determine overall loading state - only if ALL are loading
  const isLoading = candidatesLoading && requirementsLoading && dailyJobsLoading && businessLoading && interviewsLoading;
  
  // Only show error if ALL APIs failed, not just some
  const allFailed = candidatesError && requirementsError && dailyJobsError && businessError && interviewsError;
  const error = allFailed ? (candidatesError || requirementsError || dailyJobsError || businessError || interviewsError) : null;

  // Log individual errors for debugging without failing the entire dashboard
  if (candidatesError) console.warn('Candidates API failed:', candidatesError.message);
  if (requirementsError) console.warn('Requirements API failed:', requirementsError.message);
  if (dailyJobsError) console.warn('Daily Jobs API failed:', dailyJobsError.message);
  if (businessError) console.warn('Business Dev API failed:', businessError.message);
  if (interviewsError) console.warn('Interviews API failed:', interviewsError.message);

  return {
    totalCandidates,
    activeRequirements,
    dailyTasks,
    businessContacts,
    totalInterviews,
    scheduledInterviews,
    completedInterviews,
    isLoading,
    error
  };
};