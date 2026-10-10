import React from 'react';
import { 
  Users, 
  Briefcase, 
  Calendar,
  Clock3,
  CheckCircle2, 
  Clock3, 
  XCircle,
  AlertCircle
} from 'lucide-react';
import Container from '../layout/Container';
import DashboardHeader from './DashboardHeader';
import StatsCards from './StatsCards';
import UploadSection from './UploadSection';
import PipelineStatus from './PipelineStatus';
import ActivitySection from './ActivitySection';
import { Card, CardContent } from '../ui-custom/Card';
import { useDashboardStats } from '@/hooks/useDashboardStats';
import { useQuery } from '@tanstack/react-query';
import { reportsService } from '@/services/reportsService';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';

// Converts an ISO timestamp into a short relative string ("2 minutes ago",
// "3 hours ago", "5 days ago"), matching the granularity the mock data used
// to show before this was wired to real data.
function timeAgo(isoTimestamp: string): string {
  const then = new Date(isoTimestamp).getTime();
  const now = Date.now();
  const diffSeconds = Math.max(0, Math.floor((now - then) / 1000));

  if (diffSeconds < 60) return 'just now';
  const diffMinutes = Math.floor(diffSeconds / 60);
  if (diffMinutes < 60) return `${diffMinutes} minute${diffMinutes !== 1 ? 's' : ''} ago`;
  const diffHours = Math.floor(diffMinutes / 60);
  if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
}

interface DashboardProps {
  username: string;
}

const Dashboard = ({ username }: DashboardProps) => {
  // Fetch real-time dashboard stats
  const { 
    totalCandidates, 
    activeRequirements, 
    totalInterviews,
    scheduledInterviews,
    isLoading, 
    error 
  } = useDashboardStats();

  // Stats data with real numbers
  const statsData = [
    {
      title: "Total Candidates",
      value: isLoading ? "..." : totalCandidates.toString(),
      trend: "+12",
      icon: <Users />,
      trendType: "up" as const,
      link: "/candidates"
    },
    {
      title: "Open Requirements",
      value: isLoading ? "..." : activeRequirements.toString(),
      trend: "+2",
      icon: <Briefcase />,
      trendType: "up" as const,
      link: "/requirements"
    },
    {
      title: "Total Interviews",
      value: isLoading ? "..." : totalInterviews.toString(),
      trend: "",
      icon: <Calendar />,
      trendType: "up" as const,
      link: "/interviews"
    },
    {
      title: "Scheduled Interviews",
      value: isLoading ? "..." : scheduledInterviews.toString(),
      trend: "",
      icon: <Clock3 />,
      trendType: "up" as const,
      link: "/interviews"
    },
  ];

  const { data: pipelineReport = [], isLoading: pipelineLoading } = useQuery({
    queryKey: ['pipelineReport'],
    queryFn: reportsService.getPipelineStats,
  });

  const pipelineCounts = Object.fromEntries(
    pipelineReport.map((entry) => [entry.stage, entry.count]),
  );

  const pipelineData = [
    {
      label: "Screening",
      count: pipelineCounts.screening ?? 0,
      icon: <CheckCircle2 className="w-5 h-5 text-green-500" />
    },
    {
      label: "Interview",
      count: pipelineCounts.interview ?? 0,
      icon: <Clock3 className="w-5 h-5 text-yellow-500" />
    },
    {
      label: "Rejected",
      count: pipelineCounts.rejected ?? 0,
      icon: <XCircle className="w-5 h-5 text-red-500" />
    }
  ];

  // Real recent activity, replacing the previously hardcoded array
  const { data: recentActivity } = useQuery({
    queryKey: ['recentActivity'],
    queryFn: reportsService.getRecentActivity,
  });

  const activityData = (recentActivity || []).map((entry) => ({
    title: entry.title,
    description: entry.description,
    time: timeAgo(entry.timestamp),
  }));

  return (
    <section className="py-8 animate-fade-up">
      <Container>
        {/* Header */}
        <DashboardHeader username={username} />

        {/* Error display if there's any API error */}
        {error && (
          <Alert variant="destructive" className="mb-6">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Error loading dashboard data. Please try again later.
            </AlertDescription>
          </Alert>
        )}

        {/* Stats Grid */}
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {[1, 2, 3, 4].map((item) => (
              <Card key={item}>
                <CardContent className="p-6">
                  <div className="flex justify-between items-start mb-4">
                    <Skeleton className="h-10 w-10 rounded-lg" />
                    <Skeleton className="h-6 w-16 rounded-full" />
                  </div>
                  <Skeleton className="h-8 w-16 mb-1" />
                  <Skeleton className="h-5 w-24" />
                </CardContent>
              </Card>
            ))}
          </div>
        ) : (
          <StatsCards stats={statsData} />
        )}

        {/* Actions Row */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          <UploadSection />
          <PipelineStatus items={pipelineData} loading={pipelineLoading} />
        </div>

        {/* Recent Activity */}
        <ActivitySection activities={activityData} />
      </Container>
    </section>
  );
};

export default Dashboard;