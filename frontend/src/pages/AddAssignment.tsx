import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import Button from '@/components/ui-custom/Button';
import { ArrowLeft, Save } from 'lucide-react';
import { toast } from 'sonner';
import { candidateService } from '@/services/api';
import { requirementService, assignmentV1Service } from '@/services/api';
import { getErrorMessage } from '@/lib/utils';

interface Candidate {
  id: number;
  name: string;
  email: string;
}

interface Requirement {
  id: number;
  title: string;
  status: string;
}

const formSchema = z.object({
  candidateId: z.coerce.number().min(1, 'Candidate is required'),
  requirementId: z.coerce.number().min(1, 'Requirement is required'),
});

type FormData = z.infer<typeof formSchema>;

const AddAssignment = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { data: candidatesData } = useQuery({
    queryKey: ['candidates-for-assignment-form'],
    queryFn: async () => {
      const response = await candidateService.getAllCandidates();
      return (response.data?.data ?? []) as Candidate[];
    },
  });
  const candidates = candidatesData ?? [];

  const { data: requirementsData } = useQuery({
    queryKey: ['requirements-for-assignment-form'],
    queryFn: async () => {
      const response = await requirementService.getAllRequirements();
      return (response.data?.data ?? []) as Requirement[];
    },
  });
  // Only open requirements make sense to assign a candidate against.
  const requirements = (requirementsData ?? []).filter((r) => r.status === 'open');

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: { candidateId: 0, requirementId: 0 },
  });

  const onSubmit = async (data: FormData) => {
    setIsSubmitting(true);
    try {
      const response = await assignmentV1Service.createAssignment(data);
      if (response.data?.success) {
        toast.success('Assignment created successfully');
        navigate('/assignments');
      } else {
        toast.error(response.data?.message || 'Failed to create assignment');
      }
    } catch (error: unknown) {
      toast.error(getErrorMessage(error, 'Failed to create assignment'));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <Button variant="ghost" size="sm" className="mb-4 flex gap-2" onClick={() => navigate('/assignments')}>
            <ArrowLeft size={16} /> Back to Assignments
          </Button>
          <Card className="max-w-2xl mx-auto">
            <CardHeader>
              <CardTitle>New Recruitment Assignment</CardTitle>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                  <FormField
                    control={form.control}
                    name="candidateId"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Candidate</FormLabel>
                        <FormControl>
                          <select
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            {...field}
                            onChange={(e) => field.onChange(Number(e.target.value))}
                          >
                            <option value={0} disabled>
                              Select a candidate...
                            </option>
                            {candidates.map((c) => (
                              <option key={c.id} value={c.id}>
                                {c.name} ({c.email})
                              </option>
                            ))}
                          </select>
                        </FormControl>
                        {candidates.length === 0 && (
                          <p className="text-xs text-ats-gray-500">No candidates yet — add one first.</p>
                        )}
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="requirementId"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Requirement</FormLabel>
                        <FormControl>
                          <select
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            {...field}
                            onChange={(e) => field.onChange(Number(e.target.value))}
                          >
                            <option value={0} disabled>
                              Select an open requirement...
                            </option>
                            {requirements.map((r) => (
                              <option key={r.id} value={r.id}>
                                {r.title}
                              </option>
                            ))}
                          </select>
                        </FormControl>
                        {requirements.length === 0 && (
                          <p className="text-xs text-ats-gray-500">No open requirements available.</p>
                        )}
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <Button type="submit" variant="primary" className="flex gap-2" disabled={isSubmitting}>
                    <Save size={16} />
                    {isSubmitting ? 'Saving...' : 'Create Assignment'}
                  </Button>
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

export default AddAssignment;
