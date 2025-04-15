
import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { authService } from '@/services/api';

const formSchema = z.object({
  username: z.string().min(3, {
    message: "Username must be at least 3 characters",
  }),
  email: z.string().email({
    message: "Please enter a valid email address",
  }),
  password: z.string().min(6, {
    message: "Password must be at least 6 characters",
  }),
  confirmPassword: z.string().min(6, {
    message: "Password must be at least 6 characters",
  }),
  company: z.string().min(2, {
    message: "Company name must be at least 2 characters",
  }),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords do not match",
  path: ["confirmPassword"],
});

const Register = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [apiError, setApiError] = useState('');
  
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      email: "",
      password: "",
      confirmPassword: "",
      company: "",
    },
  });

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    try {
      setIsSubmitting(true);
      setApiError('');
      
      // Remove confirmPassword as it's not needed for the API
      const { confirmPassword, ...userData } = values;

      console.log("Attempting to register with:", { ...userData, password: "***" });
      console.log("API URL:", import.meta.env.VITE_API_URL || 'http://localhost:8080/api');
      
      const response = await authService.register(userData);
      
      if (response.data.success) {
        toast.success("Registration successful. Please login.");
        navigate('/login');
      } else {
        toast.error(response.data.message || "Registration failed");
        setApiError(response.data.message || "Registration failed");
      }
    } catch (error: any) {
      console.error("Registration error:", error);
      
      // Handle different types of errors
      if (error.code === 'ERR_NETWORK') {
        const errorMessage = "Cannot connect to server. Please make sure the backend is running.";
        toast.error(errorMessage);
        setApiError(errorMessage);
      } else {
        const errorMessage = error.response?.data?.message || "Registration failed. Please try again.";
        toast.error(errorMessage);
        setApiError(errorMessage);
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <div className="flex-grow flex items-center justify-center">
        <Container className="max-w-md">
          <div className="text-center mb-8">
            <img 
              src="/lovable-uploads/35d9a32a-9b4d-4be7-a93d-03a036a4ab8a.png" 
              alt="R K Consulting Logo" 
              className="h-16 w-16 rounded-full mx-auto mb-2"
            />
            <h1 className="text-2xl font-bold">R K Consulting</h1>
            <p className="text-ats-blue-500 font-medium">SkillSifter ATS</p>
          </div>
          
          <Card>
            <CardHeader>
              <CardTitle className="text-2xl text-center">Create an Account</CardTitle>
            </CardHeader>
            <CardContent>
              {apiError && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
                  <p><strong>Error:</strong> {apiError}</p>
                </div>
              )}

              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                  <FormField
                    control={form.control}
                    name="username"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Username</FormLabel>
                        <FormControl>
                          <Input placeholder="Enter a username" {...field} />
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
                          <Input type="email" placeholder="Enter your email" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="company"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Company Name</FormLabel>
                        <FormControl>
                          <Input placeholder="Enter your company name" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="password"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Password</FormLabel>
                        <FormControl>
                          <Input type="password" placeholder="Create a password" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="confirmPassword"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Confirm Password</FormLabel>
                        <FormControl>
                          <Input type="password" placeholder="Confirm your password" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <div className="flex flex-col space-y-2 pt-2">
                    <Button 
                      type="submit" 
                      variant="primary" 
                      className="w-full"
                      disabled={isSubmitting}
                    >
                      {isSubmitting ? 'Registering...' : 'Register'}
                    </Button>
                    <div className="text-center text-sm mt-4">
                      Already have an account?{" "}
                      <span 
                        className="text-ats-blue cursor-pointer hover:underline"
                        onClick={() => navigate('/login')}
                      >
                        Login
                      </span>
                    </div>
                  </div>
                </form>
              </Form>
            </CardContent>
          </Card>

          {/* Debug information in development mode */}
          {import.meta.env.DEV && (
            <div className="mt-8 p-3 bg-blue-50 border border-blue-200 rounded-md">
              <p className="text-sm text-gray-700 mb-1"><strong>Debug Info:</strong></p>
              <p className="text-sm text-gray-700">API URL: {import.meta.env.VITE_API_URL || 'http://localhost:8080/api'}</p>
              <p className="text-sm text-gray-700">Environment: {import.meta.env.MODE}</p>
            </div>
          )}
        </Container>
      </div>
      <Footer />
    </div>
  );
};

export default Register;
