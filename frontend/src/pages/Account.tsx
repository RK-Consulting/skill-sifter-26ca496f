import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import Button from '@/components/ui-custom/Button';
import { authService, subscriptionService } from '@/services/api';

const Account = () => {
  const queryClient = useQueryClient();

  const accountQuery = useQuery({
    queryKey: ['current-account'],
    queryFn: async () => (await authService.getCurrentAccount()).data.data,
  });

  const subscriptionQuery = useQuery({
    queryKey: ['subscription-account'],
    queryFn: async () => (await subscriptionService.getSubscription()).data.data,
  });

  const cancelMutation = useMutation({
    mutationFn: () => subscriptionService.cancel(),
    onSuccess: () => {
      toast.success('Cancellation requested');
      queryClient.invalidateQueries({ queryKey: ['subscription-account'] });
      queryClient.invalidateQueries({ queryKey: ['current-account'] });
    },
    onError: () => toast.error('Could not request subscription cancellation'),
  });

  const account = accountQuery.data;
  const subscription = subscriptionQuery.data;
  const isAdmin = account?.role === 'admin';

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-2">Account & Subscription</h1>
            <p className="text-ats-gray-500">View your SkillSifter account and subscription status.</p>
          </div>

          <div className="grid gap-6 lg:grid-cols-2">
            <Card>
              <CardHeader><CardTitle>Account</CardTitle></CardHeader>
              <CardContent className="space-y-3">
                <div><span className="font-medium">Company:</span> {account?.companyName || '—'}</div>
                <div><span className="font-medium">Role:</span> {account?.role || '—'}</div>
                <div><span className="font-medium">Users:</span> {account?.userCount ?? '—'} / {account?.userLimit ?? '—'}</div>
                <div><span className="font-medium">Access:</span> {account?.accountStatus || '—'}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader><CardTitle>Current Subscription</CardTitle></CardHeader>
              <CardContent className="space-y-3">
                {subscriptionQuery.isLoading ? (
                  <div>Loading subscription...</div>
                ) : subscriptionQuery.isError ? (
                  <div className="text-ats-gray-500">No subscription record found.</div>
                ) : (
                  <>
                    <div><span className="font-medium">Plan:</span> {subscription?.planName || subscription?.planCode || '—'}</div>
                    <div><span className="font-medium">Status:</span> {subscription?.status || account?.subscriptionStatus || '—'}</div>
                    <div><span className="font-medium">Provider:</span> {subscription?.provider || '—'}</div>
                    {subscription?.endsAt && <div><span className="font-medium">Ends:</span> {new Date(subscription.endsAt).toLocaleDateString()}</div>}
                    {isAdmin && ['TRIAL', 'ACTIVE', 'PAST_DUE'].includes(subscription?.status) && (
                      <Button variant="outline" onClick={() => cancelMutation.mutate()} disabled={cancelMutation.isPending}>
                        {cancelMutation.isPending ? 'Requesting...' : 'Cancel Subscription'}
                      </Button>
                    )}
                  </>
                )}
              </CardContent>
            </Card>
          </div>
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default Account;
