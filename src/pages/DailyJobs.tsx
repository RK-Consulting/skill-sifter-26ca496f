
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, Filter, PlusCircle, ChevronRight } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { toast } from 'sonner';
import { dailyJobService } from '@/services/api';

interface DailyJob {
  id: number;
  jdNo: number;
  instructions: string;
  assignedUser: number;
  assignedDate: string; // We'll format this from the timestamp
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const diffDays = Math.floor(diff / (1000 * 60 * 60 * 24));
  
  if (diffDays === 0) return 'Today';
  if (diffDays === 1) return 'Yesterday';
  if (diffDays < 7) return `${diffDays} days ago`;
  return `${Math.floor(diffDays / 7)} week${Math.floor(diffDays / 7) !== 1 ? 's' : ''} ago`;
};

// Mock data for fallback if API fails
const mockDailyJobs: DailyJob[] = [
  {
    id: 1,
    jdNo: 1001,
    instructions: 'Source candidates for Senior Java Developer position',
    assignedUser: 1,
    assignedDate: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: 2,
    jdNo: 1002,
    instructions: 'Review resumes for Frontend Developer candidates',
    assignedUser: 2,
    assignedDate: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: 3,
    jdNo: 1003,
    instructions: 'Prepare interview questions for DevOps Engineer',
    assignedUser: 1,
    assignedDate: new Date().toISOString(),
  },
  {
    id: 4,
    jdNo: 1004,
    instructions: 'Follow up with candidates from yesterday interviews',
    assignedUser: 3,
    assignedDate: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: 5,
    jdNo: 1005,
    instructions: 'Update job descriptions for open positions',
    assignedUser: 2,
    assignedDate: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
  },
];

const DailyJobs = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState('');

  // Fetch daily jobs using React Query with fallback to mock data
  const { data: dailyJobsData, isLoading, isError } = useQuery({
    queryKey: ['dailyJobs'],
    queryFn: async () => {
      try {
        const response = await dailyJobService.getAllDailyJobs();
        console.log('Daily jobs API response:', response.data);
        return response.data.data;
      } catch (error) {
        console.error('Error fetching daily jobs:', error);
        // Return mock data on error
        return mockDailyJobs;
      }
    }
  });

  // Filter jobs based on search term
  const filteredJobs = React.useMemo(() => {
    if (!dailyJobsData) return [];
    
    return dailyJobsData.filter((job: DailyJob) => 
      job.instructions.toLowerCase().includes(searchTerm.toLowerCase()) ||
      job.jdNo.toString().includes(searchTerm)
    );
  }, [searchTerm, dailyJobsData]);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const addDailyJob = () => {
    navigate('/daily-jobs/add');
  };

  const viewJobDetails = (id: number) => {
    // In a real application, this would navigate to a job details page
    console.log(`View daily job ${id}`);
    toast.info(`Viewing details for job #${id}`);
  };

  // Handle loading and error states
  if (isLoading) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar />
        <main className="pt-24 pb-10">
          <Container>
            <div className="flex justify-center items-center h-64">
              <p className="text-lg text-ats-gray-500">Loading daily job assignments...</p>
            </div>
          </Container>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Daily Job Assignments</h1>
            <p className="text-ats-gray-500">Manage and track daily tasks assigned to team members.</p>
          </div>

          <Card className="mb-8">
            <CardContent className="p-6">
              <div className="flex flex-col md:flex-row justify-between gap-4 mb-6">
                <div className="relative w-full md:w-80">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-ats-gray-400" size={18} />
                  <Input 
                    placeholder="Search assignments..." 
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
                    onClick={addDailyJob}
                  >
                    <PlusCircle size={16} />
                    Add Assignment
                  </Button>
                </div>
              </div>

              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>JD No</TableHead>
                    <TableHead>Instructions</TableHead>
                    <TableHead>Assigned User</TableHead>
                    <TableHead>Assigned Date</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredJobs.length > 0 ? (
                    filteredJobs.map((job: DailyJob) => (
                      <TableRow key={job.id}>
                        <TableCell className="font-medium">{job.jdNo}</TableCell>
                        <TableCell>{job.instructions}</TableCell>
                        <TableCell>User #{job.assignedUser}</TableCell>
                        <TableCell>{formatDate(job.assignedDate)}</TableCell>
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
                      <TableCell colSpan={5} className="text-center py-8 text-gray-500">
                        {searchTerm ? 'No assignments found matching your search.' : 'No assignments found.'}
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </Container>
      </main>
    </div>
  );
};

export default DailyJobs;
