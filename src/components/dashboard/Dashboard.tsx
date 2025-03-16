
import React from 'react';
import { 
  Users, 
  Briefcase, 
  Calendar, 
  Store,
  CheckCircle2, 
  Clock3, 
  XCircle
} from 'lucide-react';
import Container from '../layout/Container';
import DashboardHeader from './DashboardHeader';
import StatsCards from './StatsCards';
import UploadSection from './UploadSection';
import PipelineStatus from './PipelineStatus';
import ActivitySection from './ActivitySection';

const Dashboard = () => {
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

  return (
    <section className="py-8 animate-fade-up">
      <Container>
        {/* Header */}
        <DashboardHeader username="Alex" />

        {/* Stats Grid */}
        <StatsCards stats={statsData} />

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
