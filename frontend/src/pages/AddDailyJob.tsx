import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import Button from '@/components/ui-custom/Button';
import { ArrowLeft, Save } from 'lucide-react';
import { toast } from 'sonner';
import { dailyJobService, userService } from '@/services/api';
import { getErrorMessage } from '@/lib/utils';
import { Skeleton } from '@/components/ui/skeleton';

// Updated schema to use string for assignedUser (will be converted to number when submitting)
const formSchema = z.object({
  jdNo: z.coerce.number().positive('JD Number must be positive'),
  instructions: z.string().min(5, 'Instructions must be at least 5 characters'),
  assignedUser: z.string().min(1, 'Please select a user'),
});

type FormData = z.infer<typeof formSchema>;

interface User {
  id: number;
  username: string;
}

const AddDailyJob = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [users, setUsers] = useState<User[]>([]);
  const [isLoadingUsers, setIsLoadingUsers] = useState(true);

  // Fetch company users for the dropdown
  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const response = await userService.getAllUsers();
        if (response.data.success) {
          setUsers(response.data.data);
        } else {
          toast.error('Failed to load users');
        }
      } catch (error) {
        console.error('Error fetching users:', error);
        toast.error('Failed to load users');
      } finally {
        setIsLoadingUsers(false);
      }
    };

    fetchUsers();
  }, []);

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      jdNo: undefined,
      instructions: '',
      assignedUser: '',
    },
  });

  const onSubmit = async (data: FormData) => {
    setIsSubmitting(true);
    console.log("Submitting daily job:", data);
    
    try {
      // Convert assignedUser to number
      const dailyJobData = {
        jdNo: data.jdNo,
        instructions: data.instructions,
        assignedUser: parseInt(data.assignedUser), // Convert string ID to number
        assignedDate: new Date().toISOString()
      };
      
      // Send to API
      const response = await dailyJobService.createDailyJob(dailyJobData);
      console.log("API response:", response);
      
      if (response.data.success) {
        toast.success('Daily job assignment added successfully');
        navigate('/daily-jobs');
      } else {
        toast.error(response.data.message || 'Failed to add daily job assignment');
      }
    } catch (error: unknown) {
      console.error('Error saving daily job:', error);
      toast.error(getErrorMessage(error, 'Failed to add daily job assignment'));
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
                          <FormLabel>Assign To</FormLabel>
                          <FormControl>
                            {isLoadingUsers ? (
                              <Skeleton className="h-10 w-full" />
                            ) : (
                              <Select 
                                onValueChange={field.onChange} 
                                defaultValue={field.value}
                              >
                                <SelectTrigger>
                                  <SelectValue placeholder="Select a user" />
                                </SelectTrigger>
                                <SelectContent>
                                  {users.map((user) => (
                                    <SelectItem key={user.id} value={user.id.toString()}>
                                      {user.username}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              </Select>
                            )}
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
      <Footer />
    </div>
  );
};

export default AddDailyJob;