import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { Calendar, Clock, User, Briefcase, MessageSquare, ArrowLeft, CheckCircle, XCircle } from 'lucide-react';
import Button from '@/components/ui-custom/Button';
import { toast } from "sonner";
import { interviewService } from '@/services/api';
import { Skeleton } from '@/components/ui/skeleton';

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
  
  // Fetch interview data with React Query and better error handling
  const { data: interview, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['interview', id],
    queryFn: async () => {
      try {
        console.log(`Fetching interview details for ID: ${id}`);
        const response = await interviewService.getInterviewById(Number(id));
        console.log('Interview details response:', response.data);
        
        // Check if we got valid data
        if (response.data && response.data.data) {
          return response.data.data;
        } else {
          console.error('Invalid interview data format:', response.data);
          throw new Error('Invalid interview data format');
        }
      } catch (error) {
        console.error('Error fetching interview details:', error);
        toast.error('Failed to load interview details');
        throw error;
      }
    },
    retry: 1,
    retryDelay: 1000,
  });

  const goBack = () => {
    navigate('/interviews');
  };

  const updateStatus = async (status: string) => {
    if (!interview) {
      toast.error('No interview data available to update');
      return;
    }
    
    try {
      toast.loading('Updating interview status...');
      const updatedInterview = { ...interview, status };
      const response = await interviewService.updateInterview(Number(id), updatedInterview);
      
      if (response.data.success) {
        toast.dismiss();
        toast.success(`Interview status updated to ${status}`);
        // Refresh the interview data
        refetch();
      } else {
        toast.dismiss();
        toast.error('Failed to update interview status');
      }
    } catch (error) {
      toast.dismiss();
      console.error('Error updating interview:', error);
      toast.error('Error updating interview status');
    }
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

  // Format date if it exists
  const formatDate = (dateString: string) => {
    if (!dateString) return { date: 'N/A', time: 'N/A' };
    
    try {
      const date = new Date(dateString);
      return { 
        date: date.toLocaleDateString(), 
        time: date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      };
    } catch (e) {
      return { date: dateString, time: '' };
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex flex-col">
        <Navbar />
        <main className="pt-24 pb-10 flex-grow">
          <Container>
            <div className="mb-6 flex items-center">
              <Skeleton className="h-10 w-40 mr-4" />
              <Skeleton className="h-10 w-60" />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
              <Skeleton className="h-80 col-span-1 md:col-span-2" />
              <Skeleton className="h-60" />
            </div>
          </Container>
        </main>
        <Footer />
      </div>
    );
  }

  if (isError || !interview) {
    console.error('Interview details error:', error);
    
    return (
      <div className="min-h-screen bg-background flex flex-col">
        <Navbar />
        <main className="pt-24 pb-10 flex-grow">
          <Container>
            <div className="mb-6 flex items-center">
              <Button 
                variant="ghost" 
                size="sm"
                className="mr-4"
                onClick={goBack}
              >
                <ArrowLeft size={16} className="mr-2" />
                Back to Interviews
              </Button>
              <h1 className="text-3xl font-semibold tracking-tight">Interview Details</h1>
            </div>
            <Card className="mb-8 p-8 text-center">
              <h2 className="text-xl text-red-600 mb-2">Error loading interview details</h2>
              <p className="text-gray-600 mb-4">There was a problem loading the details for this interview.</p>
              <Button variant="primary" onClick={goBack}>Return to Interviews</Button>
            </Card>
          </Container>
        </main>
        <Footer />
      </div>
    );
  }

  const { date: formattedDate, time: formattedTime } = formatDate(interview.interviewDate);

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-6 flex items-center">
            <Button 
              variant="ghost" 
              size="sm"
              className="mr-4"
              onClick={goBack}
            >
              <ArrowLeft size={16} className="mr-2" />
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
                          <p className="text-lg font-medium">{interview.candidateName || 'Not specified'}</p>
                        </div>
                      </div>
                      
                      <div className="flex items-start">
                        <Briefcase className="w-5 h-5 mt-0.5 mr-3 text-ats-gray-500" />
                        <div>
                          <h3 className="text-sm font-medium text-ats-gray-500">Position</h3>
                          <p className="text-lg font-medium">{interview.position || 'Not specified'}</p>
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
                          <p className="text-lg font-medium">{formattedDate}</p>
                        </div>
                      </div>
                      
                      <div className="flex items-start">
                        <Clock className="w-5 h-5 mt-0.5 mr-3 text-ats-gray-500" />
                        <div>
                          <h3 className="text-sm font-medium text-ats-gray-500">Time</h3>
                          <p className="text-lg font-medium">{formattedTime}</p>
                        </div>
                      </div>
                      
                      <div className="flex items-start">
                        <MessageSquare className="w-5 h-5 mt-0.5 mr-3 text-ats-gray-500" />
                        <div>
                          <h3 className="text-sm font-medium text-ats-gray-500">Feedback</h3>
                          <p className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium mt-1 ${getFeedbackColor(interview.feedback)}`}>
                            {interview.feedback || 'No feedback'}
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
                    className="w-full mb-2 flex items-center justify-center gap-2"
                    onClick={() => updateStatus('Completed')}
                  >
                    <CheckCircle size={16} />
                    Mark as Completed
                  </Button>
                  <Button 
                    variant="outline" 
                    className="w-full mb-2 flex items-center justify-center gap-2"
                    onClick={() => updateStatus('Cancelled')}
                  >
                    <XCircle size={16} />
                    Cancel Interview
                  </Button>
                  <Button 
                    variant="secondary" 
                    className="w-full"
                    onClick={() => navigate(`/interviews/schedule?editId=${interview.id}`)}
                  >
                    Reschedule
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default InterviewDetails;