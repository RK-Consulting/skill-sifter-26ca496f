
import React, { useState, useEffect } from 'react';
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
  email: z.string().email({
    message: "Please enter a valid email address",
  }),
  password: z.string().min(6, {
    message: "Password must be at least 6 characters",
  }),
  companyId: z.string().min(1, {
    message: "Company ID is required",
  }),
});

// Dummy credentials
const DUMMY_USER = {
  email: "admin@example.com",
  password: "password123"
};

const Login = () => {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    // Check if user is already logged in
    const token = localStorage.getItem('token');
    const user = localStorage.getItem('user');
    
    if (token && user) {
      // User is already logged in, redirect to home
      console.log('User already logged in, redirecting to dashboard');
      navigate('/', { replace: true });
    } else {
      // Mark the component as initialized for rendering
      setIsInitialized(true);
    }
  }, [navigate]);
  
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      email: "",
      password: "",
      companyId: "",
    },
  });

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    try {
      setIsLoading(true);
      
      console.log('Login submission:', { email: values.email, companyId: values.companyId });
      
      // Check if using dummy credentials
      if (values.email === DUMMY_USER.email && values.password === DUMMY_USER.password) {
        // Store dummy user data
        localStorage.setItem('token', 'dummy-token-123456');
        localStorage.setItem('user', JSON.stringify({ 
          username: 'Admin User', 
          email: DUMMY_USER.email,
          id: 1,
          isLoggedIn: true 
        }));
        
        toast.success("Login successful");
        console.log('Login successful with dummy credentials, redirecting to dashboard');
        navigate('/', { replace: true });
        return;
      }
      
      // Regular login via API
      const response = await authService.login(values);

      if (response.data && response.data.success) {
        localStorage.setItem('token', response.data.data.token || 'mock-jwt-token');
        localStorage.setItem('user', JSON.stringify({ 
          username: response.data.data.username || 'User', 
          email: response.data.data.email || values.email,
          id: response.data.data.id || 1,
          isLoggedIn: true 
        }));
        
        // --------- Log JWT payload for debugging ----------
        if (response.data.data.token) {
          const parts = response.data.data.token.split('.');
          if (parts.length === 3) {
            const payload = JSON.parse(atob(parts[1]));
            console.log("[JWT PAYLOAD]", payload);
          }
        }
        // --------------------------------------------------
        
        toast.success("Login successful");
        console.log('Login successful via API, redirecting to dashboard');
        navigate('/', { replace: true });
      } else {
        toast.error(response.data?.message || "Login failed");
      }
    } catch (error: any) {
      console.error("Login error:", error);
      toast.error(error.response?.data?.message || "Invalid credentials. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  // Don't render anything until we've checked authentication status
  if (!isInitialized) {
    return (
      <div className="flex justify-center items-center h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-ats-blue-500"></div>
      </div>
    );
  }

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
              <CardTitle className="text-2xl text-center">Login</CardTitle>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  <FormField
                    control={form.control}
                    name="companyId"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Company ID</FormLabel>
                        <FormControl>
                          <Input placeholder="Enter your company ID" {...field} />
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
                    name="password"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Password</FormLabel>
                        <FormControl>
                          <Input type="password" placeholder="Enter your password" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <div className="flex flex-col space-y-2">
                    <Button 
                      type="submit" 
                      variant="primary" 
                      className="w-full"
                      disabled={isLoading}
                    >
                      {isLoading ? 'Logging in...' : 'Login'}
                    </Button>
                    <div className="text-center text-sm mt-4">
                      Don't have an account?{" "}
                      <span 
                        className="text-ats-blue cursor-pointer hover:underline"
                        onClick={() => navigate('/register')}
                      >
                        Register
                      </span>
                    </div>
                  </div>
                </form>
              </Form>
              
              {/* Dummy credentials info */}
              <div className="mt-8 p-3 bg-blue-50 border border-blue-200 rounded-md">
                <p className="text-sm text-gray-700 mb-1"><strong>Demo Credentials:</strong></p>
                <p className="text-sm text-gray-700">Email: {DUMMY_USER.email}</p>
                <p className="text-sm text-gray-700">Password: {DUMMY_USER.password}</p>
              </div>
            </CardContent>
          </Card>
        </Container>
      </div>
      <Footer />
    </div>
  );
};

export default Login;
