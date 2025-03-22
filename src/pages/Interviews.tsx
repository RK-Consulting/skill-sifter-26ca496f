
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, Filter, Calendar, ChevronRight, Plus } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { toast } from "sonner";
import { interviewService } from '@/services/api';

interface Interview {
  id: number;
  candidateName: string;
  position: string;
  interviewDate: string;
  status: string;
  feedback: string;
}

const Interviews = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = React.useState('');
  
  // Fetch interviews using React Query
  const { data: interviewsData, isLoading, isError } = useQuery({
    queryKey: ['interviews'],
    queryFn: async () => {
      const response = await interviewService.getAllInterviews();
      return response.data.data;
    }
  });

  // Filter interviews based on search term
  const filteredInterviews = React.useMemo(() => {
    if (!interviewsData) return [];
    
    return interviewsData.filter((interview: Interview) => 
      interview.candidateName.toLowerCase().includes(searchTerm.toLowerCase()) ||
      interview.position.toLowerCase().includes(searchTerm.toLowerCase()) ||
      interview.status.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [searchTerm, interviewsData]);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const scheduleNewInterview = () => {
    navigate('/interviews/schedule');
    toast.success("Navigating to schedule interview form");
  };

  const viewInterviewDetails = (id: number) => {
    navigate(`/interviews/${id}`);
  };

  // Handle loading and error states
  if (isLoading) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar />
        <main className="pt-24 pb-10">
          <Container>
            <div className="flex justify-center items-center h-64">
              <p className="text-lg text-ats-gray-500">Loading interviews...</p>
            </div>
          </Container>
        </main>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar />
        <main className="pt-24 pb-10">
          <Container>
            <div className="flex flex-col justify-center items-center h-64">
              <p className="text-lg text-red-500 mb-4">Failed to load interviews</p>
              <Button variant="outline" onClick={() => window.location.reload()}>
                Try Again
              </Button>
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
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Interviews</h1>
            <p className="text-ats-gray-500">Manage and track all candidate interviews.</p>
          </div>

          <Card className="mb-8 animate-fade-up">
            <CardContent className="p-6">
              <div className="flex flex-col md:flex-row justify-between gap-4 mb-6">
                <div className="relative w-full md:w-80">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-ats-gray-400" size={18} />
                  <Input 
                    placeholder="Search interviews..." 
                    className="pl-10 transition-all focus:ring-2 focus:ring-ats-blue/20"
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
                    onClick={scheduleNewInterview}
                  >
                    <Plus size={16} />
                    Schedule Interview
                  </Button>
                </div>
              </div>

              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Candidate</TableHead>
                    <TableHead>Position</TableHead>
                    <TableHead>Interview Date & Time</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Feedback</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredInterviews.length > 0 ? (
                    filteredInterviews.map((interview: Interview) => {
                      // Format the date string
                      const formattedDate = new Date(interview.interviewDate).toLocaleString();
                      
                      return (
                        <TableRow key={interview.id} className="hover:bg-ats-gray-50 transition-colors cursor-pointer" onClick={() => viewInterviewDetails(interview.id)}>
                          <TableCell className="font-medium">{interview.candidateName}</TableCell>
                          <TableCell>{interview.position}</TableCell>
                          <TableCell>
                            <div className="flex items-center">
                              <Calendar className="w-4 h-4 mr-2 text-ats-gray-400" />
                              {formattedDate}
                            </div>
                          </TableCell>
                          <TableCell>
                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
                              ${interview.status === 'Completed' ? 'bg-green-100 text-green-800' : ''}
                              ${interview.status === 'Scheduled' ? 'bg-blue-100 text-blue-800' : ''}
                              ${interview.status === 'Cancelled' ? 'bg-red-100 text-red-800' : ''}
                            `}>
                              {interview.status}
                            </span>
                          </TableCell>
                          <TableCell>
                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
                              ${interview.feedback === 'Selected' ? 'bg-green-100 text-green-800' : ''}
                              ${interview.feedback === 'Rejected' ? 'bg-red-100 text-red-800' : ''}
                              ${interview.feedback === 'Pending' ? 'bg-yellow-100 text-yellow-800' : ''}
                            `}>
                              {interview.feedback}
                            </span>
                          </TableCell>
                          <TableCell className="text-right">
                            <Button 
                              variant="ghost" 
                              size="sm"
                              onClick={(e) => {
                                e.stopPropagation();
                                viewInterviewDetails(interview.id);
                              }}
                            >
                              <ChevronRight size={16} />
                            </Button>
                          </TableCell>
                        </TableRow>
                      );
                    })
                  ) : (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                        {searchTerm ? 'No interviews found matching your search.' : 'No interviews found.'}
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

export default Interviews;
