import { useCallback, useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { ReactNode } from 'react';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from 'sonner';
import {
  candidateRecruitmentService,
  candidateService,
  clientService,
  requirementService,
  interviewService,
} from '@/services/api';

type Candidate = { id: number; name: string };
type Requirement = { id: number; jobId?: string; title?: string; status?: string; clientId?: number };
type Screening = { id: number; requirementId: number; status?: string; recruiterAssessment?: string };
type Submission = { id: number; recipientType: string; recipientName?: string; submittedAt?: string };
type Feedback = { id: number; outcome: string; reasonCode?: string; comments?: string; nextAction?: string };
type Interview = { id: number; requirementId?: number; status: string; outcome?: string; interviewDate: string };
type Selection = { id: number; decision: string; decisionNotes?: string; nextAction?: string };
type Offer = { id: number; accepted: boolean };
type Joining = { id: number; joined: boolean; joiningDate?: string | null };
type Billing = { id: number; amount: string; currency: string; invoiceReference?: string; billingDate?: string };
type Client = { id: number; name: string };

const getData = <T,>(response: { data?: { data?: T } }) => response.data?.data;

const errorMessage = (error: unknown, fallback: string) => {
  const response = (error as { response?: { data?: { message?: string } } })?.response;
  return response?.data?.message || fallback;
};

const currentUserId = () => {
  try {
    const user = JSON.parse(localStorage.getItem('user') || '{}');
    return Number(user.id || 0);
  } catch {
    return 0;
  }
};

const RecruitmentLifecycle = () => {
  const [searchParams] = useSearchParams();
  const initialCandidateId = searchParams.get('candidateId') || '';
  const initialRequirementId = searchParams.get('requirementId') || '';
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [requirements, setRequirements] = useState<Requirement[]>([]);
  const [clients, setClients] = useState<Client[]>([]);
  const [candidateId, setCandidateId] = useState(initialCandidateId);
  const [requirementId, setRequirementId] = useState(initialRequirementId);

  const [screening, setScreening] = useState<Screening | null>(null);
  const [submission, setSubmission] = useState<Submission | null>(null);
  const [feedback, setFeedback] = useState<Feedback | null>(null);
  const [interview, setInterview] = useState<Interview | null>(null);
  const [selection, setSelection] = useState<Selection | null>(null);
  const [offer, setOffer] = useState<Offer | null>(null);
  const [joining, setJoining] = useState<Joining | null>(null);
  const [billing, setBilling] = useState<Billing | null>(null);

  const [assessment, setAssessment] = useState('');
  const [feedbackOutcome, setFeedbackOutcome] = useState('shortlist');
  const [feedbackComments, setFeedbackComments] = useState('');
  const [feedbackReason, setFeedbackReason] = useState('other');
  const [interviewDate, setInterviewDate] = useState('');
  const [interviewOutcome, setInterviewOutcome] = useState('');
  const [joiningDate, setJoiningDate] = useState('');
  const [amount, setAmount] = useState('');
  const [currency, setCurrency] = useState('INR');
  const [invoiceReference, setInvoiceReference] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    Promise.all([
      candidateService.getAllCandidates(),
      requirementService.getAllRequirements(),
      clientService.getAllClients({ page: 1, limit: 100 }),
    ])
      .then(([candidateResponse, requirementResponse, clientResponse]) => {
        setCandidates((candidateResponse.data?.data || []) as Candidate[]);
        setRequirements((requirementResponse.data?.data || []) as Requirement[]);
        setClients((clientResponse.data?.data || []) as Client[]);
      })
      .catch(() => toast.error('Failed to load recruitment context'));
  }, []);

  const loadState = useCallback(async () => {
    if (!candidateId || !requirementId) return;
    const candidate = Number(candidateId);
    const requirement = Number(requirementId);

    const [screenings, submissions, interviews, selectionResponse, offerResponse, joiningResponse, billingResponse] =
      await Promise.all([
        candidateRecruitmentService.getScreenings(candidate).catch(() => null),
        candidateRecruitmentService.getSubmissions(candidate, requirement).catch(() => null),
        candidateRecruitmentService.getInterviews(candidate).catch(() => null),
        candidateRecruitmentService.getSelection(candidate, requirement).catch(() => null),
        candidateRecruitmentService.getOffer(candidate, requirement).catch(() => null),
        candidateRecruitmentService.getJoining(candidate, requirement).catch(() => null),
        candidateRecruitmentService.getBilling(candidate, requirement).catch(() => null),
      ]);

    const screeningItems = screenings ? (getData<Screening[]>(screenings) || []) : [];
    const submissionItems = submissions ? (getData<Submission[]>(submissions) || []) : [];
    const interviewItems = interviews ? (getData<Interview[]>(interviews) || []) : [];

    const currentScreening = screeningItems.find((item) => item.requirementId === requirement) || null;
    const currentSubmission = submissionItems[0] || null;
    const currentInterview =
      interviewItems.find((item) => Number(item.requirementId) === requirement) || null;

    setScreening(currentScreening);
    setSubmission(currentSubmission);
    setInterview(currentInterview);
    setSelection(selectionResponse ? getData<Selection>(selectionResponse) || null : null);
    setOffer(offerResponse ? getData<Offer>(offerResponse) || null : null);
    setJoining(joiningResponse ? getData<Joining>(joiningResponse) || null : null);
    setBilling(billingResponse ? getData<Billing>(billingResponse) || null : null);

    if (currentSubmission) {
      const feedbackResponse = await candidateRecruitmentService.getSubmissionFeedback(currentSubmission.id).catch(() => null);
      const feedbackItems = feedbackResponse ? (getData<Feedback[]>(feedbackResponse) || []) : [];
      setFeedback(feedbackItems[0] || null);
    } else {
      setFeedback(null);
    }
  }, [candidateId, requirementId]);

  useEffect(() => {
    if (!candidateId || !requirementId) {
      setScreening(null); setSubmission(null); setFeedback(null); setInterview(null);
      setSelection(null); setOffer(null); setJoining(null); setBilling(null);
      return;
    }
    setLoading(true);
    loadState().finally(() => setLoading(false));
  }, [candidateId, requirementId, loadState]);

  const selectedCandidate = candidates.find((item) => item.id === Number(candidateId));
  const selectedRequirement = requirements.find((item) => item.id === Number(requirementId));
  const selectedClient = clients.find((item) => item.id === selectedRequirement?.clientId);
  const pairReady = Boolean(candidateId && requirementId);
  const screeningCompleted = screening?.status === 'completed';
  const interviewCompleted = interview?.status === 'completed';

  const refresh = async () => {
    setLoading(true);
    try { await loadState(); } finally { setLoading(false); }
  };

  const createScreening = async () => {
    const userId = currentUserId();
    if (!userId) { toast.error('Authenticated user is required'); return; }
    try {
      const response = await candidateRecruitmentService.createScreening(Number(candidateId), {
        candidateId: Number(candidateId),
        requirementId: Number(requirementId),
        recruiterUserId: userId,
        recruiterAssessment: assessment,
      });
      const created = getData<Screening>(response);
      if (created?.id) {
        await candidateRecruitmentService.updateScreening(Number(candidateId), created.id, 'completed');
      }
      await refresh();
      toast.success('Screening completed');
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to complete screening'));
    }
  };

  const createSubmission = async () => {
    if (!selectedRequirement?.clientId) { toast.error('Requirement client is required'); return; }
    try {
      await candidateRecruitmentService.createSubmission(Number(candidateId), Number(requirementId), {
        recipientType: 'client',
        recipientClientId: selectedRequirement.clientId,
        recipientName: selectedClient?.name,
      });
      await refresh();
      toast.success('Candidate submitted');
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to submit candidate'));
    }
  };

  const createFeedback = async () => {
    if (!submission) return;
    try {
      await candidateRecruitmentService.createSubmissionFeedback(submission.id, {
        outcome: feedbackOutcome,
        comments: feedbackComments,
        ...(feedbackOutcome !== 'shortlist' ? { reasonCode: feedbackReason } : {}),
      });
      await refresh();
      toast.success('Client feedback recorded');
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to record feedback'));
    }
  };

  const scheduleInterview = async () => {
    if (!interviewDate) { toast.error('Interview date is required'); return; }
    try {
      await interviewService.createInterview({
        candidateId: Number(candidateId),
        requirementId: Number(requirementId),
        round: 1,
        interviewDate: new Date(interviewDate).toISOString(),
        status: 'scheduled',
      });
      await refresh();
      toast.success('Interview scheduled');
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to schedule interview'));
    }
  };

  const completeInterview = async () => {
    if (!interview) return;
    try {
      await interviewService.updateInterview(interview.id, {
        candidateId: Number(candidateId),
        requirementId: Number(requirementId),
        round: 1,
        interviewDate: interview.interviewDate,
        status: 'completed',
        outcome: interviewOutcome || 'completed',
      });
      await refresh();
      toast.success('Interview completed');
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to complete interview'));
    }
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
    if (!joiningDate) { toast.error('Joining date is required'); return; }
    try {
      await candidateRecruitmentService.createJoining(Number(candidateId), Number(requirementId), {
        joiningDate,
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

  const stage = (number: number, title: string, content: ReactNode) => (
    <Card>
      <CardHeader><CardTitle>{number}. {title}</CardTitle></CardHeader>
      <CardContent>{content}</CardContent>
    </Card>
  );

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight">Recruitment Lifecycle</h1>
            <p className="text-ats-gray-500 mt-2">
              Requirement → Candidate × Requirement → Screening → Submission → Feedback → Interview → Selection → Offer → Joining → Billing.
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
              {stage(1, 'Screening', screeningCompleted ? (
                <p className="font-medium">Completed{screening?.recruiterAssessment ? ` — ${screening.recruiterAssessment}` : ''}</p>
              ) : (
                <div className="space-y-3">
                  <Input value={assessment} onChange={(e) => setAssessment(e.target.value)} placeholder="Recruiter assessment (optional)" />
                  <Button onClick={createScreening}>Complete Screening</Button>
                </div>
              ))}

              {stage(2, 'Submission', !screeningCompleted ? (
                <p className="text-sm text-muted-foreground">Complete screening first.</p>
              ) : submission ? (
                <p className="font-medium">Submitted to {submission.recipientName || selectedClient?.name || 'client'}</p>
              ) : (
                <Button onClick={createSubmission}>Submit to Client</Button>
              ))}

              {stage(3, 'Feedback', !submission ? (
                <p className="text-sm text-muted-foreground">Submit the candidate first.</p>
              ) : feedback ? (
                <div>
                  <p className="font-medium capitalize">{feedback.outcome}</p>
                  {feedback.comments && <p className="text-sm text-muted-foreground mt-1">{feedback.comments}</p>}
                </div>
              ) : (
                <div className="space-y-3">
                  <Select value={feedbackOutcome} onValueChange={setFeedbackOutcome}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="shortlist">Shortlist</SelectItem>
                      <SelectItem value="hold">Hold</SelectItem>
                      <SelectItem value="reject">Reject</SelectItem>
                    </SelectContent>
                  </Select>
                  {feedbackOutcome !== 'shortlist' && (
                    <Select value={feedbackReason} onValueChange={setFeedbackReason}>
                      <SelectTrigger><SelectValue placeholder="Reason" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="skills_gap">Skills gap</SelectItem>
                        <SelectItem value="experience_gap">Experience gap</SelectItem>
                        <SelectItem value="compensation_mismatch">Compensation mismatch</SelectItem>
                        <SelectItem value="location_mismatch">Location mismatch</SelectItem>
                        <SelectItem value="notice_period">Notice period</SelectItem>
                        <SelectItem value="candidate_not_interested">Candidate not interested</SelectItem>
                        <SelectItem value="availability">Availability</SelectItem>
                        <SelectItem value="profile_mismatch">Profile mismatch</SelectItem>
                        <SelectItem value="other">Other</SelectItem>
                      </SelectContent>
                    </Select>
                  )}
                  <Input value={feedbackComments} onChange={(e) => setFeedbackComments(e.target.value)} placeholder="Client feedback (optional)" />
                  <Button onClick={createFeedback}>Record Feedback</Button>
                </div>
              ))}

              {stage(4, 'Interview', !feedback ? (
                <p className="text-sm text-muted-foreground">Record client feedback first.</p>
              ) : interview?.status === 'completed' ? (
                <div>
                  <p className="font-medium">Completed</p>
                  {interview.outcome && <p className="text-sm text-muted-foreground">{interview.outcome}</p>}
                </div>
              ) : interview ? (
                <div className="space-y-3">
                  <p className="font-medium">Scheduled for {new Date(interview.interviewDate).toLocaleString()}</p>
                  <Input value={interviewOutcome} onChange={(e) => setInterviewOutcome(e.target.value)} placeholder="Interview outcome" />
                  <Button onClick={completeInterview}>Complete Interview</Button>
                </div>
              ) : (
                <div className="flex flex-col md:flex-row md:items-end gap-3">
                  <div>
                    <Label htmlFor="interview-date">Interview date</Label>
                    <Input id="interview-date" type="datetime-local" value={interviewDate} onChange={(e) => setInterviewDate(e.target.value)} />
                  </div>
                  <Button onClick={scheduleInterview}>Schedule Interview</Button>
                </div>
              ))}

              {stage(5, 'Selection', !interviewCompleted ? (
                <p className="text-sm text-muted-foreground">Complete the interview first.</p>
              ) : selection ? (
                <p className="font-medium capitalize">{selection.decision}</p>
              ) : (
                <div className="flex gap-3">
                  <Button onClick={() => createSelection('selected')}>Select Candidate</Button>
                  <Button variant="outline" onClick={() => createSelection('rejected')}>Reject Candidate</Button>
                </div>
              ))}

              {stage(6, 'Offer', !selection || selection.decision !== 'selected' ? (
                <p className="text-sm text-muted-foreground">A selected decision is required before an offer.</p>
              ) : !offer ? (
                <Button onClick={createOffer}>Create Offer</Button>
              ) : (
                <div className="flex items-center justify-between">
                  <p className="font-medium">Offer created</p>
                  {!offer.accepted ? <Button onClick={acceptOffer}>Accept Offer</Button> : <span className="font-medium">Accepted</span>}
                </div>
              ))}

              {stage(7, 'Joining', !offer?.accepted ? (
                <p className="text-sm text-muted-foreground">Accept the offer first.</p>
              ) : joining ? (
                <div>
                  <p className="font-medium">{joining.joined ? 'Joined' : 'Joining record created'}</p>
                  {joining.joiningDate && <p className="text-sm text-muted-foreground">{joining.joiningDate}</p>}
                </div>
              ) : (
                <div className="flex flex-col md:flex-row md:items-end gap-3">
                  <div>
                    <Label htmlFor="joining-date">Joining date</Label>
                    <Input id="joining-date" type="date" value={joiningDate} onChange={(e) => setJoiningDate(e.target.value)} />
                  </div>
                  <Button onClick={createJoining}>Record Joining</Button>
                </div>
              ))}

              {stage(8, 'Billing', !joining?.joined ? (
                <p className="text-sm text-muted-foreground">A joined Candidate × Requirement is required before billing.</p>
              ) : billing ? (
                <div className="grid md:grid-cols-3 gap-4">
                  <div><span className="text-sm text-muted-foreground">Amount</span><p className="font-medium">{billing.currency} {billing.amount}</p></div>
                  <div><span className="text-sm text-muted-foreground">Billing date</span><p className="font-medium">{billing.billingDate ? new Date(billing.billingDate).toLocaleDateString() : '—'}</p></div>
                  <div><span className="text-sm text-muted-foreground">Invoice reference</span><p className="font-medium">{billing.invoiceReference || '—'}</p></div>
                </div>
              ) : (
                <div className="grid md:grid-cols-4 gap-3 items-end">
                  <div><Label htmlFor="billing-amount">Amount</Label><Input id="billing-amount" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="e.g. 50000" /></div>
                  <div><Label htmlFor="billing-currency">Currency</Label><Input id="billing-currency" maxLength={3} value={currency} onChange={(e) => setCurrency(e.target.value.toUpperCase())} /></div>
                  <div><Label htmlFor="invoice-reference">Invoice reference</Label><Input id="invoice-reference" value={invoiceReference} onChange={(e) => setInvoiceReference(e.target.value)} /></div>
                  <Button onClick={createBilling} disabled={!amount || currency.length !== 3}>Create Billing</Button>
                </div>
              ))}
            </div>
          )}
        </Container>
      </main>
    </div>
  );
};

export default RecruitmentLifecycle;
