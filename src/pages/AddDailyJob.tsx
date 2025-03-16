
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import Button from '@/components/ui-custom/Button';
import { ArrowLeft, Save } from 'lucide-react';
import { toast } from 'sonner';

const formSchema = z.object({
  jdNo: z.coerce.number().positive('JD Number must be positive'),
  instructions: z.string().min(5, 'Instructions must be at least 5 characters'),
  assignedUser: z.coerce.number().positive('User ID must be positive'),
});

type FormData = z.infer<typeof formSchema>;

const AddDailyJob = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      jdNo: undefined,
      instructions: '',
      assignedUser: undefined,
    },
  });

  const onSubmit = (data: FormData) => {
    setIsSubmitting(true);
    
    // Simulate API call
    setTimeout(() => {
      try {
        // Get existing data from localStorage
        const existingDailyJobs = JSON.parse(localStorage.getItem('dailyJobs') || '[]');
        
        // Create new daily job object
        const newDailyJob = {
          id: Date.now(), // Use timestamp as ID for simplicity
          jdNo: data.jdNo,
          instructions: data.instructions,
          assignedUser: data.assignedUser,
          assignedDate: 'Today', // For display purposes
        };
        
        // Add to array and save back to localStorage
        existingDailyJobs.push(newDailyJob);
        localStorage.setItem('dailyJobs', JSON.stringify(existingDailyJobs));
        
        toast.success('Daily job assignment added successfully');
        navigate('/daily-jobs');
      } catch (error) {
        console.error('Error saving daily job:', error);
        toast.error('Failed to add daily job assignment');
      } finally {
        setIsSubmitting(false);
      }
    }, 1000);
  };

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-6">
            <button 
              onClick={() => navigate('/daily-jobs')}
              className="flex items-center text-ats-gray-600 hover:text-ats-gray-900"
            >
              <ArrowLeft size={16} className="mr-2" />
              Back to Daily Jobs
            </button>
          </div>
          
          <Card>
            <CardHeader className="p-6">
              <CardTitle>Add Daily Job Assignment</CardTitle>
            </CardHeader>
            <CardContent className="p-6 pt-0">
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                    <FormField
                      control={form.control}
                      name="jdNo"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>JD Number</FormLabel>
                          <FormControl>
                            <Input type="number" placeholder="Enter JD Number" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    
                    <FormField
                      control={form.control}
                      name="assignedUser"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Assigned User ID</FormLabel>
                          <FormControl>
                            <Input type="number" placeholder="Enter user ID" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                  
                  <FormField
                    control={form.control}
                    name="instructions"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Instructions</FormLabel>
                        <FormControl>
                          <Textarea 
                            placeholder="Enter detailed instructions for this task"
                            className="min-h-[120px]"
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  
                  <div className="flex justify-end">
                    <Button
                      type="submit"
                      variant="primary"
                      className="flex items-center gap-2"
                      disabled={isSubmitting}
                    >
                      <Save size={16} />
                      {isSubmitting ? 'Saving...' : 'Save Assignment'}
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

export default AddDailyJob;
