import React, { ChangeEvent, InputHTMLAttributes, useEffect, useState } from 'react';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import Button from '@/components/ui-custom/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { resumeAIService, ResumeRecord, ResumeResult, UploadResult } from '@/services/resumeAIService';
import { toast } from 'sonner';

type ApiResponse<T> = { data?: T };
type OllamaStatus = { available: boolean; url?: string; model?: string };
type ResumeViewRow = ResumeRecord | ResumeResult | UploadResult;

const folderInputProps = { webkitdirectory: '', directory: '' } as InputHTMLAttributes<HTMLInputElement>;

const ResumeAI = () => {
  const [files, setFiles] = useState<File[]>([]);
  const [uploading, setUploading] = useState(false);
  const [rows, setRows] = useState<ResumeRecord[]>([]);
  const [results, setResults] = useState<ResumeViewRow[]>([]);
  const [q, setQ] = useState('');
  const [searching, setSearching] = useState(false);
  const [ollama, setOllama] = useState<OllamaStatus | null>(null);
  const [detail, setDetail] = useState<Record<string, unknown> | null>(null);
  const [actionId, setActionId] = useState<number | null>(null);

  const load = async () => {
    try {
      const response = (await resumeAIService.list()) as ApiResponse<ResumeRecord[]>;
      setRows(response.data?.data || []);
    } catch (error) {
      console.error(error);
    }
  };

  useEffect(() => {
    void load();
    resumeAIService.health()
      .then((response) => setOllama((response as ApiResponse<OllamaStatus>).data || null))
      .catch(() => setOllama({ available: false }));
  }, []);

  const upload = async () => {
    if (!files.length) return;
    setUploading(true);
    try {
      const response = (await resumeAIService.upload(files)) as ApiResponse<UploadResult[]>;
      const data = response.data?.data || [];
      setResults(data);
      await load();
      toast.success(`Processed ${data.length} resume(s)`);
    } catch (error: unknown) {
      toast.error(error instanceof Error ? error.message : 'Resume processing failed');
    } finally {
      setUploading(false);
    }
  };

  const search = async () => {
    if (!q.trim()) return;
    setSearching(true);
    try {
      const response = (await resumeAIService.search(q)) as ApiResponse<ResumeResult[]>;
      setResults(response.data?.data || []);
    } catch (error: unknown) {
      toast.error(error instanceof Error ? error.message : 'Search failed');
    } finally {
      setSearching(false);
    }
  };

  const handleFilesChange = (event: ChangeEvent<HTMLInputElement>) => {
    setFiles(Array.from(event.target.files || []));
  };

  const showDetail = async (id: number) => {
    try {
      const response = (await resumeAIService.detail(id)) as ApiResponse<Record<string, unknown>>;
      setDetail(response.data?.data || null);
    } catch (error: unknown) {
      toast.error(error instanceof Error ? error.message : 'Unable to load resume detail');
    }
  };

  const retry = async (id: number) => {
    setActionId(id);
    try {
      await resumeAIService.retry(id);
      toast.success('Resume reprocessing completed');
      await load();
      await showDetail(id);
    } catch (error: unknown) {
      toast.error(error instanceof Error ? error.message : 'Resume retry failed');
      await load();
    } finally {
      setActionId(null);
    }
  };

  const download = async (id: number, fileName: string) => {
    try {
      const response = await resumeAIService.download(id);
      const url = window.URL.createObjectURL(response.data);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = fileName;
      anchor.click();
      window.URL.revokeObjectURL(url);
    } catch (error: unknown) {
      toast.error(error instanceof Error ? error.message : 'Resume download failed');
    }
  };

  const getCandidate = (row: ResumeViewRow) => 'candidate' in row ? row.candidate : undefined;
  const getValues = (row: ResumeViewRow) => {
    const candidate = getCandidate(row);
    return {
      name: 'name' in row ? row.name : candidate?.name,
      email: 'email' in row ? row.email : candidate?.email,
      phone: 'phone' in row ? row.phone : candidate?.phone,
      skills: 'skills' in row ? row.skills : candidate?.skills,
      resumeFile: 'resumeFile' in row ? row.resumeFile : 'fileName' in row ? row.fileName : undefined,
      status: 'resumeStatus' in row ? row.resumeStatus : 'status' in row ? row.status : undefined,
      error: 'error' in row ? row.error : undefined,
    };
  };

  const display: ResumeViewRow[] = results.length ? results : rows;

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8"><h1 className="text-3xl font-semibold">Resume AI</h1><p className="text-ats-gray-500 mt-2">Upload an entire resume folder, extract name, email, phone and skills with local Ollama, then search the repository.</p></div>
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
            <Card className="lg:col-span-2"><CardHeader><CardTitle>Upload resume folder</CardTitle></CardHeader><CardContent>
              <input type="file" multiple accept=".pdf,.docx,.txt,.md" {...folderInputProps} onChange={handleFilesChange} className="w-full border rounded-lg p-3" />
              <p className="text-sm text-gray-500 mt-2">Choose a folder in the file picker. Nested PDF/DOCX/text files are supported. Maximum 10MB per file.</p>
              <div className="mt-4 flex items-center gap-3"><Button onClick={upload} disabled={!files.length || uploading}>{uploading ? 'Processing with Ollama…' : `Process ${files.length || 0} resume(s)`}</Button><span className="text-sm text-gray-500">{files.length ? `${files.length} selected` : ''}</span></div>
            </CardContent></Card>
            <Card><CardHeader><CardTitle>Local AI status</CardTitle></CardHeader><CardContent>
              <div className={`text-lg font-semibold ${ollama?.available ? 'text-green-600' : 'text-red-600'}`}>{ollama?.available ? 'Ollama connected' : 'Ollama unavailable'}</div>
              <div className="text-sm text-gray-500 mt-2">URL: {ollama?.url || 'http://127.0.0.1:11434'}</div><div className="text-sm text-gray-500">Model: {ollama?.model || 'llama3.1:8b'}</div>
            </CardContent></Card>
          </div>
          <Card className="mb-6"><CardHeader><CardTitle>AI resume search</CardTitle></CardHeader><CardContent><div className="flex gap-2">
            <input value={q} onChange={(event) => setQ(event.target.value)} onKeyDown={(event) => { if (event.key === 'Enter') void search(); }} placeholder="e.g. Java Spring AWS" className="flex-1 border rounded-lg px-3 py-2" />
            <Button onClick={search} disabled={searching}>{searching ? 'Searching…' : 'Search'}</Button>
          </div></CardContent></Card>
          <Card><CardHeader><CardTitle>{results.length ? 'Search results' : 'Resume repository'}</CardTitle></CardHeader><CardContent><div className="overflow-x-auto"><table className="w-full text-sm"><thead><tr className="border-b text-left"><th className="p-3">Name</th><th className="p-3">Email</th><th className="p-3">Phone</th><th className="p-3">Skills</th><th className="p-3">Resume</th><th className="p-3">Status</th><th className="p-3">Actions</th></tr></thead><tbody>
            {display.map((row) => { const values = getValues(row); return <tr key={'id' in row ? row.id : row.resumeId || row.fileName} className="border-b"><td className="p-3 font-medium">{values.name || '—'}</td><td className="p-3">{values.email || '—'}</td><td className="p-3">{values.phone || '—'}</td><td className="p-3 max-w-md">{values.skills || '—'}</td><td className="p-3">{values.resumeFile || '—'}</td><td className="p-3"><span className="rounded-full px-2 py-1 bg-gray-100">{values.status || '—'}</span>{values.error && <div className="text-red-500 mt-1">{values.error}</div>}</td><td className="p-3"><div className="flex gap-2"><Button onClick={() => void showDetail(('id' in row ? row.id : row.resumeId) as number)}>Details</Button>{('fileName' in row && row.status === 'failed') && <Button onClick={() => void retry(row.id)} disabled={actionId === row.id}>{actionId === row.id ? 'Retrying…' : 'Retry'}</Button>}{'fileName' in row && <Button onClick={() => void download(row.id, row.fileName)}>Download</Button>}</div></td></tr>; })}
            {!display.length && <tr><td colSpan={7} className="p-8 text-center text-gray-500">No resumes yet. Upload a folder to begin.</td></tr>}
          </tbody></table></div></CardContent></Card>
        <Card className="mt-6"><CardHeader><CardTitle>Resume detail</CardTitle></CardHeader><CardContent>{detail ? <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm"><div><strong>File:</strong> {String(detail.fileName || '—')}</div><div><strong>Status:</strong> {String(detail.status || '—')}</div><div><strong>Parser:</strong> {String(detail.parserModel || '—')}</div><div><strong>Candidate:</strong> {String(detail.candidateId || '—')}</div><div><strong>Uploaded:</strong> {String(detail.uploadedAt || '—')}</div><div><strong>Parsed:</strong> {String(detail.parsedAt || '—')}</div>{detail.error && <div className="md:col-span-2 text-red-600"><strong>Error:</strong> {String(detail.error)}</div>}<div className="md:col-span-2"><Button onClick={() => void download(Number(detail.id), String(detail.fileName || 'resume'))}>Secure download</Button></div></div> : <p className="text-sm text-gray-500">Select Details to inspect a resume.</p>}</CardContent></Card>
        </Container>
      </main><Footer />
    </div>
  );
};

export default ResumeAI;
