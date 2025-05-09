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
  Activity
} from 'lucide-react';
import Container from '../layout/Container';
import DashboardHeader from './DashboardHeader';
import StatsCards from './StatsCards';
import UploadSection from './UploadSection';
import PipelineStatus from './PipelineStatus';
import ActivitySection from './ActivitySection';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '../ui-custom/Card';

interface DashboardProps {
  username: string;
}

const Dashboard = ({ username }: DashboardProps) => {
  // Stats data
  const statsData = [
    {
      title: "Total Candidates",
      value: "148",
      trend: "+12",
      icon: <Users />,
      trendType: "up" as const,
      link: "/candidates"
    },
    {
      title: "Active Jobs",
      value: "12",
      trend: "+2",
      icon: <Briefcase />,
      trendType: "up" as const,
      link: "/jobs"
    },
    {
      title: "Daily Tasks",
      value: "8",
      trend: "+3",
      icon: <Calendar />,
      trendType: "up" as const,
      link: "/daily-jobs"
    },
    {
      title: "Business Contacts",
      value: "24",
      trend: "+4",
      icon: <Store />,
      trendType: "up" as const,
      link: "/business-dev"
    }
  ];

  // Pipeline data
  const pipelineData = [
    {
      label: "Screening",
      count: 24,
      icon: <CheckCircle2 className="w-5 h-5 text-green-500" />
    },
    {
      label: "Interview",
      count: 12,
      icon: <Clock3 className="w-5 h-5 text-yellow-500" />
    },
    {
      label: "Rejected",
      count: 8,
      icon: <XCircle className="w-5 h-5 text-red-500" />
    }
  ];

  // Activity data
  const activityData = [
    {
      title: "New candidate application",
      description: "Sarah Wilson applied for Senior UI Designer",
      time: "2 minutes ago"
    },
    {
      title: "New business contact added",
      description: "TechSolutions Inc added as a new client",
      time: "1 hour ago"
    },
    {
      title: "Daily task assigned",
      description: "Follow up with ClientX for feedback assigned to Alex",
      time: "2 hours ago"
    }
  ];

  // Chart data
  const hiringData = [
    { name: 'Jan', candidates: 4 },
    { name: 'Feb', candidates: 7 },
    { name: 'Mar', candidates: 5 },
    { name: 'Apr', candidates: 10 },
    { name: 'May', candidates: 8 },
    { name: 'Jun', candidates: 12 },
  ];

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

        {/* Stats Grid */}
        <StatsCards stats={statsData} />

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
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          <UploadSection />
          <PipelineStatus items={pipelineData} />
        </div>

        {/* Recent Activity */}
        <ActivitySection activities={activityData} />
      </Container>
    </section>
  );
};

export default Dashboard;
