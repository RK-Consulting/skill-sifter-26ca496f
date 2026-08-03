import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { Briefcase, MapPin, Calendar, FileText, ArrowLeft, Pencil } from 'lucide-react';
import Button from '@/components/ui-custom/Button';
import { toast } from 'sonner';
import { jobService } from '@/services/api';
import { Skeleton } from '@/components/ui/skeleton';

interface Job {
  id: number;
  title: string;
  department: string;
  location: string;
  status: string;
  datePosted: string;
  description?: string;
  requirements?: string;
}

const JobDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const { data: job, isLoading, isError } = useQuery({
    queryKey: ['job', id],
    queryFn: async () => {
      try {
        const response = await jobService.getJobById(Number(id));
        if (response.data && response.data.data) {
          return response.data.data as Job;
        }
        throw new Error('Invalid job data format');
      } catch (error) {
        console.error('Error fetching job details:', error);
        toast.error('Failed to load job details');
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
            onClick={() => navigate('/jobs')}
          >
            <ArrowLeft size={16} />
            Back to Jobs
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
                Could not load job details.
              </CardContent>
            </Card>
          )}

          {job && (
            <>
              <div className="flex items-start justify-between mb-6">
                <div>
                  <h1 className="text-3xl font-bold">{job.title}</h1>
                  <p className="text-gray-500 mt-1">{job.department}</p>
                </div>
                <Button
                  variant="secondary"
                  className="flex items-center gap-2"
                  onClick={() => navigate(`/jobs/add?editId=${job.id}`)}
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
                    <MapPin size={16} className="text-gray-400" />
                    <span>{job.location}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Briefcase size={16} className="text-gray-400" />
                    <span className="capitalize">{job.status}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Calendar size={16} className="text-gray-400" />
                    <span>{new Date(job.datePosted).toLocaleDateString()}</span>
                  </div>
                </CardContent>
              </Card>

              <Card className="mb-6">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <FileText size={18} />
                    Description
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="whitespace-pre-wrap text-gray-700">
                    {job.description || 'No description provided.'}
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Requirements</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="whitespace-pre-wrap text-gray-700">
                    {job.requirements || 'No requirements listed.'}
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

export default JobDetails;