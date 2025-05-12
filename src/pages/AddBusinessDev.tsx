
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
import { businessDevService } from '@/services/api';

const formSchema = z.object({
  clientName: z.string().min(2, 'Client name is required'),
  partnerName: z.string().optional(),
  contactPerson: z.string().min(2, 'Contact person is required'),
  contactNumber: z.string().min(5, 'Valid contact number is required'),
  contactEmail: z.string().email('Invalid email address'),
});

type FormData = z.infer<typeof formSchema>;

const AddBusinessDev = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      clientName: '',
      partnerName: '',
      contactPerson: '',
      contactNumber: '',
      contactEmail: '',
    },
  });

  const onSubmit = async (data: FormData) => {
    setIsSubmitting(true);
    
    try {
      // Get token from localStorage
      const token = localStorage.getItem('token');
      
      if (!token) {
        toast.error('Authentication required. Please log in.');
        navigate('/login');
        return;
      }
      
      // Send data to API using our business service
      const response = await businessDevService.createBusinessDev(data);
      
      if (response.data && response.data.success) {
        toast.success('Business contact added successfully');
        navigate('/business-dev');
      } else {
        toast.error(response.data?.message || 'Failed to add business contact');
      }
    } catch (error: any) {
      console.error('Error saving business contact:', error);
      toast.error(error.response?.data?.message || 'Failed to add business contact');
      
      // Fallback to localStorage for demo purposes if API fails
      try {
        // Get existing data from localStorage
        const existingBusinessDevs = JSON.parse(localStorage.getItem('businessDevs') || '[]');
        
        // Create new business dev object
        const newBusinessDev = {
          id: Date.now(), // Use timestamp as ID for simplicity
          clientName: data.clientName,
          partnerName: data.partnerName || '',
          contactPerson: data.contactPerson,
          contactNumber: data.contactNumber,
          contactEmail: data.contactEmail,
          createdAt: 'Today', // For display purposes
        };
        
        // Add to array and save back to localStorage
        existingBusinessDevs.push(newBusinessDev);
        localStorage.setItem('businessDevs', JSON.stringify(existingBusinessDevs));
        
        toast.success('Business contact saved locally (API connection failed)');
        navigate('/business-dev');
      } catch (localError) {
        console.error('Error saving to localStorage:', localError);
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-6">
            <button 
              onClick={() => navigate('/business-dev')}
              className="flex items-center text-ats-gray-600 hover:text-ats-gray-900"
            >
              <ArrowLeft size={16} className="mr-2" />
              Back to Business Development
            </button>
          </div>
          
          <Card>
            <CardHeader className="p-6">
              <CardTitle>Add Business Contact</CardTitle>
            </CardHeader>
            <CardContent className="p-6 pt-0">
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
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
                      name="partnerName"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Partner Name (Optional)</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter partner name if applicable" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                  
                  <FormField
                    control={form.control}
                    name="contactPerson"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Contact Person</FormLabel>
                        <FormControl>
                          <Input placeholder="Enter contact person name" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  
                  <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                    <FormField
                      control={form.control}
                      name="contactNumber"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Contact Number</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter contact phone number" {...field} />
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
                            <Input type="email" placeholder="Enter email address" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                  
                  <div className="flex justify-end">
                    <Button
                      type="submit"
                      variant="primary"
                      className="flex items-center gap-2"
                      disabled={isSubmitting}
                    >
                      <Save size={16} />
                      {isSubmitting ? 'Saving...' : 'Save Contact'}
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

export default AddBusinessDev;
