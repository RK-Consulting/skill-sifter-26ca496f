
import React from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
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
import Button from '@/components/ui-custom/Button';
import { Card, CardContent } from '@/components/ui-custom/Card';

const formSchema = z.object({
  name: z.string().min(2, {
    message: "Name must be at least 2 characters",
  }),
  email: z.string().email({
    message: "Please enter a valid email address",
  }),
  phone: z.string().min(10, {
    message: "Please enter a valid phone number",
  }),
  location: z.string().min(2, {
    message: "Location must be at least 2 characters",
  }),
  role: z.string().min(2, {
    message: "Role must be at least 2 characters",
  }),
  experience: z.string().optional(),
  skills: z.string().min(2, {
    message: "Please enter at least one skill",
  }),
  currentCTC: z.string().optional(),
  expectedCTC: z.string().optional(),
  noticePeriod: z.string().optional(),
  clientName: z.string().optional(),
  newJD: z.string().optional(),
});

const AddCandidate = () => {
  const navigate = useNavigate();
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: "",
      email: "",
      phone: "",
      role: "",
      location: "",
      experience: "",
      skills: "",
      currentCTC: "",
      expectedCTC: "",
      noticePeriod: "",
      clientName: "",
      newJD: "",
    },
  });

  const onSubmit = (values: z.infer<typeof formSchema>) => {
    // In a real application, you would send this to your backend
    console.log(values);
    
    // Get existing candidates from localStorage or initialize empty array
    const existingCandidates = JSON.parse(localStorage.getItem('candidates') || '[]');
    
    // Add new candidate with ID
    const newCandidate = {
      id: existingCandidates.length + 1,
      ...values,
      status: 'Screening',
      date: new Date().toLocaleDateString(),
      resumePath: "" // Added for the resumePath field from the struct
    };
    
    // Save updated candidates list
    localStorage.setItem('candidates', JSON.stringify([...existingCandidates, newCandidate]));
    
    toast.success("Candidate added successfully");
    navigate('/candidates');
  };

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Add New Candidate</h1>
            <p className="text-ats-gray-500">Enter candidate details to add them to your talent pool.</p>
          </div>

          <Card className="max-w-3xl mx-auto">
            <CardContent className="p-6">
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <FormField
                      control={form.control}
                      name="name"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Full Name</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter candidate's full name" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="email"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Email</FormLabel>
                          <FormControl>
                            <Input type="email" placeholder="Enter email address" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="phone"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Phone Number</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter phone number" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="role"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Role/Position</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter role or position" {...field} />
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
                      name="experience"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Experience</FormLabel>
                          <FormControl>
                            <Input placeholder="Years of experience" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    {/* New fields based on Candidate struct */}
                    <FormField
                      control={form.control}
                      name="currentCTC"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Current CTC</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter current CTC" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="expectedCTC"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Expected CTC</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter expected CTC" {...field} />
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
                            <Input placeholder="Enter notice period (in days)" {...field} />
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
                  <FormField
                    control={form.control}
                    name="skills"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Skills</FormLabel>
                        <FormControl>
                          <Textarea 
                            placeholder="Enter candidate's skills (separated by commas)" 
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
                    name="newJD"
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
                  <div className="flex justify-end gap-4">
                    <Button 
                      type="button" 
                      variant="outline"
                      onClick={() => navigate('/candidates')}
                    >
                      Cancel
                    </Button>
                    <Button type="submit" variant="primary">
                      Add Candidate
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

export default AddCandidate;
