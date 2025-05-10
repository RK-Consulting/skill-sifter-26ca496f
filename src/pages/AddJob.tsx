import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import Button from '@/components/ui-custom/Button';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { jobService } from '@/services/api';

const formSchema = z.object({
  title: z.string().min(3, {
    message: "Job title must be at least 3 characters",
  }),
  department: z.string().min(2, {
    message: "Department must be at least 2 characters",
  }),
  type: z.string().min(2, {
    message: "Job type must be at least 2 characters",
  }),
  isRemote: z.boolean().default(false),
  
  jdNo: z.string().min(1, {
    message: "JD Number is required",
  }),
  clientName: z.string().min(2, {
    message: "Client name must be at least 2 characters",
  }),
  location: z.string().min(2, {
    message: "Location must be at least 2 characters",
  }),
  experience: z.string().min(1, {
    message: "Experience is required",
  }),
  salary: z.string().min(1, {
    message: "Salary is required",
  }),
  
  description: z.string().min(10, {
    message: "Job description must be at least 10 characters",
  }),
  budget: z.string().optional(),
  position: z.string().min(2, {
    message: "Position must be at least 2 characters",
  }),
  language: z.string().optional(),
  certification: z.string().optional(),
  noticePeriod: z.string().optional(),
  requirements: z.string().min(10, {
    message: "Job requirements must be at least 10 characters",
  }),
});

type FormData = z.infer<typeof formSchema>;

const AddJob = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      title: "",
      department: "",
      location: "",
      type: "Full-time",
      description: "",
      requirements: "",
      isRemote: false,
      jdNo: "",
      clientName: "",
      experience: "",
      salary: "",
      budget: "",
      position: "",
      language: "",
      certification: "",
      noticePeriod: "",
    },
  });

  const onSubmit = async (values: FormData) => {
    setIsSubmitting(true);
    console.log("Submitting job:", values);
    
    try {
      // Format job data for API
      const jobData = {
        ...values,
        // Ensure we're using proper field names expected by the API
        status: 'open',
        datePosted: new Date().toISOString()
      };
      
      console.log("Calling API with job data:", jobData);
      
      // Send to API
      const response = await jobService.createJob(jobData);
      
      console.log("API response:", response);
      
      if (response.data?.success) {
        toast.success("Job posted successfully");
        navigate('/jobs');
      } else {
        toast.error(response.data?.message || "Failed to post job");
      }
    } catch (error: any) {
      console.error('Error posting job:', error);
      const errorMessage = error.response?.data?.message || 
                          (error.response?.status === 405 ? "Method not allowed. Please check API configuration." : 
                           "Failed to post job. Server error.");
      toast.error(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Post New Job</h1>
            <p className="text-ats-gray-500">Create a new job posting for your organization.</p>
          </div>

          <Card className="max-w-3xl mx-auto">
            <CardContent className="p-6">
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <FormField
                      control={form.control}
                      name="jdNo"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>JD Number</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter JD Number" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="title"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Job Title</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter job title" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="department"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Department</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter department" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="clientName"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Client Name</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter client name" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="position"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Position</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter position" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="location"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Location</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter location" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="type"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Job Type</FormLabel>
                          <FormControl>
                            <Input placeholder="e.g., Full-time, Part-time, Contract" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="experience"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Experience Required</FormLabel>
                          <FormControl>
                            <Input placeholder="e.g., 3-5 years" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="salary"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Salary Range</FormLabel>
                          <FormControl>
                            <Input placeholder="e.g., $80,000-$100,000" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="budget"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Budget</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter budget" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="language"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Language Requirements</FormLabel>
                          <FormControl>
                            <Input placeholder="e.g., English, Spanish" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="certification"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Certifications Required</FormLabel>
                          <FormControl>
                            <Input placeholder="e.g., PMP, AWS" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="noticePeriod"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Notice Period</FormLabel>
                          <FormControl>
                            <Input placeholder="e.g., 30 days" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                  
                  <FormField
                    control={form.control}
                    name="isRemote"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-start space-x-3 space-y-0">
                        <FormControl>
                          <Checkbox
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                        <div className="space-y-1 leading-none">
                          <FormLabel>
                            This is a remote position
                          </FormLabel>
                        </div>
                      </FormItem>
                    )}
                  />
                  
                  <FormField
                    control={form.control}
                    name="description"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Job Description</FormLabel>
                        <FormControl>
                          <Textarea 
                            placeholder="Enter job description" 
                            className="min-h-[100px]"
                            {...field} 
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  
                  <FormField
                    control={form.control}
                    name="requirements"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Requirements</FormLabel>
                        <FormControl>
                          <Textarea 
                            placeholder="Enter job requirements" 
                            className="min-h-[100px]"
                            {...field} 
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  
                  <div className="flex justify-end gap-4">
                    <Button 
                      type="button" 
                      variant="outline"
                      onClick={() => navigate('/jobs')}
                    >
                      Cancel
                    </Button>
                    <Button 
                      type="submit" 
                      variant="primary"
                      disabled={isSubmitting}
                    >
                      {isSubmitting ? 'Posting...' : 'Post Job'}
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

export default AddJob;
