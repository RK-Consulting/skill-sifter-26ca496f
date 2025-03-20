
import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { Calendar, Clock, User, Briefcase, MessageSquare, ArrowLeft, CheckCircle, XCircle } from 'lucide-react';
import Button from '@/components/ui-custom/Button';
import { toast } from "sonner";

interface Interview {
  id: number;
  candidateName: string;
  position: string;
  interviewDate: string;
  status: string;
  feedback: string;
  notes?: string;
  interviewer?: string;
}

const InterviewDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  
  // Sample data - in a real app, this would be fetched from a database
  const interview: Interview = {
    id: Number(id),
    candidateName: 'Sarah Wilson',
    position: 'Senior UI Designer',
    interviewDate: '2023-05-15 10:00 AM',
    status: 'Completed',
    feedback: 'Selected',
    notes: 'Sarah demonstrated excellent UI design skills and a solid understanding of user experience principles. Her portfolio showcased a range of impressive projects, and she communicated her design decisions clearly.',
    interviewer: 'Alex Johnson'
  };

  const goBack = () => {
    navigate('/interviews');
  };

  const updateStatus = (status: string) => {
    // In a real app, this would update the database
    toast.success(`Interview status updated to ${status}`);
  };

  const getStatusColor = (status: string) => {
    switch(status) {
      case 'Completed': return 'bg-green-100 text-green-800';
      case 'Scheduled': return 'bg-blue-100 text-blue-800';
      case 'Cancelled': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getFeedbackColor = (feedback: string) => {
    switch(feedback) {
      case 'Selected': return 'bg-green-100 text-green-800';
      case 'Rejected': return 'bg-red-100 text-red-800';
      case 'Pending': return 'bg-yellow-100 text-yellow-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-6 flex items-center">
            <Button 
              variant="ghost" 
              size="sm"
              className="mr-4"
              onClick={goBack}
              icon={<ArrowLeft size={16} />}
            >
              Back to Interviews
            </Button>
            <h1 className="text-3xl font-semibold tracking-tight">Interview Details</h1>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <Card className="col-span-1 md:col-span-2">
              <CardHeader className="p-6">
                <CardTitle className="flex items-center justify-between">
                  <span>Interview Information</span>
                  <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-sm font-medium ${getStatusColor(interview.status)}`}>
                    {interview.status}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent className="p-6 pt-0">
                <div className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-4">
                      <div className="flex items-start">
                        <User className="w-5 h-5 mt-0.5 mr-3 text-ats-gray-500" />
                        <div>
                          <h3 className="text-sm font-medium text-ats-gray-500">Candidate</h3>
                          <p className="text-lg font-medium">{interview.candidateName}</p>
                        </div>
                      </div>
                      
                      <div className="flex items-start">
                        <Briefcase className="w-5 h-5 mt-0.5 mr-3 text-ats-gray-500" />
                        <div>
                          <h3 className="text-sm font-medium text-ats-gray-500">Position</h3>
                          <p className="text-lg font-medium">{interview.position}</p>
                        </div>
                      </div>
                      
                      <div className="flex items-start">
                        <User className="w-5 h-5 mt-0.5 mr-3 text-ats-gray-500" />
                        <div>
                          <h3 className="text-sm font-medium text-ats-gray-500">Interviewer</h3>
                          <p className="text-lg font-medium">{interview.interviewer || 'Not assigned'}</p>
                        </div>
                      </div>
                    </div>
                    
                    <div className="space-y-4">
                      <div className="flex items-start">
                        <Calendar className="w-5 h-5 mt-0.5 mr-3 text-ats-gray-500" />
                        <div>
                          <h3 className="text-sm font-medium text-ats-gray-500">Date</h3>
                          <p className="text-lg font-medium">{interview.interviewDate.split(' ')[0]}</p>
                        </div>
                      </div>
                      
                      <div className="flex items-start">
                        <Clock className="w-5 h-5 mt-0.5 mr-3 text-ats-gray-500" />
                        <div>
                          <h3 className="text-sm font-medium text-ats-gray-500">Time</h3>
                          <p className="text-lg font-medium">{interview.interviewDate.split(' ')[1]} {interview.interviewDate.split(' ')[2]}</p>
                        </div>
                      </div>
                      
                      <div className="flex items-start">
                        <MessageSquare className="w-5 h-5 mt-0.5 mr-3 text-ats-gray-500" />
                        <div>
                          <h3 className="text-sm font-medium text-ats-gray-500">Feedback</h3>
                          <p className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium mt-1 ${getFeedbackColor(interview.feedback)}`}>
                            {interview.feedback}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                  
                  <div className="pt-4 border-t border-ats-gray-200">
                    <h3 className="text-sm font-medium text-ats-gray-500 mb-2">Interview Notes</h3>
                    <p className="text-ats-gray-700">{interview.notes || 'No notes available.'}</p>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="p-6">
                <CardTitle>Actions</CardTitle>
              </CardHeader>
              <CardContent className="p-6 pt-0">
                <div className="space-y-4">
                  <Button 
                    variant="primary" 
                    className="w-full mb-2"
                    icon={<CheckCircle size={16} />}
                    onClick={() => updateStatus('Completed')}
                  >
                    Mark as Completed
                  </Button>
                  <Button 
                    variant="outline" 
                    className="w-full mb-2"
                    icon={<XCircle size={16} />}
                    onClick={() => updateStatus('Cancelled')}
                  >
                    Cancel Interview
                  </Button>
                  <Button 
                    variant="secondary" 
                    className="w-full"
                    onClick={() => navigate('/interviews/schedule')}
                  >
                    Reschedule
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </Container>
      </main>
    </div>
  );
};

export default InterviewDetails;
