import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, Filter, UserPlus, ChevronRight } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { toast } from 'sonner';
import { useIsMobile } from '@/hooks/use-mobile';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { candidateService } from '@/services/api';

interface Candidate {
  id: number;
  name: string;
  role: string;
  location: string;
  status: string;
  date: string;
  email: string;
  phone?: string;
  position?: string;
  experience?: string;
  currentCTC?: string;
  expectedCTC?: string;
  noticePeriod?: string;
  skills?: string;
}

// Shape actually returned by GET /api/candidates as of the schema-mismatch
// fix (docs/architecture.md). Note there is currently no `status` or
// `source` field on the backend — status filtering/updating in this page is
// client-side only until the candidate_statuses/status_id design (section
// 12.5/13.6) is actually implemented.
interface ApiCandidate {
  id: number;
  name: string;
  email: string;
  phone?: string;
  position?: string;
  location?: string;
  experience?: string;
  currentCTC?: string;
  expectedCTC?: string;
  noticePeriod?: string;
  skills?: string;
  createdAt?: string;
}

const Candidates = () => {
  const navigate = useNavigate();
  const isMobile = useIsMobile();
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [filteredCandidates, setFilteredCandidates] = useState<Candidate[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchCandidates = async () => {
    setIsLoading(true);
    setError(null);
    try {
      console.log('Fetching candidates from API...');
      const response = await candidateService.getAllCandidates();
      console.log('API response:', response);
      
      if (response.data && response.data.data) {
        // Transform API data to match our Candidate interface
        const apiCandidates = response.data.data.map((candidate: ApiCandidate) => ({
          id: candidate.id,
          name: candidate.name,
          email: candidate.email,
          phone: candidate.phone,
          role: candidate.position || 'No Position',
          location: candidate.location || 'Not specified',
          // No real status field exists on the backend yet (see ApiCandidate
          // comment above) — this is a client-side-only placeholder until
          // section 12.5/13.6's candidate_statuses design is implemented.
          status: 'applied',
          date: candidate.createdAt ? new Date(candidate.createdAt).toLocaleDateString() : 'Recently',
          position: candidate.position,
          experience: candidate.experience,
          currentCTC: candidate.currentCTC,
          expectedCTC: candidate.expectedCTC,
          noticePeriod: candidate.noticePeriod,
          skills: candidate.skills,
        }));
        console.log('Transformed candidates:', apiCandidates);
        setCandidates(apiCandidates);
      } else {
        console.warn('No candidate data received from API');
        setCandidates([]);
      }
    } catch (error) {
      console.error('Error fetching candidates from API:', error);
      setError('Failed to load candidates from database. Please check your connection.');
      setCandidates([]);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchCandidates();
  }, []);

  useEffect(() => {
    if (searchTerm) {
      const filtered = candidates.filter(candidate => 
        candidate.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        candidate.role.toLowerCase().includes(searchTerm.toLowerCase()) ||
        candidate.email.toLowerCase().includes(searchTerm.toLowerCase())
      );
      setFilteredCandidates(filtered);
    } else {
      setFilteredCandidates(candidates);
    }
  }, [searchTerm, candidates]);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const addCandidate = () => {
    navigate('/candidates/add');
  };

  const updateCandidateStatus = async (id: number, status: string) => {
    try {
      const candidateToUpdate = candidates.find(c => c.id === id);
      if (!candidateToUpdate) return;

      // NOTE: the backend's candidates table has no status column yet (see
      // ApiCandidate comment above) — UpdateCandidate will silently ignore
      // this field. This call currently only updates local UI state below;
      // it does not persist. Real persistence needs the candidate_statuses/
      // status_id migration (docs/architecture.md section 12.5/13.6).
      await candidateService.updateCandidate(id, {
        ...candidateToUpdate,
        status: status
      });

      // Update local state
      const updatedCandidates = candidates.map(candidate => 
        candidate.id === id ? { ...candidate, status } : candidate
      );
      setCandidates(updatedCandidates);
      toast.success(`Candidate status updated to ${status}`);
    } catch (error) {
      console.error('Error updating candidate status:', error);
      toast.error('Failed to update candidate status');
    }
  };

  const viewCandidateDetails = (id: number) => {
    console.log(`View candidate ${id}`);
    // TODO: Navigate to candidate details page when implemented
  };

  const renderCandidateRow = (candidate: Candidate) => (
    <TableRow key={candidate.id}>
      <TableCell className="font-medium">{candidate.name}</TableCell>
      {!isMobile && <TableCell>{candidate.role}</TableCell>}
      {!isMobile && <TableCell>{candidate.email}</TableCell>}
      <TableCell>
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
          ${candidate.status === 'applied' ? 'bg-blue-100 text-blue-800' : ''}
          ${candidate.status === 'screening' ? 'bg-blue-100 text-blue-800' : ''}
          ${candidate.status === 'interview' ? 'bg-yellow-100 text-yellow-800' : ''}
          ${candidate.status === 'offer' ? 'bg-green-100 text-green-800' : ''}
          ${candidate.status === 'rejected' ? 'bg-red-100 text-red-800' : ''}
          ${candidate.status === 'hired' ? 'bg-green-100 text-green-800' : ''}
        `}>
          {candidate.status}
        </span>
      </TableCell>
      {!isMobile && <TableCell>{candidate.date}</TableCell>}
      <TableCell className="text-right">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm">
              <ChevronRight size={16} />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-[160px]">
            <DropdownMenuItem onClick={() => viewCandidateDetails(candidate.id)}>
              View Details
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => updateCandidateStatus(candidate.id, 'screening')}>
              Mark Screening
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => updateCandidateStatus(candidate.id, 'interview')}>
              Mark Interview
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => updateCandidateStatus(candidate.id, 'rejected')}>
              Mark Rejected
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  );

  const renderMobileCandidateList = () => (
    <div className="space-y-4">
      {filteredCandidates.length > 0 ? (
        filteredCandidates.map((candidate) => (
          <div key={candidate.id} className="bg-white p-4 rounded-md border border-gray-200 shadow-sm">
            <div className="flex justify-between items-start mb-2">
              <h3 className="font-medium">{candidate.name}</h3>
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
                ${candidate.status === 'applied' ? 'bg-blue-100 text-blue-800' : ''}
                ${candidate.status === 'screening' ? 'bg-blue-100 text-blue-800' : ''}
                ${candidate.status === 'interview' ? 'bg-yellow-100 text-yellow-800' : ''}
                ${candidate.status === 'offer' ? 'bg-green-100 text-green-800' : ''}
                ${candidate.status === 'rejected' ? 'bg-red-100 text-red-800' : ''}
                ${candidate.status === 'hired' ? 'bg-green-100 text-green-800' : ''}
              `}>
                {candidate.status}
              </span>
            </div>
            <div className="text-sm text-gray-500 mb-1">{candidate.role}</div>
            <div className="text-sm text-gray-500 mb-3">{candidate.email} • {candidate.date}</div>
            <div className="flex justify-end">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="sm">
                    Actions <ChevronRight size={16} />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-[160px]">
                  <DropdownMenuItem onClick={() => viewCandidateDetails(candidate.id)}>
                    View Details
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => updateCandidateStatus(candidate.id, 'screening')}>
                    Mark Screening
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => updateCandidateStatus(candidate.id, 'interview')}>
                    Mark Interview
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => updateCandidateStatus(candidate.id, 'rejected')}>
                    Mark Rejected
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        ))
      ) : (
        <div className="text-center py-8 text-gray-500">
          {searchTerm ? 'No candidates found matching your search.' : 'No candidates found in database.'}
        </div>
      )}
    </div>
  );

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Candidates</h1>
            <p className="text-ats-gray-500">Manage and track all candidates in your talent pool.</p>
          </div>

          <Card className="mb-8">
            <CardContent className="p-6">
              <div className="flex flex-col md:flex-row justify-between gap-4 mb-6">
                <div className="relative w-full md:w-80">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-ats-gray-400" size={18} />
                  <Input 
                    placeholder="Search candidates..." 
                    className="pl-10" 
                    value={searchTerm}
                    onChange={handleSearch}
                  />
                </div>
                
                <div className="flex gap-3">
                  <Button variant="outline" size="sm" className="flex gap-2">
                    <Filter size={16} />
                    Filter
                  </Button>
                  <Button 
                    variant="primary" 
                    size="sm" 
                    className="flex gap-2"
                    onClick={addCandidate}
                  >
                    <UserPlus size={16} />
                    Add Candidate
                  </Button>
                </div>
              </div>

              {error ? (
                <div className="py-12 text-center">
                  <p className="text-red-600 mb-4">{error}</p>
                  <Button onClick={fetchCandidates} variant="outline">
                    Retry
                  </Button>
                </div>
              ) : isLoading ? (
                <div className="py-12 text-center">
                  <p>Loading candidates from database...</p>
                </div>
              ) : isMobile ? (
                renderMobileCandidateList()
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead>Email</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Applied</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredCandidates.length > 0 ? (
                      filteredCandidates.map(renderCandidateRow)
                    ) : (
                      <TableRow>
                        <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                          {searchTerm ? 'No candidates found matching your search.' : 'No candidates found in database.'}
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default Candidates;