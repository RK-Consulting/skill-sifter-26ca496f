import { useEffect, useState } from 'react';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from 'sonner';
import { candidateRecruitmentService, candidateService, requirementService } from '@/services/api';

type Candidate = { id: number; name: string };
type Requirement = { id: number; jobId?: string; title?: string; status?: string };
type Joining = { joined: boolean; joiningDate?: string | null };
type Billing = {
  id: number;
  amount: string;
  currency: string;
  invoiceReference?: string;
  billingDate?: string;
};

const getData = <T,>(response: { data?: { data?: T } }) => response.data?.data;

const errorMessage = (error: unknown, fallback: string) => {
  const response = (error as { response?: { data?: { message?: string } } })?.response;
  return response?.data?.message || fallback;
};

const Billing = () => {
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [requirements, setRequirements] = useState<Requirement[]>([]);
  const [candidateId, setCandidateId] = useState('');
  const [requirementId, setRequirementId] = useState('');
  const [joining, setJoining] = useState<Joining | null>(null);
  const [billing, setBilling] = useState<Billing | null>(null);
  const [amount, setAmount] = useState('');
  const [currency, setCurrency] = useState('INR');
  const [invoiceReference, setInvoiceReference] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    Promise.all([candidateService.getAllCandidates(), requirementService.getAllRequirements()])
      .then(([candidateResponse, requirementResponse]) => {
        setCandidates((candidateResponse.data?.data || []) as Candidate[]);
        setRequirements((requirementResponse.data?.data || []) as Requirement[]);
      })
      .catch(() => toast.error('Failed to load candidates and requirements'));
  }, []);

  useEffect(() => {
    if (!candidateId || !requirementId) {
      setJoining(null);
      setBilling(null);
      return;
    }

    setLoading(true);
    const candidate = Number(candidateId);
    const requirement = Number(requirementId);

    Promise.all([
      candidateRecruitmentService.getJoining(candidate, requirement).catch(() => null),
      candidateRecruitmentService.getBilling(candidate, requirement).catch(() => null),
    ])
      .then(([joiningResponse, billingResponse]) => {
        setJoining(joiningResponse ? getData<Joining>(joiningResponse) || null : null);
        setBilling(billingResponse ? getData<Billing>(billingResponse) || null : null);
      })
      .finally(() => setLoading(false));
  }, [candidateId, requirementId]);

  const createBilling = async () => {
    if (!candidateId || !requirementId || !amount || currency.length !== 3) return;

    try {
      await candidateRecruitmentService.createBilling(Number(candidateId), Number(requirementId), {
        amount,
        currency,
        invoiceReference: invoiceReference || undefined,
      });

      const response = await candidateRecruitmentService.getBilling(Number(candidateId), Number(requirementId));
      setBilling(getData<Billing>(response) || null);
      toast.success('Billing record created');
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to create billing record'));
    }
  };

  const selectedCandidate = candidates.find((candidate) => candidate.id === Number(candidateId));
  const selectedRequirement = requirements.find((requirement) => requirement.id === Number(requirementId));

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Billing</h1>
            <p className="text-ats-gray-500">
              Record and view the commercial billing event for a joined Candidate × Requirement.
            </p>
          </div>

          <Card className="mb-6">
            <CardHeader><CardTitle>Billing Context</CardTitle></CardHeader>
            <CardContent className="grid md:grid-cols-2 gap-4">
              <div>
                <Label>Candidate</Label>
                <Select value={candidateId} onValueChange={setCandidateId}>
                  <SelectTrigger><SelectValue placeholder="Select candidate" /></SelectTrigger>
                  <SelectContent>
                    {candidates.map((candidate) => (
                      <SelectItem key={candidate.id} value={String(candidate.id)}>
                        {candidate.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Requirement</Label>
                <Select value={requirementId} onValueChange={setRequirementId}>
                  <SelectTrigger><SelectValue placeholder="Select requirement" /></SelectTrigger>
                  <SelectContent>
                    {requirements.filter((item) => item.status !== 'cancelled').map((requirement) => (
                      <SelectItem key={requirement.id} value={String(requirement.id)}>
                        {requirement.jobId
                          ? `${requirement.jobId} — ${requirement.title || 'Requirement'}`
                          : requirement.title || `Requirement #${requirement.id}`}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          {candidateId && requirementId && (
            <Card>
              <CardHeader>
                <CardTitle>
                  {selectedCandidate?.name} × {selectedRequirement?.title || `Requirement #${requirementId}`}
                </CardTitle>
              </CardHeader>
              <CardContent>
                {loading ? (
                  <p className="text-sm text-muted-foreground">Loading billing state…</p>
                ) : !joining?.joined ? (
                  <p className="text-sm text-muted-foreground">
                    Billing can be recorded only after the Candidate × Requirement is marked as joined.
                  </p>
                ) : billing ? (
                  <div className="grid md:grid-cols-3 gap-6">
                    <div>
                      <span className="text-sm text-muted-foreground">Amount</span>
                      <p className="font-medium">{billing.currency} {billing.amount}</p>
                    </div>
                    <div>
                      <span className="text-sm text-muted-foreground">Billing date</span>
                      <p className="font-medium">
                        {billing.billingDate ? new Date(billing.billingDate).toLocaleDateString() : '—'}
                      </p>
                    </div>
                    <div>
                      <span className="text-sm text-muted-foreground">Invoice reference</span>
                      <p className="font-medium">{billing.invoiceReference || '—'}</p>
                    </div>
                  </div>
                ) : (
                  <div className="grid md:grid-cols-4 gap-3 items-end">
                    <div>
                      <Label htmlFor="billing-amount">Amount</Label>
                      <Input
                        id="billing-amount"
                        value={amount}
                        onChange={(event) => setAmount(event.target.value)}
                        placeholder="e.g. 50000"
                      />
                    </div>
                    <div>
                      <Label htmlFor="billing-currency">Currency</Label>
                      <Input
                        id="billing-currency"
                        maxLength={3}
                        value={currency}
                        onChange={(event) => setCurrency(event.target.value.toUpperCase())}
                      />
                    </div>
                    <div>
                      <Label htmlFor="invoice-reference">Invoice reference</Label>
                      <Input
                        id="invoice-reference"
                        value={invoiceReference}
                        onChange={(event) => setInvoiceReference(event.target.value)}
                      />
                    </div>
                    <Button onClick={createBilling} disabled={!amount || currency.length !== 3}>
                      Create Billing
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default Billing;
