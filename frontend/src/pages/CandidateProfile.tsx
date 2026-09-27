import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import Button from '@/components/ui-custom/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { candidateService } from '@/services/api';

type Row = Record<string, unknown>;
type Intelligence = {
  sourceResume?: Row | null;
  professionalProfile?: Row | null;
  languages?: Row[];
  technicalExpertise?: Row[];
  employmentHistory?: Row[];
  education?: Row[];
  certifications?: Row[];
  projects?: Row[];
};

const text=(v:unknown)=>v===null||v===undefined||v===''?'—':String(v);

const CandidateProfile=()=>{
  const {id}=useParams();
  const navigate=useNavigate();
  const [candidate,setCandidate]=useState<Row|null>(null);
  const [data,setData]=useState<Intelligence|null>(null);
  const [loading,setLoading]=useState(true);
  const [error,setError]=useState<string|null>(null);

  useEffect(()=>{
    const candidateID=Number(id);
    if(!candidateID){setError('Invalid candidate ID');setLoading(false);return;}
    Promise.all([candidateService.getCandidateById(candidateID),candidateService.getResumeIntelligence(candidateID)])
      .then(([candidateResponse,intelligenceResponse])=>{
        setCandidate(candidateResponse.data?.data || null);
        setData(intelligenceResponse.data?.data || null);
      })
      .catch(()=>setError('Failed to load candidate profile'))
      .finally(()=>setLoading(false));
  },[id]);

  if(loading)return <div className="min-h-screen flex items-center justify-center">Loading candidate profile...</div>;
  if(error)return <div className="min-h-screen flex items-center justify-center"><div className="text-center"><p className="text-red-600 mb-4">{error}</p><Button onClick={()=>navigate('/candidates')}>Back to Candidates</Button></div></div>;

  const profile=data?.professionalProfile;
  const source=data?.sourceResume;
  const renderList=(title:string,rows:Row[]|undefined,fields:string[])=>(
    <Card><CardHeader><CardTitle>{title}</CardTitle></CardHeader><CardContent>
      {rows?.length ? <div className="space-y-4">{rows.map((row,index)=><div key={String(row.id??index)} className="border-b last:border-b-0 pb-3 last:pb-0"><div className="font-medium">{text(row[fields[0]])}</div>{fields.slice(1).map(field=><div key={field} className="text-sm text-gray-600 mt-1">{field}: {text(row[field])}</div>)}</div>)}</div> : <p className="text-sm text-gray-500">No Resume AI data available.</p>}
    </CardContent></Card>
  );

  return <div className="min-h-screen bg-background flex flex-col"><Navbar/><main className="pt-24 pb-10 flex-grow"><Container>
    <div className="mb-6 flex items-center justify-between"><div><h1 className="text-3xl font-semibold">{text(candidate?.name)}</h1><p className="text-gray-500">{text(candidate?.email)} · {text(candidate?.phone)}</p></div><Button variant="outline" onClick={()=>navigate('/candidates')}>Back</Button></div>
    <Card className="mb-6"><CardHeader><CardTitle>Resume AI source</CardTitle></CardHeader><CardContent><p className="text-sm"><strong>Current source:</strong> {text(source?.fileName)} · Parsed {text(source?.parsedAt)} · Parser {text(source?.parserModel)}</p><p className="text-xs text-gray-500 mt-2">Resume AI intelligence shown here comes from the latest successfully processed resume. Historical source records remain preserved.</p></CardContent></Card>
    <Card className="mb-6"><CardHeader><CardTitle>Professional Profile</CardTitle></CardHeader><CardContent><div className="grid md:grid-cols-2 gap-4"><div><strong>Current title:</strong> {text(profile?.currentTitle)}</div><div><strong>Location:</strong> {text(profile?.location)}</div><div><strong>Total experience:</strong> {text(profile?.totalExperience)}</div><div><strong>Relevant experience:</strong> {text(profile?.relevantExperience)}</div><div className="md:col-span-2"><strong>Summary:</strong><p className="mt-1 text-gray-700">{text(profile?.professionalSummary)}</p></div></div></CardContent></Card>
    <div className="grid lg:grid-cols-2 gap-6">
      {renderList('Technical Expertise',data?.technicalExpertise,['skill','category','proficiencyLevel'])}
      {renderList('Languages',data?.languages,['language','proficiencyLevel','proficiencyFramework'])}
      {renderList('Employment History',data?.employmentHistory,['employer','jobTitle','startDate','endDate','description'])}
      {renderList('Education',data?.education,['institution','degree','fieldOfStudy','startDate','endDate'])}
      {renderList('Certifications',data?.certifications,['name','issuer','issueDate','expiryDate','credentialReference'])}
      {renderList('Projects',data?.projects,['projectName','role','technologies','description'])}
    </div>
  </Container></main><Footer/></div>;
};

export default CandidateProfile;
