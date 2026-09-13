import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, UserPlus, ChevronRight } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { clientService } from '@/services/api';

interface Client {
  id: number;
  name: string;
  status: string;
  contactEmail?: string;
  contactPhone?: string;
  partnerName?: string;
  contactPerson?: string;
  createdAt: string;
}

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
}

const statusColors: Record<string, string> = {
  prospect: 'bg-gray-100 text-gray-700',
  active: 'bg-green-100 text-green-700',
  inactive: 'bg-red-100 text-red-700',
};

const Clients = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState('');
  const [page, setPage] = useState(1);
  const limit = 20;

  const { data, isLoading, isError } = useQuery({
    queryKey: ['clients', page],
    queryFn: async () => {
      const response = await clientService.getAllClients({ page, limit });
      return response.data as { data: Client[]; pagination: PaginationInfo };
    },
  });

  const clients = data?.data ?? [];
  const pagination = data?.pagination;

  const filtered = React.useMemo(() => {
    const list = data?.data ?? [];
    if (!searchTerm) return list;
    const term = searchTerm.toLowerCase();
    return list.filter(
      (c) =>
        c.name.toLowerCase().includes(term) ||
        (c.contactPerson ?? '').toLowerCase().includes(term) ||
        (c.contactEmail ?? '').toLowerCase().includes(term)
    );
  }, [data, searchTerm]);

  const totalPages = pagination ? Math.max(1, Math.ceil(pagination.total / pagination.limit)) : 1;

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Clients</h1>
            <p className="text-ats-gray-500">Companies you recruit for.</p>
          </div>

          <Card className="mb-8 animate-fade-up">
            <CardContent className="p-6">
              <div className="flex flex-col md:flex-row justify-between gap-4 mb-6">
                <div className="relative w-full md:w-80">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-ats-gray-400" size={18} />
                  <Input
                    placeholder="Search clients..."
                    className="pl-10"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                  />
                </div>
                <Button variant="primary" size="sm" className="flex gap-2" onClick={() => navigate('/clients/add')}>
                  <UserPlus size={16} />
                  Add Client
                </Button>
              </div>

              {isLoading ? (
                <div className="h-64 flex items-center justify-center">
                  <p className="text-ats-gray-500">Loading clients...</p>
                </div>
              ) : isError ? (
                <div className="h-64 flex items-center justify-center">
                  <p className="text-red-500">Failed to load clients. Please refresh the page or try again.</p>
                </div>
              ) : (
                <>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Contact Person</TableHead>
                        <TableHead>Contact Email</TableHead>
                        <TableHead>Contact Phone</TableHead>
                        <TableHead>Partner</TableHead>
                        <TableHead className="text-right">Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filtered.length > 0 ? (
                        filtered.map((c) => (
                          <TableRow key={c.id}>
                            <TableCell className="font-medium">{c.name}</TableCell>
                            <TableCell>
                              <span className={`rounded-full px-2 py-1 text-xs capitalize ${statusColors[c.status] ?? 'bg-gray-100 text-gray-700'}`}>
                                {c.status}
                              </span>
                            </TableCell>
                            <TableCell>{c.contactPerson || '—'}</TableCell>
                            <TableCell>{c.contactEmail || '—'}</TableCell>
                            <TableCell>{c.contactPhone || '—'}</TableCell>
                            <TableCell>{c.partnerName || '—'}</TableCell>
                            <TableCell className="text-right">
                              <Button variant="ghost" size="sm" onClick={() => navigate(`/requirements?clientId=${c.id}`)}>
                                <ChevronRight size={16} />
                              </Button>
                            </TableCell>
                          </TableRow>
                        ))
                      ) : (
                        <TableRow>
                          <TableCell colSpan={7} className="text-center py-8 text-gray-500">
                            {searchTerm ? 'No clients found matching your search.' : 'No clients yet.'}
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>

                  {pagination && pagination.total > pagination.limit && (
                    <div className="flex items-center justify-between mt-4 text-sm text-ats-gray-500">
                      <span>
                        Page {pagination.page} of {totalPages} ({pagination.total} total)
                      </span>
                      <div className="flex gap-2">
                        <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                          Previous
                        </Button>
                        <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
                          Next
                        </Button>
                      </div>
                    </div>
                  )}
                </>
              )}
            </CardContent>
          </Card>
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default Clients;
