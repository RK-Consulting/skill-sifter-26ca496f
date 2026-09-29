import React, { useEffect, useState } from 'react';
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
import { authService, subscriptionService } from '@/services/api';
import { getErrorMessage } from '@/lib/utils';

type Plan = {
  code: string;
  name: string;
  amountMinor: number;
  currency: string;
  billingInterval: number;
  billingPeriod: string;
  userLimit: number;
};

const formSchema = z.object({
  username: z.string().min(2, { message: 'Name must be at least 2 characters' }),
  email: z.string().email({ message: 'Please enter a valid email address' }),
  password: z.string().min(6, { message: 'Password must be at least 6 characters' }),
  confirmPassword: z.string(),
  company: z.string().min(2, { message: 'Company name must be at least 2 characters' }),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ['confirmPassword'],
});

const Register = () => {
  const navigate = useNavigate();
  const [plans, setPlans] = useState<Plan[]>([]);
  const [selectedPlan, setSelectedPlan] = useState('');
  const [plansLoading, setPlansLoading] = useState(true);
  const [isLoading, setIsLoading] = useState(false);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      email: '',
      password: '',
      confirmPassword: '',
      company: '',
    },
  });

  useEffect(() => {
    subscriptionService.getPlans()
      .then((response) => {
        const availablePlans = response.data?.data ?? [];
        setPlans(availablePlans);
        if (availablePlans.length > 0) {
          setSelectedPlan(availablePlans[0].code);
        }
      })
      .catch(() => toast.error('Could not load subscription plans.'))
      .finally(() => setPlansLoading(false));
  }, []);

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    if (!selectedPlan) {
      toast.error('Please select a subscription plan.');
      return;
    }

    try {
      setIsLoading(true);

      const response = await authService.register({
        username: values.username,
        email: values.email,
        password: values.password,
        companyName: values.company,
        planCode: selectedPlan,
      });

      if (!response.data?.success) {
        throw new Error(response.data?.message || 'Registration failed');
      }

      const data = response.data.data;
      localStorage.setItem('token', data.token);
      localStorage.setItem('user', JSON.stringify({
        ...data.user,
        isLoggedIn: true,
        subscriptionStatus: data.subscriptionStatus,
        planCode: data.planCode,
      }));

      const checkout = await subscriptionService.checkout(selectedPlan);
      const checkoutUrl = checkout.data?.data?.checkoutUrl;

      if (!checkoutUrl) {
        throw new Error('Subscription checkout could not be created.');
      }

      toast.success('Account created. Continue to payment.');
      window.location.assign(checkoutUrl);
    } catch (error: unknown) {
      toast.error(getErrorMessage(error, 'Could not complete SaaS onboarding.'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <div className="flex-grow flex items-center justify-center py-10">
        <Container className="max-w-3xl">
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
              <CardTitle className="text-2xl text-center">Start Using SkillSifter</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="mb-6 p-3 bg-blue-50 border border-blue-200 rounded-md text-sm text-gray-700">
                Create your company account, choose a plan and continue directly to subscription checkout.
              </div>

              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  <div className="grid gap-6 md:grid-cols-2">
                    <FormField
                      control={form.control}
                      name="username"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Administrator name</FormLabel>
                          <FormControl><Input placeholder="Enter your name" {...field} /></FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="company"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Company / Tenant Name</FormLabel>
                          <FormControl><Input placeholder="Enter your company name" {...field} /></FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>

                  <FormField
                    control={form.control}
                    name="email"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Email</FormLabel>
                        <FormControl><Input type="email" autoComplete="email" placeholder="Enter your email" {...field} /></FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <div className="grid gap-6 md:grid-cols-2">
                    <FormField
                      control={form.control}
                      name="password"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Password</FormLabel>
                          <FormControl><Input type="password" autoComplete="new-password" placeholder="Create a password" {...field} /></FormControl>
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
                          <FormControl><Input type="password" autoComplete="new-password" placeholder="Confirm your password" {...field} /></FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>

                  <div>
                    <div className="font-medium mb-3">Choose your plan</div>
                    {plansLoading ? (
                      <div className="text-sm text-ats-gray-500">Loading plans...</div>
                    ) : plans.length === 0 ? (
                      <div className="text-sm text-red-600">No subscription plans are currently available.</div>
                    ) : (
                      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                        {plans.map((plan) => {
                          const selected = selectedPlan === plan.code;
                          return (
                            <button
                              key={plan.code}
                              type="button"
                              onClick={() => setSelectedPlan(plan.code)}
                              className={`text-left border rounded-lg p-5 transition ${selected ? 'ring-2 ring-ats-blue-500' : ''}`}
                            >
                              <div className="font-semibold text-lg">{plan.name}</div>
                              <div className="mt-2">{plan.currency} {(plan.amountMinor / 100).toFixed(2)}</div>
                              <div className="text-sm text-ats-gray-500 mt-1">
                                {plan.billingInterval} {plan.billingPeriod} · {plan.userLimit} users
                              </div>
                            </button>
                          );
                        })}
                      </div>
                    )}
                  </div>

                  <div className="flex flex-col space-y-2">
                    <Button type="submit" variant="primary" className="w-full" disabled={isLoading || plansLoading || !selectedPlan}>
                      {isLoading ? 'Creating account...' : 'Create Account & Continue to Payment'}
                    </Button>
                    <div className="text-center text-sm mt-4">
                      Already have an account?{' '}
                      <span className="text-ats-blue cursor-pointer hover:underline" onClick={() => navigate('/login')}>
                        Login
                      </span>
                    </div>
                  </div>
                </form>
              </Form>
            </CardContent>
          </Card>
        </Container>
      </div>
      <Footer />
    </div>
  );
};

export default Register;
