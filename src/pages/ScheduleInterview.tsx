
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { toast } from 'sonner';
import {
  Calendar as CalendarIcon,
  Check,
  Clock,
  Mail,
  Phone,
  User,
  ArrowLeft
} from 'lucide-react';

import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Calendar } from '@/components/ui/calendar';
import { format } from 'date-fns';
import { cn } from '@/lib/utils';
import { interviewService } from '@/services/api';

// Define the schema for form validation matching backend Interview model
const formSchema = z.object({
  candidateName: z.string().min(1, { message: 'Candidate name is required' }),
  position: z.string().min(1, { message: 'Position is required' }),
  interviewDate: z.date({
    required_error: 'Interview date and time is required',
  }),
  status: z.string().default('scheduled'),
  feedback: z.string().optional(),
});

type FormValues = z.infer<typeof formSchema>;

const ScheduleInterview = () => {
  const navigate = useNavigate();
  
  // Initialize form with zod resolver
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      status: 'scheduled',
      feedback: '',
    },
  });

  // Handle form submission
  const onSubmit = async (data: FormValues) => {
    try {
      console.log('Interview data to submit:', data);
      
      // Create interview object matching backend Interview model
      const interviewData = {
        candidateName: data.candidateName,
        position: data.position,
        interviewDate: data.interviewDate.toISOString(),
        status: data.status,
        feedback: data.feedback || '',
      };

      console.log('Sending interview data:', interviewData);
      
      const response = await interviewService.createInterview(interviewData);
      console.log('Interview created successfully:', response);
      
      toast.success('Interview successfully scheduled!');
      form.reset();
      navigate('/interviews');
    } catch (error) {
      console.error('Error scheduling interview:', error);
      toast.error('Failed to schedule interview. Please try again.');
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-8">
            <div className="flex items-center gap-4 mb-6">
              <Button 
                variant="ghost" 
                size="sm" 
                className="flex items-center gap-2"
                onClick={() => navigate('/interviews')}
              >
                <ArrowLeft size={16} />
                Back to Interviews
              </Button>
            </div>
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Schedule Interview</h1>
            <p className="text-ats-gray-500">Create a new interview for a candidate.</p>
          </div>

          <Card className="mb-8">
            <CardContent className="p-6">
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {/* Candidate Name */}
                    <FormField
                      control={form.control}
                      name="candidateName"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Candidate Name</FormLabel>
                          <FormControl>
                            <div className="flex">
                              <User className="w-4 h-4 absolute mt-3 ml-3 text-ats-gray-500" />
                              <Input className="pl-10" placeholder="John Doe" {...field} />
                            </div>
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {/* Position */}
                    <FormField
                      control={form.control}
                      name="position"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Position</FormLabel>
                          <FormControl>
                            <Input placeholder="Software Engineer" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {/* Status */}
                    <FormField
                      control={form.control}
                      name="status"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Status</FormLabel>
                          <Select onValueChange={field.onChange} defaultValue={field.value}>
                            <FormControl>
                              <SelectTrigger>
                                <SelectValue placeholder="Select status" />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              <SelectItem value="scheduled">Scheduled</SelectItem>
                              <SelectItem value="completed">Completed</SelectItem>
                              <SelectItem value="cancelled">Cancelled</SelectItem>
                              <SelectItem value="rescheduled">Rescheduled</SelectItem>
                            </SelectContent>
                          </Select>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {/* Interview Date & Time */}
                    <FormField
                      control={form.control}
                      name="interviewDate"
                      render={({ field }) => (
                        <FormItem className="flex flex-col">
                          <FormLabel>Interview Date & Time</FormLabel>
                          <Popover>
                            <PopoverTrigger asChild>
                              <FormControl>
                                <Button
                                  variant={"outline"}
                                  className={cn(
                                    "w-full pl-3 text-left font-normal flex justify-between",
                                    !field.value && "text-muted-foreground"
                                  )}
                                >
                                  <div className="flex items-center">
                                    <CalendarIcon className="mr-2 h-4 w-4" />
                                    <Clock className="mr-2 h-4 w-4" />
                                    {field.value ? (
                                      format(field.value, "PPP p")
                                    ) : (
                                      <span>Select date and time</span>
                                    )}
                                  </div>
                                  <CalendarIcon className="h-4 w-4 opacity-50" />
                                </Button>
                              </FormControl>
                            </PopoverTrigger>
                            <PopoverContent className="w-auto p-0" align="start">
                              <div className="p-3">
                                <Calendar
                                  mode="single"
                                  selected={field.value}
                                  onSelect={(date) => {
                                    if (date) {
                                      const currentTime = field.value ? field.value : new Date();
                                      date.setHours(currentTime.getHours());
                                      date.setMinutes(currentTime.getMinutes());
                                      field.onChange(date);
                                    }
                                  }}
                                  initialFocus
                                  className="pointer-events-auto"
                                />
                                
                                <div className="px-3 pt-3 pb-1 border-t border-border">
                                  <div className="flex items-center">
                                    <Clock className="mr-2 h-4 w-4" />
                                    <label className="text-sm font-medium mr-3">Time:</label>
                                    <input
                                      type="time"
                                      className="border rounded px-2 py-1 text-sm"
                                      onChange={(e) => {
                                        const [hours, minutes] = e.target.value.split(':').map(Number);
                                        const newDate = field.value || new Date();
                                        newDate.setHours(hours);
                                        newDate.setMinutes(minutes);
                                        field.onChange(newDate);
                                      }}
                                      value={field.value ? `${field.value.getHours().toString().padStart(2, '0')}:${field.value.getMinutes().toString().padStart(2, '0')}` : ''}
                                    />
                                  </div>
                                </div>
                              </div>
                            </PopoverContent>
                          </Popover>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>

                  {/* Feedback */}
                  <FormField
                    control={form.control}
                    name="feedback"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Feedback (Optional)</FormLabel>
                        <FormControl>
                          <Input placeholder="Enter feedback..." {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <div className="flex justify-end gap-3">
                    <Button 
                      type="button" 
                      variant="outline"
                      onClick={() => navigate('/interviews')}
                    >
                      Cancel
                    </Button>
                    <Button type="submit">
                      Schedule Interview
                    </Button>
                  </div>
                </form>
              </Form>
            </CardContent>
          </Card>
        </Container>
      </main>
    </div>
  );
};

export default ScheduleInterview;
