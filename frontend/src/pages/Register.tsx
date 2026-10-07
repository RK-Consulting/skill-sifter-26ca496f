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
  const [registrationId, setRegistrationId] = useState<number | null>(null);
  const [verificationCode, setVerificationCode] = useState('');

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
      setRegistrationId(response.data.data.registrationId);
      toast.success('Verification code sent to your email.');
    } catch (error: unknown) {
      toast.error(getErrorMessage(error, 'Could not start registration.'));
    } finally {
      setIsLoading(false);
    }
  };

  const verifyEmail = async () => {
    if (!registrationId || verificationCode.length !== 6) {
      toast.error('Enter the 6-digit verification code.');
      return;
    }
    try {
      setIsLoading(true);
      const response = await authService.verifyEmail(registrationId, verificationCode);
      if (!response.data?.success) {
        throw new Error(response.data?.message || 'Verification failed');
      }
      toast.success('Email verified. Your 2-day trial has started.');
      const values = form.getValues();
      const login = await authService.login({ email: values.email, password: values.password });
      const data = login.data?.data;
      if (!login.data?.success || !data?.token) throw new Error('Verification succeeded, but login could not be completed.');
      localStorage.setItem('token', data.token);
      localStorage.setItem('user', JSON.stringify({
        ...data.user,
        isLoggedIn: true,
        subscriptionStatus: data.subscriptionStatus,
        planCode: data.planCode,
      }));
      navigate('/dashboard', { replace: true });
    } catch (error: unknown) {
      toast.error(getErrorMessage(error, 'Verification failed.'));
    } finally {
      setIsLoading(false);
    }
  };

  if (registrationId) {
    return (
      <div className="min-h-screen bg-slate-950 text-white flex items-center justify-center p-6">
        <Card className="w-full max-w-md bg-slate-900 border-white/10">
          <CardHeader>
            <CardTitle className="text-2xl text-center">Verify your email</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-slate-400 text-center mb-6">
              We sent a 6-digit verification code to <strong>{form.getValues('email')}</strong>.
            </p>
            <Input
              value={verificationCode}
              onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
              inputMode="numeric"
              maxLength={6}
              autoFocus
              className="bg-white text-center text-2xl tracking-[0.4em] text-slate-900 placeholder:text-slate-400"
              placeholder="000000"
            />
            <Button onClick={verifyEmail} variant="primary" className="w-full mt-5" disabled={isLoading}>
              {isLoading ? 'Verifying...' : 'Verify email & start trial'}
            </Button>
            <button onClick={() => { setRegistrationId(null); setVerificationCode(''); }} className="w-full mt-4 text-sm text-slate-500 hover:text-white">
              Back to registration
            </button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-950 text-white flex flex-col">
      <div className="flex-grow relative overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_right,_rgba(59,130,246,0.22),_transparent_38%),radial-gradient(circle_at_bottom_left,_rgba(14,165,233,0.14),_transparent_34%)]" />
        <Container className="relative max-w-6xl py-8 lg:py-12">
          <div className="grid gap-10 lg:grid-cols-[0.82fr_1.18fr] items-start">
            <div className="lg:sticky lg:top-8">
              <div className="flex items-center gap-3 mb-10">
                <img
                  src="/lovable-uploads/35d9a32a-9b4d-4be7-a93d-03a036a4ab8a.png"
                  alt="R K Consulting Logo"
                  className="h-11 w-11 rounded-xl"
                />
                <div>
                  <div className="font-bold text-lg">SkillSifter</div>
                  <div className="text-xs text-slate-400">Recruit. Smarter. Faster.</div>
                </div>
              </div>

              <div className="max-w-xl">
                <div className="inline-flex items-center rounded-full border border-blue-400/20 bg-blue-400/10 px-3 py-1 text-xs font-medium text-blue-200 mb-5">
                  2-day free trial · No credit card required
                </div>
                <h1 className="text-4xl font-bold tracking-tight sm:text-5xl">
                  Build your recruitment workspace in minutes.
                </h1>
                <p className="mt-5 text-lg leading-8 text-slate-300">
                  One workspace for requirements, candidates, screening, interviews, selection and joining.
                </p>
                <div className="mt-8 rounded-2xl border border-white/10 bg-white/[0.06] p-5">
                  <div className="text-sm font-semibold text-white">Simple, predictable pricing</div>
                  <div className="mt-2 text-2xl font-bold">Pay for the platform, not for every click.</div>
                  <p className="mt-2 text-sm text-slate-400">Choose your plan now. You can review and manage your subscription from your account.</p>
                </div>
                <div className="mt-7 grid gap-3 text-sm text-slate-300">
                  {['Full platform access during your trial', 'Company workspace created automatically', 'Upgrade or manage your subscription from Account'].map((item) => (
                    <div key={item} className="flex items-center gap-3">
                      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-emerald-400/15 text-emerald-300">✓</span>
                      {item}
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <Card className="border-white/10 bg-white text-slate-900 shadow-2xl shadow-black/20">
              <CardHeader className="pb-3">
                <div className="text-sm font-semibold text-blue-600">CREATE YOUR ACCOUNT</div>
                <CardTitle className="text-3xl tracking-tight">Start Using SkillSifter</CardTitle>
                <p className="text-sm text-slate-500">Set up your administrator account and choose the plan that fits your team.</p>
              </CardHeader>
              <CardContent>
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
                    <div className="flex items-end justify-between gap-4 mb-3">
                      <div>
                        <div className="font-semibold text-lg">Choose your plan</div>
                        <div className="text-sm text-slate-500 mt-1">You can change your plan later.</div>
                      </div>
                      <div className="hidden sm:block text-xs font-medium text-slate-400">STEP 2 OF 2</div>
                    </div>
                    {plansLoading ? (
                      <div className="text-sm text-ats-gray-500">Loading plans...</div>
                    ) : plans.length === 0 ? (
                      <div className="text-sm text-red-600">No subscription plans are currently available.</div>
                    ) : (
                      <div className="grid gap-4 md:grid-cols-2">
                        {plans.map((plan) => {
                          const selected = selectedPlan === plan.code;
                          return (
                            <button
                              key={plan.code}
                              type="button"
                              onClick={() => setSelectedPlan(plan.code)}
                              className={`relative text-left rounded-2xl border p-5 transition-all hover:-translate-y-0.5 hover:shadow-lg ${selected ? 'border-blue-600 ring-2 ring-blue-100 bg-blue-50/50' : 'border-slate-200 bg-white'}`}
                            >
                              {selected && <span className="absolute right-4 top-4 rounded-full bg-blue-600 px-2 py-1 text-[10px] font-bold uppercase tracking-wide text-white">Selected</span>}
                              <div className="font-semibold text-lg">{plan.name}</div>
                              <div className="mt-3 flex items-baseline gap-1">
                                <span className="text-2xl font-bold">{plan.currency} {(plan.amountMinor / 100).toLocaleString('en-IN')}</span>
                                <span className="text-xs text-slate-500">/ {plan.billingPeriod}</span>
                              </div>
                              <div className="text-sm text-slate-500 mt-2">Up to {plan.userLimit} users</div>
                              <div className="mt-4 text-xs font-medium text-emerald-700">✓ No per-click charges</div>
                            </button>
                          );
                        })}
                      </div>
                    )}
                  </div>

                  <div className="pt-2">
                    <div className="mb-4 rounded-xl bg-slate-50 border border-slate-200 px-4 py-3 text-xs text-slate-500">
                      By continuing, you agree to create a SkillSifter company workspace. Email verification is required before your free trial begins.
                    </div>
                    <div className="flex flex-col space-y-2">
                    <Button type="submit" variant="primary" className="w-full" disabled={isLoading || plansLoading || !selectedPlan}>
                      {isLoading ? 'Creating account...' : 'Create Account & Verify Email'}
                    </Button>
                    <div className="text-center text-sm mt-4">
                      Already have an account?{' '}
                      <span className="text-ats-blue cursor-pointer hover:underline" onClick={() => navigate('/login')}>
                        Login
                      </span>
                    </div>
                  </div>
                </div>
                </form>
              </Form>
            </CardContent>
          </Card>
          </div>
        </Container>
      </div>
      <div className="bg-slate-950 border-t border-white/10">
        <Footer />
      </div>
    </div>
  );
};

export default Register;
