
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

// Define the schema for form validation
const formSchema = z.object({
  interviewNumber: z.string().min(1, { message: 'Interview number is required' }),
  jdNo: z.string().min(1, { message: 'JD No. is required' }),
  candidateName: z.string().min(1, { message: 'Candidate name is required' }),
  mobile: z.string().min(10, { message: 'Valid mobile number is required' }),
  email: z.string().email({ message: 'Valid email is required' }),
  status: z.string(),
  interviewDate: z.date({
    required_error: 'Interview date and time is required',
  }),
  feedback: z.string().optional(),
  bill: z.string().optional(),
  followupClient: z.boolean().default(false),
  followupCandidate: z.boolean().default(false),
  joiningDate: z.boolean().default(false),
  confirmationEmail: z.boolean().default(false),
});

type FormValues = z.infer<typeof formSchema>;

const ScheduleInterview = () => {
  const navigate = useNavigate();
  
  // Initialize form with zod resolver
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      status: 'Partner screening round',
      feedback: 'Pending from Client',
      bill: 'No',
      followupClient: false,
      followupCandidate: false,
      joiningDate: false,
      confirmationEmail: false,
    },
  });

  // Handle form submission
  const onSubmit = (data: FormValues) => {
    console.log('Interview scheduled:', data);
    toast.success('Interview successfully scheduled!');
    // In a real application, you would save the data to a database here
    form.reset();
    navigate('/interviews');
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
                    {/* Interview Number */}
                    <FormField
                      control={form.control}
                      name="interviewNumber"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Interview Number</FormLabel>
                          <FormControl>
                            <Input placeholder="INT-0001" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {/* JD No */}
                    <FormField
                      control={form.control}
                      name="jdNo"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>JD No.</FormLabel>
                          <FormControl>
                            <Input placeholder="JD-0001" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

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

                    {/* Mobile Number */}
                    <FormField
                      control={form.control}
                      name="mobile"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Mobile Number</FormLabel>
                          <FormControl>
                            <div className="flex">
                              <Phone className="w-4 h-4 absolute mt-3 ml-3 text-ats-gray-500" />
                              <Input className="pl-10" placeholder="+1 123 456 7890" {...field} />
                            </div>
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {/* Email */}
                    <FormField
                      control={form.control}
                      name="email"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Email ID</FormLabel>
                          <FormControl>
                            <div className="flex">
                              <Mail className="w-4 h-4 absolute mt-3 ml-3 text-ats-gray-500" />
                              <Input className="pl-10" placeholder="john.doe@example.com" {...field} />
                            </div>
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
                              <SelectItem value="Partner screening round">Partner screening round</SelectItem>
                              <SelectItem value="Round 1">Round 1</SelectItem>
                              <SelectItem value="Round 2">Round 2</SelectItem>
                              <SelectItem value="Round 3">Round 3</SelectItem>
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
                          <FormLabel>Candidate Interview Date & Time</FormLabel>
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
                                      // Keep the time from the current value or set it to noon
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

                    {/* Feedback */}
                    <FormField
                      control={form.control}
                      name="feedback"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Feedback</FormLabel>
                          <Select onValueChange={field.onChange} defaultValue={field.value}>
                            <FormControl>
                              <SelectTrigger>
                                <SelectValue placeholder="Select feedback" />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              <SelectItem value="Selected">Selected</SelectItem>
                              <SelectItem value="Failed/Rejected">Failed/Rejected</SelectItem>
                              <SelectItem value="Selected but rejected by Candidate">Selected but rejected by Candidate</SelectItem>
                              <SelectItem value="Pending from Client">Pending from Client</SelectItem>
                              <SelectItem value="Pending from Candidate">Pending from Candidate</SelectItem>
                            </SelectContent>
                          </Select>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {/* Bill */}
                    <FormField
                      control={form.control}
                      name="bill"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Bill</FormLabel>
                          <Select onValueChange={field.onChange} defaultValue={field.value}>
                            <FormControl>
                              <SelectTrigger>
                                <SelectValue placeholder="Select option" />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              <SelectItem value="Yes">Yes</SelectItem>
                              <SelectItem value="No">No</SelectItem>
                            </SelectContent>
                          </Select>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>

                  {/* Completion Checklist */}
                  <div className="mt-6">
                    <h3 className="text-lg font-medium mb-4">Completion</h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <FormField
                        control={form.control}
                        name="followupClient"
                        render={({ field }) => (
                          <FormItem className="flex flex-row items-start space-x-3 space-y-0">
                            <FormControl>
                              <Checkbox
                                checked={field.value}
                                onCheckedChange={field.onChange}
                              />
                            </FormControl>
                            <div className="space-y-1 leading-none">
                              <FormLabel>Follow-up Client for Offer Letter</FormLabel>
                            </div>
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="followupCandidate"
                        render={({ field }) => (
                          <FormItem className="flex flex-row items-start space-x-3 space-y-0">
                            <FormControl>
                              <Checkbox
                                checked={field.value}
                                onCheckedChange={field.onChange}
                              />
                            </FormControl>
                            <div className="space-y-1 leading-none">
                              <FormLabel>Follow-up Candidate for Offer Acceptance</FormLabel>
                            </div>
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="joiningDate"
                        render={({ field }) => (
                          <FormItem className="flex flex-row items-start space-x-3 space-y-0">
                            <FormControl>
                              <Checkbox
                                checked={field.value}
                                onCheckedChange={field.onChange}
                              />
                            </FormControl>
                            <div className="space-y-1 leading-none">
                              <FormLabel>Joining Date Confirmation</FormLabel>
                            </div>
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="confirmationEmail"
                        render={({ field }) => (
                          <FormItem className="flex flex-row items-start space-x-3 space-y-0">
                            <FormControl>
                              <Checkbox
                                checked={field.value}
                                onCheckedChange={field.onChange}
                              />
                            </FormControl>
                            <div className="space-y-1 leading-none">
                              <FormLabel>Confirmation Email from Client for Onboarding</FormLabel>
                            </div>
                          </FormItem>
                        )}
                      />
                    </div>
                  </div>

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
