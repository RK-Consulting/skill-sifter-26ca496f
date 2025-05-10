
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, Filter, Briefcase, ChevronRight } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { toast } from "sonner";
import { jobService } from '@/services/api';

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

// Mock data for fallback if API fails
const mockJobs: Job[] = [
  {
    id: 1,
    title: 'Senior Java Developer',
    department: 'Engineering',
    location: 'New York',
    status: 'open',
    datePosted: '2023-01-01',
    description: 'Experienced Java developer needed for backend development.',
    requirements: '5+ years of Java experience'
  },
  {
    id: 2,
    title: 'Frontend Developer',
    department: 'Engineering',
    location: 'San Francisco',
    status: 'active',
    datePosted: '2023-02-15',
    description: 'Passionate frontend developer to build user interfaces.',
    requirements: '3+ years of React experience'
  },
  {
    id: 3,
    title: 'Data Scientist',
    department: 'Data Science',
    location: 'Chicago',
    status: 'closed',
    datePosted: '2023-03-10',
    description: 'Data scientist to analyze and interpret complex data.',
    requirements: 'Master\'s in Statistics or related field'
  },
  {
    id: 4,
    title: 'UX Designer',
    department: 'Design',
    location: 'Los Angeles',
    status: 'open',
    datePosted: '2023-04-01',
    description: 'Creative UX designer to enhance user experience.',
    requirements: '3+ years of UX design experience'
  },
  {
    id: 5,
    title: 'Product Manager',
    department: 'Product',
    location: 'Seattle',
    status: 'active',
    datePosted: '2023-05-01',
    description: 'Experienced product manager to lead product development.',
    requirements: '5+ years of product management experience'
  }
];

const Jobs = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState('');

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const addJob = () => {
    navigate('/jobs/add');
  };

  const viewJobDetails = (id: number) => {
    navigate(`/jobs/${id}`);
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
        // Show toast but still return mock data to avoid UI breaks
        toast.error('Failed to fetch jobs. Using sample data instead.');
        return mockJobs;
      }
    }
  });

  // Filter jobs based on search term
  const filteredJobs = React.useMemo(() => {
    if (!jobsData) return [];

    return jobsData.filter((job: Job) =>
      job.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      job.department.toLowerCase().includes(searchTerm.toLowerCase()) ||
      job.location.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [searchTerm, jobsData]);

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
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
                <div className="py-8 text-center text-ats-gray-500">Loading jobs...</div>
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
                    {filteredJobs.length > 0 ? (
                      filteredJobs.map((job: Job) => (
                        <TableRow key={job.id}>
                          <TableCell className="font-medium">{job.title}</TableCell>
                          <TableCell>{job.department}</TableCell>
                          <TableCell>{job.location}</TableCell>
                          <TableCell>{job.status}</TableCell>
                          <TableCell>{job.datePosted}</TableCell>
                          <TableCell className="text-right">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => viewJobDetails(job.id)}
                            >
                              <ChevronRight size={16} />
                            </Button>
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
