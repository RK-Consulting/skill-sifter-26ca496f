import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
import { clientService } from '@/services/api';
import { getErrorMessage } from '@/lib/utils';

const formSchema = z.object({
  name: z.string().min(2, 'Client name is required'),
  status: z.enum(['prospect', 'active', 'inactive']).default('prospect'),
  contactEmail: z.string().email('Invalid email').optional().or(z.literal('')),
  contactPhone: z.string().optional(),
});

type FormData = z.infer<typeof formSchema>;

const AddClient = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: { name: '', status: 'prospect', contactEmail: '', contactPhone: '' },
  });

  const onSubmit = async (data: FormData) => {
    setIsSubmitting(true);
    try {
      const response = await clientService.createClient(data);
      if (response.data?.success) {
        toast.success('Client added successfully');
        navigate('/clients');
      } else {
        toast.error(response.data?.message || 'Failed to add client');
      }
    } catch (error: unknown) {
      toast.error(getErrorMessage(error, 'Failed to add client'));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <Button variant="ghost" size="sm" className="mb-4 flex gap-2" onClick={() => navigate('/clients')}>
            <ArrowLeft size={16} /> Back to Clients
          </Button>
          <Card className="max-w-2xl mx-auto">
            <CardHeader>
              <CardTitle>Add Client</CardTitle>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                  <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Client Name</FormLabel>
                        <FormControl>
                          <Input placeholder="Acme Corp" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="status"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Status</FormLabel>
                        <FormControl>
                          <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...field}>
                            <option value="prospect">Prospect</option>
                            <option value="active">Active</option>
                            <option value="inactive">Inactive</option>
                          </select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="contactEmail"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Contact Email</FormLabel>
                        <FormControl>
                          <Input type="email" placeholder="contact@acme.com" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="contactPhone"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Contact Phone</FormLabel>
                        <FormControl>
                          <Input placeholder="+1 555 123 4567" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <Button type="submit" variant="primary" className="flex gap-2" disabled={isSubmitting}>
                    <Save size={16} />
                    {isSubmitting ? 'Saving...' : 'Save Client'}
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

export default AddClient;
