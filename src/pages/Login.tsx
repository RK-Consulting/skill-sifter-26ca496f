
import React from 'react';
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

const formSchema = z.object({
  username: z.string().min(3, {
    message: "Username must be at least 3 characters",
  }),
  password: z.string().min(6, {
    message: "Password must be at least 6 characters",
  }),
});

const Login = () => {
  const navigate = useNavigate();
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  const onSubmit = (values: z.infer<typeof formSchema>) => {
    // In a real application, you would validate credentials against your backend
    // For demo purposes, we'll just accept any input and redirect
    console.log(values);
    localStorage.setItem('user', JSON.stringify({ username: values.username, isLoggedIn: true }));
    toast.success("Login successful");
    navigate('/');
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center">
      <Container className="max-w-md">
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl text-center">Login to ATS</CardTitle>
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
                  <Button type="submit" variant="primary" className="w-full">
                    Login
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
          </CardContent>
        </Card>
      </Container>
    </div>
  );
};

export default Login;
