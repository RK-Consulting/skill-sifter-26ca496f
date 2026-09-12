import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { UserPlus } from 'lucide-react';
import Button from '@/components/ui-custom/Button';
import { assignmentV1Service } from '@/services/api';
import { toast } from 'sonner';
import { getErrorMessage } from '@/lib/utils';

interface Assignment {
  id: number;
  candidateId: number;
  requirementId: number;
  status: string;
  ownerUserId: number;
  createdAt: string;
}

// The full ADR 0003 transition matrix. Only forward/terminal moves are
// offered; illegal transitions are rejected server-side regardless, but
// keeping the UI honest about what's actually legal avoids a round trip
// just to find out a move isn't allowed.
const nextStatuses: Record<string, string[]> = {
  draft: ['screening'],
  screening: ['submitted', 'rejected', 'withdrawn'],
  submitted: ['interviewing', 'rejected', 'withdrawn'],
  interviewing: ['offered', 'rejected', 'withdrawn'],
  offered: ['joined', 'rejected', 'withdrawn'],
  joined: [],
  rejected: [],
  withdrawn: [],
};

const statusColors: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  screening: 'bg-blue-100 text-blue-700',
  submitted: 'bg-indigo-100 text-indigo-700',
  interviewing: 'bg-purple-100 text-purple-700',
  offered: 'bg-yellow-100 text-yellow-700',
  joined: 'bg-green-100 text-green-700',
  rejected: 'bg-red-100 text-red-700',
  withdrawn: 'bg-gray-100 text-gray-500',
};

const Assignments = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [transitioning, setTransitioning] = useState<number | null>(null);

  const { data, isLoading, isError } = useQuery({
    queryKey: ['assignments'],
    queryFn: async () => {
      const response = await assignmentV1Service.getAllAssignments();
      return (response.data?.data ?? []) as Assignment[];
    },
  });

  const assignments = data ?? [];

  const handleTransition = async (id: number, status: string) => {
    setTransitioning(id);
    try {
      await assignmentV1Service.transitionAssignment(id, status);
      toast.success(`Moved to ${status.replace('_', ' ')}`);
      await queryClient.invalidateQueries({ queryKey: ['assignments'] });
    } catch (error: unknown) {
      toast.error(getErrorMessage(error, 'Failed to update assignment status'));
    } finally {
      setTransitioning(null);
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8 flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-semibold tracking-tight mb-3">Recruitment Assignments</h1>
              <p className="text-ats-gray-500">Candidates matched to requirements, and their lifecycle status.</p>
            </div>
            <Button variant="primary" size="sm" className="flex gap-2" onClick={() => navigate('/assignments/add')}>
              <UserPlus size={16} />
              New Assignment
            </Button>
          </div>

          <Card className="mb-8 animate-fade-up">
            <CardContent className="p-6">
              {isLoading ? (
                <div className="h-64 flex items-center justify-center">
                  <p className="text-ats-gray-500">Loading assignments...</p>
                </div>
              ) : isError ? (
                <div className="h-64 flex items-center justify-center">
                  <p className="text-red-500">Failed to load assignments. Please refresh the page or try again.</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Candidate ID</TableHead>
                      <TableHead>Requirement ID</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Next step</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {assignments.length > 0 ? (
                      assignments.map((a) => (
                        <TableRow key={a.id}>
                          <TableCell className="font-medium">#{a.candidateId}</TableCell>
                          <TableCell>#{a.requirementId}</TableCell>
                          <TableCell>
                            <span className={`rounded-full px-2 py-1 text-xs capitalize ${statusColors[a.status] ?? 'bg-gray-100 text-gray-700'}`}>
                              {a.status}
                            </span>
                          </TableCell>
                          <TableCell>
                            <div className="flex gap-2 flex-wrap">
                              {(nextStatuses[a.status] ?? []).map((next) => (
                                <Button
                                  key={next}
                                  variant="outline"
                                  size="sm"
                                  disabled={transitioning === a.id}
                                  onClick={() => handleTransition(a.id, next)}
                                >
                                  {transitioning === a.id ? '...' : next.replace('_', ' ')}
                                </Button>
                              ))}
                              {(nextStatuses[a.status] ?? []).length === 0 && (
                                <span className="text-xs text-ats-gray-400">Terminal</span>
                              )}
                            </div>
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={4} className="text-center py-8 text-gray-500">
                          No assignments yet.
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

export default Assignments;
