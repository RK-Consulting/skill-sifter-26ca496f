import React, { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate, useSearchParams } from 'react-router-dom';
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
import { isAxiosError } from 'axios';

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
  const [searchParams] = useSearchParams();
  const editId = searchParams.get('editId');
  const isEditMode = !!editId;
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [isLoadingJob, setIsLoadingJob] = useState(isEditMode);
  const [existingStatus, setExistingStatus] = useState('open');
  
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

  // Edit mode: fetch the existing job and prefill what's actually stored
  // separately (title, department, location, requirements). The other
  // fields (JD No, Client, Experience, etc.) were folded into the
  // description text on create, not stored as their own columns — they
  // can't be reliably parsed back out, so they're left blank for the user
  // to re-enter if relevant. The raw stored description is prefilled as-is
  // into the Description field.
  useEffect(() => {
    if (!editId) return;
    const fetchJob = async () => {
      try {
        const response = await jobService.getJobById(Number(editId));
        const job = response.data?.data;
        if (job) {
          form.reset({
            ...form.getValues(),
            title: job.title || '',
            department: job.department || '',
            location: job.location || '',
            description: job.description || '',
            requirements: job.requirements || '',
          });
          setExistingStatus(job.status || 'open');
        }
      } catch (error) {
        console.error('Error fetching job for edit:', error);
        toast.error('Failed to load job details for editing');
      } finally {
        setIsLoadingJob(false);
      }
    };
    fetchJob();
  }, [editId, form]);

  const onSubmit = async (values: FormData) => {
    setIsSubmitting(true);
    setFormError(null);
    
    console.log("Submitting job:", values);
    
    try {
      // Format job data according to what the backend expects
      const jobData = {
        title: values.title,
        department: values.department,
        location: values.location,
        status: existingStatus, // 'open' for new jobs; preserved as-is when editing
        description: `
Position: ${values.position}
JD Number: ${values.jdNo}
Client: ${values.clientName}
Experience Required: ${values.experience}
Salary Range: ${values.salary}
${values.budget ? `Budget: ${values.budget}` : ''}
${values.language ? `Language Requirements: ${values.language}` : ''}
${values.certification ? `Certifications: ${values.certification}` : ''}
${values.noticePeriod ? `Notice Period: ${values.noticePeriod}` : ''}
${values.isRemote ? 'This is a remote position.' : ''}
Job Type: ${values.type}

Job Description:
${values.description}
`,
        requirements: values.requirements,
      };
      
      console.log("Calling API with job data:", jobData);
      
      // Send to API — update if editing an existing job, create otherwise
      const response = isEditMode
        ? await jobService.updateJob(Number(editId), jobData)
        : await jobService.createJob(jobData);
      
      console.log("API response:", response);
      
      if (response.data?.success) {
        toast.success(isEditMode ? "Job updated successfully" : "Job posted successfully");
        navigate('/jobs');
      } else {
        const errorMsg = response.data?.message || "Failed to post job";
        setFormError(errorMsg);
        toast.error(errorMsg);
      }
    } catch (error: unknown) {
      console.error('Error posting job:', error);
      let errorMessage = "Failed to post job. Server error.";
      if (isAxiosError(error)) {
        const data = error.response?.data as { message?: string } | undefined;
        errorMessage = data?.message ||
          (error.response?.status === 405 ? "Method not allowed. Please check API configuration." :
           errorMessage);
      }
      setFormError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container className="px-4 md:px-6 mx-auto max-w-7xl">
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">{isEditMode ? 'Edit Job' : 'Post New Job'}</h1>
            <p className="text-ats-gray-500">Create a new job posting for your organization.</p>
          </div>

          <Card className="max-w-3xl mx-auto">
            <CardContent className="p-6">
              {formError && (
                <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md text-red-600">
                  <p className="font-medium">Error</p>
                  <p>{formError}</p>
                </div>
              )}
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
                  </div>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
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
                  </div>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
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
                  </div>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
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
                  </div>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
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
                  </div>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
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
                  </div>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
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
                    <FormField
                      control={form.control}
                      name="isRemote"
                      render={({ field }) => (
                        <FormItem className="flex flex-row items-start space-x-3 space-y-0 mt-8">
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
                  </div>
                  
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
                      {isSubmitting ? (isEditMode ? 'Updating...' : 'Posting...') : (isEditMode ? 'Update Job' : 'Post Job')}
                    </Button>
                  </div>
                </form>
              </Form>
            </CardContent>
          </Card>
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default AddJob;