import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from 'sonner';
import { candidateRecruitmentService } from '@/services/api';

type BillingItem = {
  candidateId: number;
  candidateName: string;
  requirementId: number;
  requirementJobId?: string;
  requirementTitle: string;
  clientId: number;
  clientName: string;
  joiningId: number;
  joiningDate: string;
  billingId?: number;
  billingDate?: string;
  amount?: string;
  currency?: string;
  invoiceReference?: string;
  billed: boolean;
};

const getData = <T,>(response: { data?: { data?: T } }) => response.data?.data;

const errorMessage = (error: unknown, fallback: string) => {
  const response = (error as { response?: { data?: { message?: string } } })?.response;
  return response?.data?.message || fallback;
};

const Billing = () => {
  const navigate = useNavigate();
  const [items, setItems] = useState<BillingItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('');
  const [creatingKey, setCreatingKey] = useState<string | null>(null);
  const [amounts, setAmounts] = useState<Record<string, string>>({});
  const [invoiceReferences, setInvoiceReferences] = useState<Record<string, string>>({});

  const loadWorklist = async () => {
    setLoading(true);
    try {
      const response = await candidateRecruitmentService.getBillingWorklist();
      setItems(getData<BillingItem[]>(response) || []);
    } catch (error) {
      toast.error(errorMessage(error, 'Failed to load billing worklist'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadWorklist();
  }, []);

  const filteredItems = useMemo(() => {
    const query = filter.trim().toLowerCase();
    if (!query) return items;
    return items.filter((item) =>
      [item.candidateName, item.requirementTitle, item.requirementJobId, item.clientName]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(query)),
    );
  }, [filter, items]);

  const createBilling = async (item: BillingItem) => {
    const key = `${item.candidateId}-${item.requirementId}`;
    const amount = amounts[key]?.trim() || '';
    if (!amount) {
      toast.error('Billing amount is required');
      return;
    }

    setCreatingKey(key);
    try {
      await candidateRecruitmentService.createBilling(item.candidateId, item.requirementId, {
        amount,
        currency: 'INR',
        invoiceReference: invoiceReferences[key]?.trim() || undefined,
      });
      toast.success('Billing record created');
      await loadWorklist();
    } catch (error) {
      toast.error(errorMessage(error, 'Unable to create billing record'));
    } finally {
      setCreatingKey(null);
    }
  };

  const billedCount = items.filter((item) => item.billed).length;
  const pendingCount = items.length - billedCount;

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Billing</h1>
            <p className="text-ats-gray-500">
              Worklist of every joined Candidate × Requirement, showing billing recorded or still pending.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-4 mb-6">
            <Card><CardContent className="pt-6"><p className="text-sm text-muted-foreground">Joined candidates</p><p className="text-2xl font-semibold">{items.length}</p></CardContent></Card>
            <Card><CardContent className="pt-6"><p className="text-sm text-muted-foreground">Billed</p><p className="text-2xl font-semibold">{billedCount}</p></CardContent></Card>
            <Card><CardContent className="pt-6"><p className="text-sm text-muted-foreground">Billing pending</p><p className="text-2xl font-semibold">{pendingCount}</p></CardContent></Card>
          </div>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between gap-4">
              <CardTitle>Candidate Billing Worklist</CardTitle>
              <Input
                className="max-w-sm"
                placeholder="Search candidate, client, requirement…"
                value={filter}
                onChange={(event) => setFilter(event.target.value)}
              />
            </CardHeader>
            <CardContent>
              {loading ? (
                <p className="text-sm text-muted-foreground">Loading billing worklist…</p>
              ) : filteredItems.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  {items.length === 0 ? 'No joined candidates are ready for billing.' : 'No billing records match the search.'}
                </p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-left">
                        <th className="p-3">Candidate</th>
                        <th className="p-3">Client</th>
                        <th className="p-3">Requirement</th>
                        <th className="p-3">Joining date</th>
                        <th className="p-3">Billing</th>
                        <th className="p-3">Amount</th>
                        <th className="p-3">Invoice reference</th>
                        <th className="p-3">Action</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredItems.map((item) => {
                        const key = `${item.candidateId}-${item.requirementId}`;
                        return (
                          <tr key={key} className="border-b align-top">
                            <td className="p-3 font-medium">
                              <button
                                className="text-left hover:underline"
                                onClick={() => navigate(`/recruitment/lifecycle?candidateId=${item.candidateId}&requirementId=${item.requirementId}`)}
                              >
                                {item.candidateName}
                              </button>
                            </td>
                            <td className="p-3">{item.clientName}</td>
                            <td className="p-3">
                              {item.requirementJobId ? `${item.requirementJobId} — ${item.requirementTitle}` : item.requirementTitle}
                            </td>
                            <td className="p-3">{new Date(item.joiningDate).toLocaleDateString()}</td>
                            <td className="p-3">
                              {item.billed ? (
                                <span className="font-medium">Billed</span>
                              ) : (
                                <span className="text-muted-foreground">Pending</span>
                              )}
                            </td>
                            <td className="p-3">
                              {item.billed ? (
                                <span>{item.currency} {item.amount}</span>
                              ) : (
                                <div className="flex gap-2 min-w-[180px]">
                                  <Label htmlFor={`amount-${key}`} className="sr-only">Amount</Label>
                                  <Input
                                    id={`amount-${key}`}
                                    placeholder="Amount"
                                    value={amounts[key] || ''}
                                    onChange={(event) => setAmounts((current) => ({ ...current, [key]: event.target.value }))}
                                  />
                                </div>
                              )}
                            </td>
                            <td className="p-3">
                              {item.billed ? (
                                item.invoiceReference || '—'
                              ) : (
                                <Input
                                  placeholder="Optional"
                                  value={invoiceReferences[key] || ''}
                                  onChange={(event) => setInvoiceReferences((current) => ({ ...current, [key]: event.target.value }))}
                                />
                              )}
                            </td>
                            <td className="p-3">
                              {item.billed ? (
                                <Button variant="outline" size="sm" onClick={() => navigate(`/recruitment/lifecycle?candidateId=${item.candidateId}&requirementId=${item.requirementId}`)}>
                                  Lifecycle
                                </Button>
                              ) : (
                                <Button size="sm" disabled={creatingKey === key} onClick={() => void createBilling(item)}>
                                  {creatingKey === key ? 'Saving…' : 'Create Billing'}
                                </Button>
                              )}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default Billing;
