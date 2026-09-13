import React, { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, UserPlus } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { requirementService } from '@/services/api';

interface Requirement {
  id: number;
  clientId: number;
  title: string;
  department?: string;
  location?: string;
  status: string;
  headcount: number;
  createdAt: string;
}

const statusColors: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  open: 'bg-green-100 text-green-700',
  on_hold: 'bg-yellow-100 text-yellow-700',
  filled: 'bg-blue-100 text-blue-700',
  cancelled: 'bg-red-100 text-red-700',
};

const Requirements = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const clientIdFilter = searchParams.get('clientId');
  const [searchTerm, setSearchTerm] = useState('');

  const { data, isLoading, isError } = useQuery({
    queryKey: ['requirements'],
    queryFn: async () => {
      const response = await requirementService.getAllRequirements();
      return (response.data?.data ?? []) as Requirement[];
    },
  });

  const requirements = data ?? [];

  const filtered = React.useMemo(() => {
    let list = data ?? [];
    if (clientIdFilter) {
      list = list.filter((r) => String(r.clientId) === clientIdFilter);
    }
    if (searchTerm) {
      const term = searchTerm.toLowerCase();
      list = list.filter((r) => r.title.toLowerCase().includes(term));
    }
    return list;
  }, [data, clientIdFilter, searchTerm]);

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Requirements</h1>
            <p className="text-ats-gray-500">Open roles you're recruiting for, by client.</p>
          </div>

          <Card className="mb-8 animate-fade-up">
            <CardContent className="p-6">
              <div className="flex flex-col md:flex-row justify-between gap-4 mb-6">
                <div className="relative w-full md:w-80">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-ats-gray-400" size={18} />
                  <Input
                    placeholder="Search requirements..."
                    className="pl-10"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                  />
                </div>
                <Button variant="primary" size="sm" className="flex gap-2" onClick={() => navigate('/requirements/add')}>
                  <UserPlus size={16} />
                  Add Requirement
                </Button>
              </div>

              {isLoading ? (
                <div className="h-64 flex items-center justify-center">
                  <p className="text-ats-gray-500">Loading requirements...</p>
                </div>
              ) : isError ? (
                <div className="h-64 flex items-center justify-center">
                  <p className="text-red-500">Failed to load requirements. Please refresh the page or try again.</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Title</TableHead>
                      <TableHead>Department</TableHead>
                      <TableHead>Location</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Headcount</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filtered.length > 0 ? (
                      filtered.map((r) => (
                        <TableRow key={r.id}>
                          <TableCell className="font-medium">{r.title}</TableCell>
                          <TableCell>{r.department || '—'}</TableCell>
                          <TableCell>{r.location || '—'}</TableCell>
                          <TableCell>
                            <span className={`rounded-full px-2 py-1 text-xs capitalize ${statusColors[r.status] ?? 'bg-gray-100 text-gray-700'}`}>
                              {r.status.replace('_', ' ')}
                            </span>
                          </TableCell>
                          <TableCell>{r.headcount}</TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={5} className="text-center py-8 text-gray-500">
                          {searchTerm || clientIdFilter ? 'No requirements found matching your filters.' : 'No requirements yet.'}
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

export default Requirements;
