import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, Filter, Briefcase, ChevronRight, Pencil, Eye } from 'lucide-react';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { toast } from "sonner";
import { jobService } from '@/services/api';
import { Skeleton } from '@/components/ui/skeleton';

interface Job {
  id: number;
  title: string;
  department: string;
  location: string;
  status: string;
  datePosted: string;
  description: string;
  requirements: string;
}

const Jobs = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState('');
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(true);
  
  // Check authentication status
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      setIsAuthenticated(false);
      navigate('/login', { replace: true });
    } else {
      setIsAuthenticated(true);
    }
  }, [navigate]);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const addJob = () => {
    navigate('/jobs/add');
  };

  const viewJobDetails = (id: number) => {
    // Added a try-catch block and toast notification for job navigation
    try {
      console.log(`Navigating to job details page for job ID: ${id}`);
      navigate(`/jobs/${id}`);
    } catch (error) {
      console.error('Error navigating to job details:', error);
      toast.error('Failed to open job details. Please try again.');
    }
  };
  
  // Fetch jobs using React Query with proper error handling
  const { data: jobsData, isLoading, isError } = useQuery({
    queryKey: ['jobs'],
    queryFn: async () => {
      try {
        console.log('Attempting to fetch jobs from API...');
        const response = await jobService.getAllJobs();
        console.log('Jobs API response:', response.data);
        return response.data.data || [];
      } catch (error) {
        console.error('Error fetching jobs:', error);
        
        // Show toast only if we're still authenticated
        if (isAuthenticated) {
          toast.error('Failed to fetch jobs. Please try again later.');
        }
        return [];
      }
    },
    // Only enable the query if we're authenticated
    enabled: isAuthenticated
  });

  // Filter jobs based on search term. Guards against any job record having a
  // missing/null field (only title is NOT NULL in the schema) — without this,
  // .toLowerCase() on undefined throws, and with no error boundary anywhere
  // in the app, React unmounts the whole tree to a blank page.
  const filteredJobs = React.useMemo(() => {
    if (!jobsData) return [];

    return jobsData.filter((job: Job) =>
      (job.title || '').toLowerCase().includes(searchTerm.toLowerCase()) ||
      (job.department || '').toLowerCase().includes(searchTerm.toLowerCase()) ||
      (job.location || '').toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [searchTerm, jobsData]);

  // If we're not authenticated, don't render anything
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container className="px-4 md:px-6 mx-auto max-w-7xl">
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Jobs</h1>
            <p className="text-ats-gray-500">Manage job postings and track applications.</p>
          </div>

          <Card className="mb-8">
            <CardContent className="p-6">
              <div className="flex flex-col md:flex-row justify-between gap-4 mb-6">
                <div className="relative w-full md:w-80">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-ats-gray-400" size={18} />
                  <Input
                    placeholder="Search jobs..."
                    className="pl-10"
                    value={searchTerm}
                    onChange={handleSearch}
                  />
                </div>

                <div className="flex gap-3">
                  <Button variant="outline" size="sm" className="flex gap-2">
                    <Filter size={16} />
                    Filter
                  </Button>
                  <Button
                    variant="primary"
                    size="sm"
                    className="flex gap-2"
                    onClick={addJob}
                  >
                    <Briefcase size={16} />
                    Add Job
                  </Button>
                </div>
              </div>

              {isLoading ? (
                <div className="space-y-2">
                  {[1, 2, 3, 4, 5].map((i) => (
                    <Skeleton key={i} className="h-12 w-full" />
                  ))}
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Title</TableHead>
                      <TableHead>Department</TableHead>
                      <TableHead>Location</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Date Posted</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {isError ? (
                      <TableRow>
                        <TableCell colSpan={6} className="text-center py-8 text-red-600">
                          Error loading jobs. Please try refreshing the page.
                        </TableCell>
                      </TableRow>
                    ) : filteredJobs.length > 0 ? (
                      filteredJobs.map((job: Job) => (
                        <TableRow key={job.id}>
                          <TableCell className="font-medium">{job.title}</TableCell>
                          <TableCell>{job.department}</TableCell>
                          <TableCell>{job.location}</TableCell>
                          <TableCell>{job.status}</TableCell>
                          <TableCell>{job.datePosted}</TableCell>
                          <TableCell className="text-right">
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="sm">
                                  <ChevronRight size={16} />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end" className="w-[160px]">
                                <DropdownMenuItem onClick={() => viewJobDetails(job.id)}>
                                  <Eye size={14} className="mr-2" />
                                  View Details
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => navigate(`/jobs/add?editId=${job.id}`)}>
                                  <Pencil size={14} className="mr-2" />
                                  Edit
                                </DropdownMenuItem>
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                          {searchTerm ? 'No jobs found matching your search.' : 'No jobs found.'}
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default Jobs;