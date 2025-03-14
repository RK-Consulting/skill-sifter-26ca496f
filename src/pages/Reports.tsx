
import React from 'react';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { BarChart, Calendar, Download, TrendingUp, TrendingDown, Users, Clock, CheckCircle, XCircle } from 'lucide-react';
import Button from '@/components/ui-custom/Button';
import { BarChart as RechartsBarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { cn } from '@/lib/utils';

const Reports = () => {
  // Example data for the charts
  const hiringData = [
    { name: 'Jan', candidates: 12 },
    { name: 'Feb', candidates: 19 },
    { name: 'Mar', candidates: 15 },
    { name: 'Apr', candidates: 21 },
    { name: 'May', candidates: 18 },
    { name: 'Jun', candidates: 24 },
    { name: 'Jul', candidates: 28 },
    { name: 'Aug', candidates: 20 },
  ];

  const timeToHireData = [
    { name: 'Engineering', days: 24 },
    { name: 'Design', days: 18 },
    { name: 'Marketing', days: 15 },
    { name: 'Product', days: 22 },
    { name: 'Sales', days: 12 },
  ];

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Reports</h1>
            <p className="text-ats-gray-500">Analyze your recruitment metrics and performance.</p>
          </div>
          
          {/* Top Stats */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            <StatCard 
              title="Candidates This Month"
              value="48"
              change="+12%"
              trend="up"
              icon={<Users size={20} />}
            />
            <StatCard 
              title="Average Time to Hire"
              value="18 days"
              change="-3 days"
              trend="down"
              positive
              icon={<Clock size={20} />}
            />
            <StatCard 
              title="Successful Hires"
              value="8"
              change="+2"
              trend="up"
              icon={<CheckCircle size={20} />}
            />
            <StatCard 
              title="Rejection Rate"
              value="24%"
              change="-5%"
              trend="down"
              positive
              icon={<XCircle size={20} />}
            />
          </div>
          
          {/* Charts */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg">Candidates Over Time</CardTitle>
                  <Button variant="ghost" size="sm" className="h-8 gap-1">
                    <Calendar size={14} />
                    <span>This Year</span>
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <div className="h-80">
                  <ResponsiveContainer width="100%" height="100%">
                    <RechartsBarChart data={hiringData} margin={{ top: 20, right: 30, left: 20, bottom: 20 }}>
                      <CartesianGrid strokeDasharray="3 3" vertical={false} />
                      <XAxis dataKey="name" />
                      <YAxis />
                      <Tooltip />
                      <Bar dataKey="candidates" fill="#3b82f6" radius={[4, 4, 0, 0]} />
                    </RechartsBarChart>
                  </ResponsiveContainer>
                </div>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg">Time to Hire by Department</CardTitle>
                  <Button variant="outline" size="sm" className="h-8 gap-1">
                    <Download size={14} />
                    <span>Export</span>
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <div className="h-80">
                  <ResponsiveContainer width="100%" height="100%">
                    <RechartsBarChart
                      layout="vertical"
                      data={timeToHireData}
                      margin={{ top: 20, right: 30, left: 50, bottom: 20 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" horizontal={false} />
                      <XAxis type="number" />
                      <YAxis dataKey="name" type="category" />
                      <Tooltip />
                      <Bar dataKey="days" fill="#10b981" radius={[0, 4, 4, 0]} />
                    </RechartsBarChart>
                  </ResponsiveContainer>
                </div>
              </CardContent>
            </Card>
          </div>
          
          {/* Summary */}
          <Card>
            <CardHeader>
              <CardTitle>Recruitment Summary</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <SummaryItem 
                  title="Open Positions"
                  value="12"
                  description="4 urgent positions"
                  icon={<BarChart size={20} className="text-blue-500" />}
                />
                <SummaryItem 
                  title="Time in Pipeline"
                  value="15 days avg."
                  description="3 days faster than last month"
                  icon={<Clock size={20} className="text-amber-500" />}
                />
                <SummaryItem 
                  title="Offer Acceptance"
                  value="92%"
                  description="Higher than industry average (85%)"
                  icon={<CheckCircle size={20} className="text-green-500" />}
                />
              </div>
            </CardContent>
          </Card>

        </Container>
      </main>
    </div>
  );
};

// Stat Card Component
const StatCard = ({ title, value, change, trend, positive = false, icon }) => (
  <Card>
    <CardContent className="p-6">
      <div className="flex justify-between mb-3">
        <div className="p-2 bg-ats-gray-100/50 rounded-lg">
          {icon}
        </div>
        <div className={cn(
          "flex items-center px-2 py-1 rounded-full text-xs font-medium",
          trend === "up" && !positive ? "bg-ats-blue/10 text-ats-blue" : "",
          trend === "down" && !positive ? "bg-red-100 text-red-700" : "",
          trend === "up" && positive ? "bg-green-100 text-green-700" : "",
          trend === "down" && positive ? "bg-green-100 text-green-700" : ""
        )}>
          {trend === "up" ? <TrendingUp size={12} className="mr-1" /> : <TrendingDown size={12} className="mr-1" />}
          {change}
        </div>
      </div>
      <h3 className="text-2xl font-semibold mb-1">{value}</h3>
      <p className="text-ats-gray-500 text-sm">{title}</p>
    </CardContent>
  </Card>
);

// Summary Item Component
const SummaryItem = ({ title, value, description, icon }) => (
  <div className="flex items-start">
    <div className="p-3 rounded-lg bg-ats-gray-100/50 mr-4">
      {icon}
    </div>
    <div>
      <h3 className="font-medium mb-1">{title}</h3>
      <p className="text-2xl font-semibold mb-1">{value}</p>
      <p className="text-ats-gray-500 text-sm">{description}</p>
    </div>
  </div>
);

export default Reports;
