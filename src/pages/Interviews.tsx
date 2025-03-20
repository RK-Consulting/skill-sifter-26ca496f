
import React from 'react';
import { useNavigate } from 'react-router-dom';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, Filter, Calendar, ChevronRight, Plus } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { toast } from "sonner";

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
  
  // Sample interview data
  const interviews: Interview[] = [
    { 
      id: 1, 
      candidateName: 'Sarah Wilson', 
      position: 'Senior UI Designer', 
      interviewDate: '2023-05-15 10:00 AM', 
      status: 'Completed', 
      feedback: 'Selected' 
    },
    { 
      id: 2, 
      candidateName: 'John Doe', 
      position: 'Software Engineer', 
      interviewDate: '2023-05-16 11:30 AM', 
      status: 'Scheduled', 
      feedback: 'Pending' 
    },
    { 
      id: 3, 
      candidateName: 'Emma Thompson', 
      position: 'Product Manager', 
      interviewDate: '2023-05-17 2:00 PM', 
      status: 'Completed', 
      feedback: 'Rejected' 
    },
    { 
      id: 4, 
      candidateName: 'Michael Brown', 
      position: 'Data Scientist', 
      interviewDate: '2023-05-18 3:30 PM', 
      status: 'Scheduled', 
      feedback: 'Pending' 
    },
  ];

  const scheduleNewInterview = () => {
    navigate('/interviews/schedule');
    toast.success("Navigating to schedule interview form");
  };

  const viewInterviewDetails = (id: number) => {
    navigate(`/interviews/${id}`);
  };

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
                  {interviews.length > 0 ? (
                    interviews.map((interview) => (
                      <TableRow key={interview.id} className="hover:bg-ats-gray-50 transition-colors cursor-pointer" onClick={() => viewInterviewDetails(interview.id)}>
                        <TableCell className="font-medium">{interview.candidateName}</TableCell>
                        <TableCell>{interview.position}</TableCell>
                        <TableCell>
                          <div className="flex items-center">
                            <Calendar className="w-4 h-4 mr-2 text-ats-gray-400" />
                            {interview.interviewDate}
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
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                        No interviews found.
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
