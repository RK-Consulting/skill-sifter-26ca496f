import React, { useCallback, useEffect, useState } from 'react';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { resumeAIService, PeriodRow, ActivityRow } from '@/services/resumeAIService';

const periods = ['daily', 'monthly', 'quarterly', 'yearly'];

const Reports = () => {
  const [period, setPeriod] = useState('monthly');
  const [rows, setRows] = useState<PeriodRow[]>([]);
  const [activity, setActivity] = useState<ActivityRow[]>([]);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async (selectedPeriod: string) => {
    setLoading(true);
    try {
      const [reportResponse, activityResponse] = await Promise.all([
        resumeAIService.periodic(selectedPeriod),
        resumeAIService.activity(),
      ]);
      setRows(reportResponse.data?.data || []);
      setActivity(activityResponse.data?.data || []);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load(period);
  }, [load, period]);

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow"><Container>
        <div className="mb-8"><h1 className="text-3xl font-semibold">Reports & Activity</h1><p className="text-ats-gray-500 mt-2">Operational reporting across candidates, resumes, searches, jobs, interviews and business development.</p></div>
        <div className="flex gap-2 mb-6 flex-wrap">{periods.map((p) => <button key={p} onClick={() => setPeriod(p)} className={`px-4 py-2 rounded-lg border capitalize ${period === p ? 'bg-ats-blue text-white' : 'bg-white'}`}>{p}</button>)}</div>
        <Card className="mb-8"><CardHeader><CardTitle>{period[0].toUpperCase() + period.slice(1)} report</CardTitle></CardHeader><CardContent><div className="overflow-x-auto"><table className="w-full text-sm"><thead><tr className="border-b text-left"><th className="p-3">Period</th><th className="p-3">Activities</th><th className="p-3">Candidates</th><th className="p-3">Resumes</th><th className="p-3">AI Searches</th><th className="p-3">Jobs</th><th className="p-3">Interviews</th><th className="p-3">Hires</th><th className="p-3">Business Dev</th></tr></thead><tbody>
          {rows.map((row) => <tr key={row.period} className="border-b"><td className="p-3 font-medium">{row.period}</td><td className="p-3">{row.activities}</td><td className="p-3">{row.candidates}</td><td className="p-3">{row.resumes}</td><td className="p-3">{row.resumeSearches}</td><td className="p-3">{row.jobs}</td><td className="p-3">{row.interviews}</td><td className="p-3">{row.hires}</td><td className="p-3">{row.businessDev}</td></tr>)}
          {!rows.length && !loading && <tr><td colSpan={9} className="p-8 text-center text-gray-500">No activity in this reporting period.</td></tr>}
        </tbody></table></div></CardContent></Card>
        <Card><CardHeader><CardTitle>Activity log</CardTitle></CardHeader><CardContent><div className="overflow-x-auto"><table className="w-full text-sm"><thead><tr className="border-b text-left"><th className="p-3">Time</th><th className="p-3">Action</th><th className="p-3">Entity</th><th className="p-3">ID</th><th className="p-3">Description</th><th className="p-3">Actor</th></tr></thead><tbody>
          {activity.map((item) => <tr key={item.id} className="border-b"><td className="p-3 whitespace-nowrap">{new Date(item.createdAt).toLocaleString()}</td><td className="p-3">{item.action}</td><td className="p-3">{item.entityType}</td><td className="p-3">{item.entityId || '—'}</td><td className="p-3">{item.description}</td><td className="p-3">{item.actorUserId || 'system'}</td></tr>)}
          {!activity.length && <tr><td colSpan={6} className="p-8 text-center text-gray-500">No activity recorded yet.</td></tr>}
        </tbody></table></div></CardContent></Card>
      </Container></main><Footer />
    </div>
  );
};

export default Reports;
