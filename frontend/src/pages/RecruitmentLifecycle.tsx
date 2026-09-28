import { useEffect, useState } from 'react';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from 'sonner';
import {
  candidateRecruitmentService,
  candidateService,
  requirementService,
} from '@/services/api';

type Candidate = { id: number; name: string };
type Requirement = { id: number; jobId?: string; title?: string; status?: string };

type Selection = {
  id: number;
  decision: string;
  decisionNotes?: string;
  nextAction?: string;
};

type Offer = {
  id: number;
  accepted: boolean;
  offerReference?: string;
};

type Joining = {
  id: number;
  joined: boolean;
  joiningDate?: string | null;
};

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

const RecruitmentLifecycle = () => {
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [requirements, setRequirements] = useState<Requirement[]>([]);
  const [candidateId, setCandidateId] = useState('');
  const [requirementId, setRequirementId] = useState('');

  const [selection, setSelection] = useState<Selection | null>(null);
  const [offer, setOffer] = useState<Offer | null>(null);
  const [joining, setJoining] = useState<Joining | null>(null);
  const [billing, setBilling] = useState<Billing | null>(null);
  const [joiningDate, setJoiningDate] = useState('');
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
      setSelection(null);
      setOffer(null);
      setJoining(null);
      setBilling(null);
      return;
    }

    setLoading(true);
    const candidate = Number(candidateId);
    const requirement = Number(requirementId);

    Promise.all([
      candidateRecruitmentService.getSelection(candidate, requirement).catch(() => null),
      candidateRecruitmentService.getOffer(candidate, requirement).catch(() => null),
      candidateRecruitmentService.getJoining(candidate, requirement).catch(() => null),
      candidateRecruitmentService.getBilling(candidate, requirement).catch(() => null),
    ])
      .then(([selectionResponse, offerResponse, joiningResponse, billingResponse]) => {
        setSelection(selectionResponse ? getData<Selection>(selectionResponse) || null : null);
        setOffer(offerResponse ? getData<Offer>(offerResponse) || null : null);
        setJoining(joiningResponse ? getData<Joining>(joiningResponse) || null : null);
        setBilling(billingResponse ? getData<Billing>(billingResponse) || null : null);
      })
      .finally(() => setLoading(false));
  }, [candidateId, requirementId]);

  const selectedCandidate = candidates.find((item) => item.id === Number(candidateId));
  const selectedRequirement = requirements.find((item) => item.id === Number(requirementId));
  const pairReady = Boolean(candidateId && requirementId);

  const refresh = async () => {
    if (!pairReady) return;
    const candidate = Number(candidateId);
    const requirement = Number(requirementId);
    const [selectionResponse, offerResponse, joiningResponse, billingResponse] = await Promise.all([
      candidateRecruitmentService.getSelection(candidate, requirement).catch(() => null),
      candidateRecruitmentService.getOffer(candidate, requirement).catch(() => null),
      candidateRecruitmentService.getJoining(candidate, requirement).catch(() => null),
      candidateRecruitmentService.getBilling(candidate, requirement).catch(() => null),
    ]);
    setSelection(selectionResponse ? getData<Selection>(selectionResponse) || null : null);
    setOffer(offerResponse ? getData<Offer>(offerResponse) || null : null);
    setJoining(joiningResponse ? getData<Joining>(joiningResponse) || null : null);
    setBilling(billingResponse ? getData<Billing>(billingResponse) || null : null);
  };

  const createSelection = async (decision: 'selected' | 'rejected') => {
    try {
      await candidateRecruitmentService.createSelection(Number(candidateId), Number(requirementId), {
        decision,
        decisionNotes: decision === 'selected' ? 'Selected from recruitment lifecycle' : 'Rejected from recruitment lifecycle',
      });
      await refresh();
      toast.success(`Candidate marked ${decision}`);
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to record selection'));
    }
  };

  const createOffer = async () => {
    try {
      await candidateRecruitmentService.createOffer(Number(candidateId), Number(requirementId));
      await refresh();
      toast.success('Offer created');
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to create offer'));
    }
  };

  const acceptOffer = async () => {
    try {
      await candidateRecruitmentService.updateOffer(Number(candidateId), Number(requirementId), true);
      await refresh();
      toast.success('Offer accepted');
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to accept offer'));
    }
  };

  const createJoining = async () => {
    if (!joiningDate) {
      toast.error('Joining date is required');
      return;
    }
    try {
      await candidateRecruitmentService.createJoining(Number(candidateId), Number(requirementId), {
        joiningDate: new Date(`${joiningDate}T00:00:00`).toISOString(),
        joined: true,
      });
      await refresh();
      toast.success('Candidate marked as joined');
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to record joining'));
    }
  };

  const createBilling = async () => {
    try {
      await candidateRecruitmentService.createBilling(Number(candidateId), Number(requirementId), {
        amount,
        currency,
        invoiceReference: invoiceReference || undefined,
      });
      await refresh();
      toast.success('Billing record created');
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to create billing record'));
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight">Recruitment Lifecycle</h1>
            <p className="text-ats-gray-500 mt-2">
              Candidate × Requirement context from Selection through Offer, Joining and Billing.
            </p>
          </div>

          <Card className="mb-6">
            <CardHeader><CardTitle>Recruitment Context</CardTitle></CardHeader>
            <CardContent className="grid md:grid-cols-2 gap-4">
              <div>
                <Label>Candidate</Label>
                <Select value={candidateId} onValueChange={setCandidateId}>
                  <SelectTrigger><SelectValue placeholder="Select candidate" /></SelectTrigger>
                  <SelectContent>
                    {candidates.map((candidate) => (
                      <SelectItem key={candidate.id} value={String(candidate.id)}>{candidate.name}</SelectItem>
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
                        {requirement.jobId ? `${requirement.jobId} — ${requirement.title || 'Requirement'}` : requirement.title || `Requirement #${requirement.id}`}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          {loading && <p className="mb-6 text-sm text-muted-foreground">Loading recruitment state…</p>}

          {pairReady && (
            <div className="space-y-6">
              <Card>
                <CardHeader><CardTitle>1. Selection</CardTitle></CardHeader>
                <CardContent>
                  <p className="text-sm mb-4">
                    {selectedCandidate?.name} × {selectedRequirement?.title || `Requirement #${requirementId}`}
                  </p>
                  {selection ? (
                    <div className="flex items-center justify-between">
                      <span className="font-medium capitalize">{selection.decision}</span>
                      {selection.nextAction && <span className="text-sm text-muted-foreground">{selection.nextAction}</span>}
                    </div>
                  ) : (
                    <div className="flex gap-3">
                      <Button onClick={() => createSelection('selected')}>Select Candidate</Button>
                      <Button variant="outline" onClick={() => createSelection('rejected')}>Reject Candidate</Button>
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader><CardTitle>2. Offer</CardTitle></CardHeader>
                <CardContent>
                  {!selection || selection.decision !== 'selected' ? (
                    <p className="text-sm text-muted-foreground">A selected decision is required before an offer can be created.</p>
                  ) : !offer ? (
                    <Button onClick={createOffer}>Create Offer</Button>
                  ) : (
                    <div className="flex items-center justify-between gap-4">
                      <div>
                        <p className="font-medium">Offer created</p>
                        <p className="text-sm text-muted-foreground">{offer.offerReference || `Offer #${offer.id}`}</p>
                      </div>
                      {!offer.accepted && <Button onClick={acceptOffer}>Accept Offer</Button>}
                      {offer.accepted && <span className="font-medium">Accepted</span>}
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader><CardTitle>3. Joining</CardTitle></CardHeader>
                <CardContent>
                  {!offer?.accepted ? (
                    <p className="text-sm text-muted-foreground">An accepted offer is required before joining can be recorded.</p>
                  ) : joining ? (
                    <div>
                      <p className="font-medium">{joining.joined ? 'Joined' : 'Joining record created'}</p>
                      {joining.joiningDate && <p className="text-sm text-muted-foreground">{new Date(joining.joiningDate).toLocaleDateString()}</p>}
                    </div>
                  ) : (
                    <div className="flex flex-col md:flex-row md:items-end gap-3">
                      <div>
                        <Label htmlFor="joining-date">Joining date</Label>
                        <Input id="joining-date" type="date" value={joiningDate} onChange={(event) => setJoiningDate(event.target.value)} />
                      </div>
                      <Button onClick={createJoining}>Record Joining</Button>
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader><CardTitle>4. Billing</CardTitle></CardHeader>
                <CardContent>
                  {!joining?.joined ? (
                    <p className="text-sm text-muted-foreground">A joined Candidate × Requirement is required before billing.</p>
                  ) : billing ? (
                    <div className="grid md:grid-cols-3 gap-4">
                      <div><span className="text-sm text-muted-foreground">Amount</span><p className="font-medium">{billing.currency} {billing.amount}</p></div>
                      <div><span className="text-sm text-muted-foreground">Billing date</span><p className="font-medium">{billing.billingDate ? new Date(billing.billingDate).toLocaleDateString() : '—'}</p></div>
                      <div><span className="text-sm text-muted-foreground">Invoice reference</span><p className="font-medium">{billing.invoiceReference || '—'}</p></div>
                    </div>
                  ) : (
                    <div className="grid md:grid-cols-4 gap-3 items-end">
                      <div><Label htmlFor="billing-amount">Amount</Label><Input id="billing-amount" value={amount} onChange={(event) => setAmount(event.target.value)} placeholder="e.g. 50000" /></div>
                      <div><Label htmlFor="billing-currency">Currency</Label><Input id="billing-currency" maxLength={3} value={currency} onChange={(event) => setCurrency(event.target.value.toUpperCase())} /></div>
                      <div><Label htmlFor="invoice-reference">Invoice reference</Label><Input id="invoice-reference" value={invoiceReference} onChange={(event) => setInvoiceReference(event.target.value)} /></div>
                      <Button onClick={createBilling} disabled={!amount || currency.length !== 3}>Create Billing</Button>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          )}
        </Container>
      </main>
    </div>
  );
};

export default RecruitmentLifecycle;
