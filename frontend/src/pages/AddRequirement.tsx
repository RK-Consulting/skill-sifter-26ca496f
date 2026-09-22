import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import Button from '@/components/ui-custom/Button';
import { ArrowLeft, Save } from 'lucide-react';
import { toast } from 'sonner';
import { clientService, requirementService } from '@/services/api';
import { getErrorMessage } from '@/lib/utils';

interface Client {
  id: number;
  name: string;
}

const formSchema = z.object({
  jobId: z.string().min(1, 'Job ID is required'),
  clientId: z.coerce.number().min(1, 'Client is required'),
  jobType: z.enum(['fulltime', 'contract']),
  title: z.string().min(2, 'Job title is required'),
  department: z.string().optional(),
  experienceRequired: z.string().optional(),
  budget: z.string().optional(),
  languageRequirements: z.string().optional(),
  certificationsRequired: z.string().optional(),
  noticePeriod: z.string().optional(),
  workArrangement: z.enum(['hybrid', 'remote', 'office']),
  mandatoryRequirements: z.string().optional(),
  description: z.string().optional(),
  status: z.enum(['open', 'closed', 'on_hold', 'cancelled']),
  location: z.string().optional(),
  headcount: z.coerce.number().min(1, 'Number of open positions must be at least 1'),
});

type FormData = z.infer<typeof formSchema>;

const AddRequirement = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { data: clientsData } = useQuery({
    queryKey: ['clients-for-requirement-form'],
    queryFn: async () => {
      const response = await clientService.getAllClients({ page: 1, limit: 100, status: 'active' });
      return (response.data?.data ?? []) as Client[];
    },
  });
  const clients = clientsData ?? [];

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      jobId: '',
      clientId: 0,
      jobType: 'fulltime',
      title: '',
      department: '',
      experienceRequired: '',
      budget: '',
      languageRequirements: '',
      certificationsRequired: '',
      noticePeriod: '',
      workArrangement: 'hybrid',
      mandatoryRequirements: '',
      description: '',
      status: 'open',
      location: '',
      headcount: 1,
    },
  });

  const onSubmit = async (data: FormData) => {
    setIsSubmitting(true);
    try {
      const response = await requirementService.createRequirement(data);
      if (response.data?.success) {
        toast.success('Requirement added successfully');
        navigate('/requirements');
      } else {
        toast.error(response.data?.message || 'Failed to add requirement');
      }
    } catch (error: unknown) {
      toast.error(getErrorMessage(error, 'Failed to add requirement'));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <Button variant="ghost" size="sm" className="mb-4 flex gap-2" onClick={() => navigate('/requirements')}>
            <ArrowLeft size={16} /> Back to Requirements
          </Button>

          <Card className="max-w-4xl mx-auto">
            <CardHeader>
              <CardTitle>Add Requirement</CardTitle>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <FormField control={form.control} name="jobId" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Job ID *</FormLabel>
                        <FormControl><Input placeholder="e.g. RK-2026-001" {...field} /></FormControl>
                        <FormMessage />
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="clientId" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Client *</FormLabel>
                        <FormControl>
                          <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...field} onChange={(e) => field.onChange(Number(e.target.value))}>
                            <option value={0} disabled>Select a client...</option>
                            {clients.map((client) => <option key={client.id} value={client.id}>{client.name}</option>)}
                          </select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="jobType" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Job Type *</FormLabel>
                        <FormControl>
                          <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...field}>
                            <option value="fulltime">Full-time</option>
                            <option value="contract">Contract</option>
                          </select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="title" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Job Title *</FormLabel>
                        <FormControl><Input placeholder="e.g. Senior Software Engineer" {...field} /></FormControl>
                        <FormMessage />
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="department" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Department</FormLabel>
                        <FormControl><Input placeholder="e.g. Engineering" {...field} /></FormControl>
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="experienceRequired" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Experience Required</FormLabel>
                        <FormControl><Input placeholder="e.g. 5-8 years" {...field} /></FormControl>
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="budget" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Budget</FormLabel>
                        <FormControl><Input placeholder="e.g. ₹18-24 LPA" {...field} /></FormControl>
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="languageRequirements" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Language Requirements</FormLabel>
                        <FormControl><Input placeholder="e.g. English - C1, German - B2" {...field} /></FormControl>
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="certificationsRequired" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Certifications Required</FormLabel>
                        <FormControl><Input placeholder="e.g. AWS, PMP, CFA" {...field} /></FormControl>
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="noticePeriod" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Notice Period</FormLabel>
                        <FormControl><Input placeholder="e.g. Immediate, 30 days, 60 days" {...field} /></FormControl>
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="workArrangement" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Mode of Work *</FormLabel>
                        <FormControl>
                          <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...field}>
                            <option value="hybrid">Hybrid</option>
                            <option value="remote">Remote</option>
                            <option value="office">Office</option>
                          </select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="status" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Status *</FormLabel>
                        <FormControl>
                          <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...field}>
                            <option value="open">Open</option>
                            <option value="closed">Closed</option>
                            <option value="on_hold">On Hold</option>
                            <option value="cancelled">Cancelled</option>
                          </select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="location" render={({ field }) => (
                      <FormItem>
                        <FormLabel>Job Location</FormLabel>
                        <FormControl><Input placeholder="e.g. Bengaluru, Karnataka" {...field} /></FormControl>
                      </FormItem>
                    )} />

                    <FormField control={form.control} name="headcount" render={({ field }) => (
                      <FormItem>
                        <FormLabel>No. of Open Positions *</FormLabel>
                        <FormControl><Input type="number" min={1} {...field} onChange={(e) => field.onChange(Number(e.target.value))} /></FormControl>
                        <FormMessage />
                      </FormItem>
                    )} />
                  </div>

                  <FormField control={form.control} name="mandatoryRequirements" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Mandatory Requirements</FormLabel>
                      <FormControl><textarea className="min-h-28 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="List mandatory skills, qualifications, technologies, or domain requirements..." {...field} /></FormControl>
                    </FormItem>
                  )} />

                  <FormField control={form.control} name="description" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Job Description</FormLabel>
                      <FormControl><textarea className="min-h-40 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Enter the complete job description..." {...field} /></FormControl>
                    </FormItem>
                  )} />

                  <div className="flex justify-end gap-3 pt-2">
                    <Button type="button" variant="outline" onClick={() => navigate('/requirements')}>Cancel</Button>
                    <Button type="submit" variant="primary" disabled={isSubmitting}>
                      <Save size={16} className="mr-2" />
                      {isSubmitting ? 'Saving...' : 'Save Requirement'}
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

export default AddRequirement;
