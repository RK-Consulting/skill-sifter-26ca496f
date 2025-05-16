
import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { TrendingUp, Activity, Users, BarChart2 } from 'lucide-react';
import { toast } from 'sonner';
import { Skeleton } from '@/components/ui/skeleton';
import { reportsService } from '@/services/reportsService';

const Reports = () => {
  // Fetch hiring statistics
  const { data: hiringData, isLoading: isHiringLoading } = useQuery({
    queryKey: ['hiringStats'],
    queryFn: reportsService.getHiringStats,
    staleTime: 300000, // 5 minutes
    gcTime: 600000 // 10 minutes
  });

  // Fetch source statistics
  const { data: sourceData, isLoading: isSourceLoading } = useQuery({
    queryKey: ['sourceStats'],
    queryFn: reportsService.getSourceStats,
    staleTime: 300000, // 5 minutes
    gcTime: 600000 // 10 minutes
  });

  // Colors for the pie chart
  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

  // Use fallback data if API returns an error
  const hiringStats = hiringData || [
    { name: 'Jan', candidates: 4 },
    { name: 'Feb', candidates: 7 },
    { name: 'Mar', candidates: 5 },
    { name: 'Apr', candidates: 10 },
    { name: 'May', candidates: 8 },
    { name: 'Jun', candidates: 12 },
  ];

  const sourceStats = sourceData || [
    { name: 'LinkedIn', value: 40 },
    { name: 'Referrals', value: 25 },
    { name: 'Job Boards', value: 20 },
    { name: 'Direct', value: 15 },
  ];

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Reports & Analytics</h1>
            <p className="text-ats-gray-500">View recruitment insights and metrics.</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <Card>
              <CardHeader className="p-6">
                <CardTitle className="flex items-center gap-2">
                  <Users className="w-5 h-5 text-ats-blue" />
                  Candidates
                </CardTitle>
              </CardHeader>
              <CardContent className="p-6 pt-0">
                <div className="text-4xl font-bold">
                  {isHiringLoading ? (
                    <Skeleton className="h-10 w-24" />
                  ) : (
                    hiringStats.reduce((total, month) => total + month.candidates, 0) || 0
                  )}
                </div>
                <div className="text-sm text-ats-gray-500 mt-1">
                  Total candidates this year
                </div>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader className="p-6">
                <CardTitle className="flex items-center gap-2">
                  <TrendingUp className="w-5 h-5 text-ats-blue" />
                  Growth Rate
                </CardTitle>
              </CardHeader>
              <CardContent className="p-6 pt-0">
                <div className="text-4xl font-bold text-green-600">
                  {isHiringLoading ? (
                    <Skeleton className="h-10 w-24" />
                  ) : (
                    "+24%"
                  )}
                </div>
                <div className="text-sm text-ats-gray-500 mt-1">
                  Year over year increase
                </div>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader className="p-6">
                <CardTitle className="flex items-center gap-2">
                  <Activity className="w-5 h-5 text-ats-blue" />
                  Conversion Rate
                </CardTitle>
              </CardHeader>
              <CardContent className="p-6 pt-0">
                <div className="text-4xl font-bold text-ats-blue">
                  {isSourceLoading ? (
                    <Skeleton className="h-10 w-24" />
                  ) : (
                    "18%"
                  )}
                </div>
                <div className="text-sm text-ats-gray-500 mt-1">
                  Applications to hires
                </div>
              </CardContent>
            </Card>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
            <Card>
              <CardHeader className="p-6">
                <CardTitle className="flex items-center gap-2">
                  <BarChart2 className="w-5 h-5 text-ats-blue" />
                  Hiring Trend
                </CardTitle>
              </CardHeader>
              <CardContent className="p-6 pt-0">
                {isHiringLoading ? (
                  <div className="h-64 flex items-center justify-center">
                    <Skeleton className="h-full w-full" />
                  </div>
                ) : (
                  <div className="h-64">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart
                        data={hiringStats}
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
                )}
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
                {isSourceLoading ? (
                  <div className="h-64 flex items-center justify-center">
                    <Skeleton className="h-full w-full" />
                  </div>
                ) : (
                  <div className="h-64 flex items-center justify-center">
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChart>
                        <Pie
                          data={sourceStats}
                          cx="50%"
                          cy="50%"
                          innerRadius={60}
                          outerRadius={90}
                          fill="#8884d8"
                          paddingAngle={5}
                          dataKey="value"
                          label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                        >
                          {sourceStats.map((entry, index) => (
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
                )}
              </CardContent>
            </Card>
          </div>
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default Reports;
