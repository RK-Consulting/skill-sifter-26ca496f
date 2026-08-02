import React from 'react';
import { 
  Users, 
  Briefcase, 
  Calendar, 
  Store,
  CheckCircle2, 
  Clock3, 
  XCircle,
  TrendingUp,
  Activity,
  AlertCircle
} from 'lucide-react';
import Container from '../layout/Container';
import DashboardHeader from './DashboardHeader';
import StatsCards from './StatsCards';
import PipelineStatus from './PipelineStatus';
import ActivitySection from './ActivitySection';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '../ui-custom/Card';
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
    activeJobs, 
    dailyTasks, 
    businessContacts, 
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
      title: "Active Jobs",
      value: isLoading ? "..." : activeJobs.toString(),
      trend: "+2",
      icon: <Briefcase />,
      trendType: "up" as const,
      link: "/jobs"
    },
    {
      title: "Daily Tasks",
      value: isLoading ? "..." : dailyTasks.toString(),
      trend: "+3",
      icon: <Calendar />,
      trendType: "up" as const,
      link: "/daily-jobs"
    },
    {
      title: "Business Contacts",
      value: isLoading ? "..." : businessContacts.toString(),
      trend: "+4",
      icon: <Store />,
      trendType: "up" as const,
      link: "/business-dev"
    }
  ];

  // Real pipeline counts, replacing the previously hardcoded 24/12/8 values.
  // "Screening" = candidates with no interview record yet (no dedicated
  // field exists for this — see docs/architecture.md section 12.4).
  const { data: pipeline } = useQuery({
    queryKey: ['pipeline'],
    queryFn: reportsService.getPipeline,
  });

  const pipelineData = [
    {
      label: "Screening",
      count: pipeline?.screening ?? 0,
      icon: <CheckCircle2 className="w-5 h-5 text-green-500" />
    },
    {
      label: "Interview",
      count: pipeline?.interview ?? 0,
      icon: <Clock3 className="w-5 h-5 text-yellow-500" />
    },
    {
      label: "Rejected",
      count: pipeline?.rejected ?? 0,
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

  // Real hiring trend, replacing the previously hardcoded array. Backend
  // returns { date: "2026-01", totalInterviews: N } — reformatted here into
  // the { name: "Jan", candidates: N } shape the chart expects.
  const { data: hiringStats } = useQuery({
    queryKey: ['hiringStats'],
    queryFn: reportsService.getHiringStats,
  });

  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  const hiringData = (hiringStats || []).map((entry) => {
    const [, monthNum] = entry.date.split('-');
    const monthIndex = parseInt(monthNum, 10) - 1;
    return {
      name: monthNames[monthIndex] ?? entry.date,
      candidates: entry.totalInterviews,
    };
  });

  const sourceData = [
    { name: 'LinkedIn', value: 40 },
    { name: 'Referrals', value: 25 },
    { name: 'Job Boards', value: 20 },
    { name: 'Direct', value: 15 },
  ];

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

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

        {/* Charts Row */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <Card>
            <CardHeader className="p-6">
              <CardTitle className="flex items-center gap-2">
                <TrendingUp className="w-5 h-5 text-ats-blue" />
                Hiring Trend
              </CardTitle>
            </CardHeader>
            <CardContent className="p-6 pt-0">
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart
                    data={hiringData}
                    margin={{ top: 10, right: 10, left: 10, bottom: 20 }}
                  >
                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                    <XAxis dataKey="name" />
                    <YAxis />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'white',
                        borderRadius: '8px',
                        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
                        border: 'none'
                      }}
                    />
                    <Bar dataKey="candidates" fill="#0A84FF" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="p-6">
              <CardTitle className="flex items-center gap-2">
                <Activity className="w-5 h-5 text-ats-blue" />
                Candidate Sources
              </CardTitle>
            </CardHeader>
            <CardContent className="p-6 pt-0">
              <div className="h-64 flex items-center justify-center">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={sourceData}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={90}
                      fill="#8884d8"
                      paddingAngle={5}
                      dataKey="value"
                      label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                    >
                      {sourceData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'white',
                        borderRadius: '8px',
                        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
                        border: 'none'
                      }}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Actions Row */}
        {/* Upload Resumes removed — not functional yet; AI-based resume
            extraction is a future milestone (docs/architecture.md section 11) */}
        <div className="grid grid-cols-1 gap-6 mb-8">
          <PipelineStatus items={pipelineData} />
        </div>

        {/* Recent Activity */}
        <ActivitySection activities={activityData} />
      </Container>
    </section>
  );
};

export default Dashboard;