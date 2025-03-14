
import React from 'react';
import { 
  BarChart3, 
  Briefcase, 
  Users, 
  Clock, 
  TrendingUp, 
  CheckCircle2, 
  Clock3, 
  XCircle,
  ChevronRight,
  Upload
} from 'lucide-react';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '../ui-custom/Card';
import Container from '../layout/Container';
import Button from '../ui-custom/Button';
import { cn } from '@/lib/utils'; // Import the cn utility function

const Dashboard = () => {
  return (
    <section className="py-8 animate-fade-up">
      <Container>
        {/* Header */}
        <div className="mb-8">
          <div className="inline-flex items-center px-3 py-1 rounded-full bg-ats-blue/10 text-ats-blue text-sm font-medium mb-4">
            <BarChart3 className="w-4 h-4 mr-2" />
            Dashboard Overview
          </div>
          <h1 className="text-3xl font-semibold tracking-tight">Welcome back, Alex</h1>
          <p className="text-ats-gray-500 mt-2">Here's what's happening with your recruitment pipeline.</p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <StatsCard 
            title="Active Jobs"
            value="12"
            trend="+2"
            icon={<Briefcase />}
            trendType="up"
          />
          <StatsCard 
            title="Total Candidates"
            value="148"
            trend="+12"
            icon={<Users />}
            trendType="up"
          />
          <StatsCard 
            title="Time to Hire"
            value="18 days"
            trend="-2"
            icon={<Clock />}
            trendType="down"
            positive
          />
          <StatsCard 
            title="Conversion Rate"
            value="24%"
            trend="+4%"
            icon={<TrendingUp />}
            trendType="up"
          />
        </div>

        {/* Actions Row */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          <Card hover className="col-span-full lg:col-span-1">
            <CardContent className="p-6">
              <div className="flex flex-col items-center text-center">
                <div className="w-12 h-12 rounded-full bg-ats-blue/10 text-ats-blue flex items-center justify-center mb-4">
                  <Upload className="w-6 h-6" />
                </div>
                <h3 className="text-lg font-semibold mb-2">Upload Resumes</h3>
                <p className="text-sm text-ats-gray-500 mb-4">
                  Batch upload resumes to process multiple candidates at once
                </p>
                <Button variant="primary" className="w-full">
                  Upload Files
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Pipeline Status */}
          <Card className="col-span-full lg:col-span-2">
            <CardHeader className="p-6">
              <CardTitle>Recruitment Pipeline</CardTitle>
            </CardHeader>
            <CardContent className="px-6">
              <div className="space-y-4">
                <PipelineItem
                  label="Screening"
                  count={24}
                  icon={<CheckCircle2 className="w-5 h-5 text-green-500" />}
                />
                <PipelineItem
                  label="Interview"
                  count={12}
                  icon={<Clock3 className="w-5 h-5 text-yellow-500" />}
                />
                <PipelineItem
                  label="Rejected"
                  count={8}
                  icon={<XCircle className="w-5 h-5 text-red-500" />}
                />
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Recent Activity */}
        <Card>
          <CardHeader className="p-6 flex flex-row items-center justify-between">
            <CardTitle>Recent Activity</CardTitle>
            <Button variant="ghost" size="sm" className="text-ats-blue">
              View All
            </Button>
          </CardHeader>
          <CardContent className="px-6">
            <div className="space-y-4">
              <ActivityItem
                title="New candidate application"
                description="Sarah Wilson applied for Senior UI Designer"
                time="2 minutes ago"
              />
              <ActivityItem
                title="Interview scheduled"
                description="Technical interview for John Doe - Software Engineer"
                time="1 hour ago"
              />
              <ActivityItem
                title="Candidate hired"
                description="Mike Johnson accepted the offer for Product Manager"
                time="2 hours ago"
              />
            </div>
          </CardContent>
        </Card>
      </Container>
    </section>
  );
};

// Stats Card Component
const StatsCard = ({ title, value, trend, icon, trendType, positive = false }) => (
  <Card hover>
    <CardContent className="p-6">
      <div className="flex justify-between items-start mb-4">
        <div className="p-2 rounded-lg bg-ats-gray-100/50">
          {React.cloneElement(icon, { className: "w-5 h-5 text-ats-gray-600" })}
        </div>
        <div className={cn(
          "text-sm font-medium px-2 py-1 rounded-full",
          positive ? "bg-green-100 text-green-700" : "bg-ats-blue/10 text-ats-blue"
        )}>
          {trend}
        </div>
      </div>
      <h3 className="text-2xl font-semibold mb-1">{value}</h3>
      <p className="text-sm text-ats-gray-500">{title}</p>
    </CardContent>
  </Card>
);

// Pipeline Item Component
const PipelineItem = ({ label, count, icon }) => (
  <div className="flex items-center justify-between p-4 rounded-lg border border-ats-gray-200 hover:border-ats-gray-300 transition-colors">
    <div className="flex items-center space-x-3">
      {icon}
      <span className="font-medium">{label}</span>
    </div>
    <div className="flex items-center">
      <span className="text-ats-gray-500 mr-2">{count} candidates</span>
      <ChevronRight className="w-4 h-4 text-ats-gray-400" />
    </div>
  </div>
);

// Activity Item Component
const ActivityItem = ({ title, description, time }) => (
  <div className="flex items-start space-x-4 p-4 rounded-lg hover:bg-ats-gray-50 transition-colors">
    <div className="flex-1">
      <h4 className="font-medium mb-1">{title}</h4>
      <p className="text-sm text-ats-gray-500">{description}</p>
    </div>
    <time className="text-xs text-ats-gray-400">{time}</time>
  </div>
);

export default Dashboard;
