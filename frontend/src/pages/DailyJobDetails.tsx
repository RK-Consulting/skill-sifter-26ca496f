import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { User, Calendar, FileText, ArrowLeft, Pencil, Hash } from 'lucide-react';
import Button from '@/components/ui-custom/Button';
import { toast } from 'sonner';
import { dailyJobService } from '@/services/api';
import { Skeleton } from '@/components/ui/skeleton';

interface DailyJob {
  id: number;
  jdNo: number;
  instructions: string;
  assignedUser: number;
  assignedUsername?: string;
  assignedDate: string;
}

const DailyJobDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const { data: job, isLoading, isError } = useQuery({
    queryKey: ['dailyJob', id],
    queryFn: async () => {
      try {
        const response = await dailyJobService.getDailyJobById(Number(id));
        if (response.data && response.data.data) {
          return response.data.data as DailyJob;
        }
        throw new Error('Invalid daily job data format');
      } catch (error) {
        console.error('Error fetching daily job details:', error);
        toast.error('Failed to load daily job details');
        throw error;
      }
    },
  });

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-16">
        <Container>
          <Button
            variant="ghost"
            className="mb-6 flex items-center gap-2"
            onClick={() => navigate('/daily-jobs')}
          >
            <ArrowLeft size={16} />
            Back to Daily Tasks
          </Button>

          {isLoading && (
            <div className="space-y-4">
              <Skeleton className="h-8 w-1/3" />
              <Skeleton className="h-32 w-full" />
            </div>
          )}

          {isError && (
            <Card>
              <CardContent className="p-6 text-center text-gray-500">
                Could not load daily task details.
              </CardContent>
            </Card>
          )}

          {job && (
            <>
              <div className="flex items-start justify-between mb-6">
                <div>
                  <h1 className="text-3xl font-bold">JD #{job.jdNo}</h1>
                </div>
                <Button
                  variant="secondary"
                  className="flex items-center gap-2"
                  onClick={() => navigate(`/daily-jobs/add?editId=${job.id}`)}
                >
                  <Pencil size={16} />
                  Edit
                </Button>
              </div>

              <Card className="mb-6">
                <CardHeader>
                  <CardTitle>Overview</CardTitle>
                </CardHeader>
                <CardContent className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="flex items-center gap-2">
                    <Hash size={16} className="text-gray-400" />
                    <span>JD #{job.jdNo}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <User size={16} className="text-gray-400" />
                    <span>{job.assignedUsername || `User #${job.assignedUser}`}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Calendar size={16} className="text-gray-400" />
                    <span>{new Date(job.assignedDate).toLocaleDateString()}</span>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <FileText size={18} />
                    Instructions
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="whitespace-pre-wrap text-gray-700">
                    {job.instructions || 'No instructions provided.'}
                  </p>
                </CardContent>
              </Card>
            </>
          )}
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default DailyJobDetails;