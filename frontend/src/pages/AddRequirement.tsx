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
  clientId: z.coerce.number().min(1, 'Client is required'),
  title: z.string().min(2, 'Title is required'),
  department: z.string().optional(),
  location: z.string().optional(),
  status: z.enum(['draft', 'open', 'on_hold', 'filled', 'cancelled']).default('draft'),
  headcount: z.coerce.number().min(1, 'Headcount must be at least 1').default(1),
  description: z.string().optional(),
  requiredSkills: z.string().optional(),
});

type FormData = z.infer<typeof formSchema>;

const AddRequirement = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Requirements need a client to belong to; fetch a reasonably large
  // page of active/prospect clients for the dropdown. This form-level
  // fetch does not attempt to paginate — a company with more clients
  // than one page can hold is a scale point worth its own UI later.
  const { data: clientsData } = useQuery({
    queryKey: ['clients-for-requirement-form'],
    queryFn: async () => {
      const response = await clientService.getAllClients({ page: 1, limit: 100 });
      return (response.data?.data ?? []) as Client[];
    },
  });
  const clients = clientsData ?? [];

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      clientId: 0,
      title: '',
      department: '',
      location: '',
      status: 'draft',
      headcount: 1,
      description: '',
      requiredSkills: '',
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
          <Card className="max-w-2xl mx-auto">
            <CardHeader>
              <CardTitle>Add Requirement</CardTitle>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                  <FormField
                    control={form.control}
                    name="clientId"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Client</FormLabel>
                        <FormControl>
                          <select
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            {...field}
                            onChange={(e) => field.onChange(Number(e.target.value))}
                          >
                            <option value={0} disabled>
                              Select a client...
                            </option>
                            {clients.map((c) => (
                              <option key={c.id} value={c.id}>
                                {c.name}
                              </option>
                            ))}
                          </select>
                        </FormControl>
                        {clients.length === 0 && (
                          <p className="text-xs text-ats-gray-500">No clients yet — add one first.</p>
                        )}
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="title"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Title</FormLabel>
                        <FormControl>
                          <Input placeholder="Senior Backend Engineer" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <div className="grid grid-cols-2 gap-4">
                    <FormField
                      control={form.control}
                      name="department"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Department</FormLabel>
                          <FormControl>
                            <Input placeholder="Engineering" {...field} />
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
                            <Input placeholder="Bengaluru" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <FormField
                      control={form.control}
                      name="status"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Status</FormLabel>
                          <FormControl>
                            <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...field}>
                              <option value="draft">Draft</option>
                              <option value="open">Open</option>
                              <option value="on_hold">On Hold</option>
                              <option value="filled">Filled</option>
                              <option value="cancelled">Cancelled</option>
                            </select>
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="headcount"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Headcount</FormLabel>
                          <FormControl>
                            <Input type="number" min={1} {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                  <FormField
                    control={form.control}
                    name="requiredSkills"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Required Skills</FormLabel>
                        <FormControl>
                          <Input placeholder="Go, PostgreSQL, React" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="description"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Description</FormLabel>
                        <FormControl>
                          <textarea
                            className="flex min-h-[100px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            placeholder="Role description..."
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <Button type="submit" variant="primary" className="flex gap-2" disabled={isSubmitting}>
                    <Save size={16} />
                    {isSubmitting ? 'Saving...' : 'Save Requirement'}
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

export default AddRequirement;
