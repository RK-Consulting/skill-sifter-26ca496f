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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { authService } from '@/services/api';
import { getErrorMessage } from '@/lib/utils';

const formSchema = z.object({
  username: z.string().min(2, {
    message: "Username must be at least 2 characters",
  }),
  email: z.string().email({
    message: "Please enter a valid email address",
  }),
  password: z.string().min(6, {
    message: "Password must be at least 6 characters",
  }),
  confirmPassword: z.string(),
  company: z.string().min(2, {
    message: "Company name must be at least 2 characters",
  }),
  role: z.string().min(2, {
    message: "Please select a role",
  }),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

const Register = () => {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [generatedCompanyId, setGeneratedCompanyId] = useState<string>('');
  const [showCompanyIdDialog, setShowCompanyIdDialog] = useState(false);
  
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      email: "",
      password: "",
      confirmPassword: "",
      company: "",
      role: "",
    },
  });

  const generateCompanyId = (companyName: string) => {
    const randomNum = Math.floor(Math.random() * 10000).toString().padStart(4, '0');
    return `${companyName.toLowerCase().replace(/\s+/g, '_')}_${randomNum}`;
  };

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    try {
      setIsLoading(true);
      const companyId = generateCompanyId(values.company);
      setGeneratedCompanyId(companyId);
      
      // Updated payload to match backend expectations
      const payload = {
        username: values.username,
        email: values.email,
        password: values.password,
        companyName: values.company, // Changed from company to companyName
        role: values.role
      };
      
      const response = await authService.register(payload);

      if (response.data.success) {
        localStorage.setItem('token', response.data.data.token || 'mock-jwt-token');
        localStorage.setItem('user', JSON.stringify({ 
          username: response.data.data.user.username, 
          email: response.data.data.user.email,
          id: response.data.data.user.id,
          isLoggedIn: true 
        }));
        
        // Show company ID dialog
        setShowCompanyIdDialog(true);
        
        // Log JWT payload for debugging
        if (response.data.data.token) {
          const parts = response.data.data.token.split('.');
          if (parts.length === 3) {
            const payload = JSON.parse(atob(parts[1]));
            console.log("[JWT PAYLOAD]", payload);
          }
        }
      } else {
        toast.error(response.data.message || "Registration failed");
      }
    } catch (error: unknown) {
      console.error("Registration error:", error);
      toast.error(getErrorMessage(error, "Registration failed."));
    } finally {
      setIsLoading(false);
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
              <CardTitle className="text-2xl text-center">Register</CardTitle>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  <FormField
                    control={form.control}
                    name="username"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Username</FormLabel>
                        <FormControl>
                          <Input placeholder="Enter your username" {...field} />
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
                    name="role"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Role</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="Select a role" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value="admin">Admin</SelectItem>
                            <SelectItem value="manager">Manager</SelectItem>
                            <SelectItem value="recruiter">Recruiter</SelectItem>
                            <SelectItem value="team_leader">Team Leader</SelectItem>
                          </SelectContent>
                        </Select>
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
                      {isLoading ? 'Registering...' : 'Register'}
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

          {showCompanyIdDialog && (
            <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
              <div className="bg-white rounded-lg p-6 max-w-md w-full">
                <h3 className="text-lg font-semibold mb-4">Important: Save Your Company ID</h3>
                <p className="mb-4 text-gray-600">
                  Please save your Company ID. You will need this for all future logins:
                </p>
                <div className="bg-gray-100 p-3 rounded mb-4 font-mono text-center">
                  {generatedCompanyId}
                </div>
                <div className="flex justify-end">
                  <Button 
                    onClick={() => {
                      setShowCompanyIdDialog(false);
                      navigate('/');
                    }}
                    variant="primary"
                  >
                    I've Saved It
                  </Button>
                </div>
              </div>
            </div>
          )}
          
        </Container>
      </div>
      <Footer />
    </div>
  );
};

export default Register;