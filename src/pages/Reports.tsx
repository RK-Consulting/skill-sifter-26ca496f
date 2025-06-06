
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
import { reportsService, HiringReportEntry, SourceReportEntry } from '@/services/reportsService';
import { candidateService, interviewService } from '@/services/api';

const Reports = () => {
  // Fetch hiring statistics
  const { data: hiringData, isLoading: isHiringLoading, error: hiringError } = useQuery({
    queryKey: ['hiringStats'],
    queryFn: reportsService.getHiringStats,
    staleTime: 300000, // 5 minutes
    gcTime: 600000 // 10 minutes
  });

  // Fetch source statistics
  const { data: sourceData, isLoading: isSourceLoading, error: sourceError } = useQuery({
    queryKey: ['sourceStats'],
    queryFn: reportsService.getSourceStats,
    staleTime: 300000, // 5 minutes
    gcTime: 600000 // 10 minutes
  });

  // Fetch candidates for total count
  const { data: candidatesData, isLoading: isCandidatesLoading } = useQuery({
    queryKey: ['candidatesReport'],
    queryFn: candidateService.getAllCandidates,
    staleTime: 300000
  });

  // Fetch interviews for total count
  const { data: interviewsData, isLoading: isInterviewsLoading } = useQuery({
    queryKey: ['interviewsReport'],
    queryFn: interviewService.getAllInterviews,
    staleTime: 300000
  });

  // Colors for the pie chart
  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

  // Transform hiring data for chart display
  const hiringStats = React.useMemo(() => {
    if (hiringData && Array.isArray(hiringData) && hiringData.length > 0) {
      return hiringData.map((entry: HiringReportEntry) => ({
        name: entry.date,
        candidates: entry.totalInterviews
      }));
    }
    return [];
  }, [hiringData]);

  // Transform source data for chart display
  const sourceStats = React.useMemo(() => {
    if (sourceData && Array.isArray(sourceData) && sourceData.length > 0) {
      return sourceData.map((entry: SourceReportEntry) => ({
        name: entry.source,
        value: entry.count
      }));
    }
    return [];
  }, [sourceData]);

  // Calculate total candidates from database
  const totalCandidatesCount = React.useMemo(() => {
    if (candidatesData?.data?.data && Array.isArray(candidatesData.data.data)) {
      return candidatesData.data.data.length;
    }
    return 0;
  }, [candidatesData]);

  // Calculate total interviews from database
  const totalInterviewsCount = React.useMemo(() => {
    if (interviewsData?.data?.data && Array.isArray(interviewsData.data.data)) {
      return interviewsData.data.data.length;
    }
    return 0;
  }, [interviewsData]);

  // Calculate total from hiring trend
  const totalFromHiringTrend = React.useMemo(() => {
    if (hiringStats.length > 0) {
      return hiringStats.reduce((total, month) => total + month.candidates, 0);
    }
    return totalInterviewsCount;
  }, [hiringStats, totalInterviewsCount]);

  // Calculate conversion rate from real data
  const conversionRate = React.useMemo(() => {
    if (totalCandidatesCount > 0 && totalInterviewsCount > 0) {
      const rate = (totalInterviewsCount / totalCandidatesCount * 100).toFixed(1);
      return `${rate}%`;
    }
    return "0%";
  }, [totalCandidatesCount, totalInterviewsCount]);

  const isAnyLoading = isHiringLoading || isSourceLoading || isCandidatesLoading || isInterviewsLoading;
  const hasErrors = hiringError || sourceError;

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Reports & Analytics</h1>
            <p className="text-ats-gray-500">View recruitment insights and metrics from database.</p>
          </div>

          {hasErrors && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-red-600">
                Error loading some report data. Please check your connection to the database.
              </p>
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <Card>
              <CardHeader className="p-6">
                <CardTitle className="flex items-center gap-2">
                  <Users className="w-5 h-5 text-ats-blue" />
                  Total Candidates
                </CardTitle>
              </CardHeader>
              <CardContent className="p-6 pt-0">
                <div className="text-4xl font-bold">
                  {isCandidatesLoading ? (
                    <Skeleton className="h-10 w-24" />
                  ) : (
                    totalCandidatesCount
                  )}
                </div>
                <div className="text-sm text-ats-gray-500 mt-1">
                  From database records
                </div>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader className="p-6">
                <CardTitle className="flex items-center gap-2">
                  <TrendingUp className="w-5 h-5 text-ats-blue" />
                  Total Interviews
                </CardTitle>
              </CardHeader>
              <CardContent className="p-6 pt-0">
                <div className="text-4xl font-bold text-green-600">
                  {isInterviewsLoading ? (
                    <Skeleton className="h-10 w-24" />
                  ) : (
                    totalFromHiringTrend
                  )}
                </div>
                <div className="text-sm text-ats-gray-500 mt-1">
                  Scheduled and completed
                </div>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader className="p-6">
                <CardTitle className="flex items-center gap-2">
                  <Activity className="w-5 h-5 text-ats-blue" />
                  Interview Rate
                </CardTitle>
              </CardHeader>
              <CardContent className="p-6 pt-0">
                <div className="text-4xl font-bold text-ats-blue">
                  {isAnyLoading ? (
                    <Skeleton className="h-10 w-24" />
                  ) : (
                    conversionRate
                  )}
                </div>
                <div className="text-sm text-ats-gray-500 mt-1">
                  Candidates to interviews
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
                ) : hiringStats.length === 0 ? (
                  <div className="h-64 flex items-center justify-center">
                    <div className="text-center text-gray-500">
                      <p className="mb-2">No hiring data available</p>
                      <p className="text-sm">Data will appear here once interviews are scheduled</p>
                    </div>
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
                ) : sourceStats.length === 0 ? (
                  <div className="h-64 flex items-center justify-center">
                    <div className="text-center text-gray-500">
                      <p className="mb-2">No source data available</p>
                      <p className="text-sm">Data will appear here once candidates with sources are added</p>
                    </div>
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
