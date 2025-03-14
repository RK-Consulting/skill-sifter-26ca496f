
import React from 'react';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { Search, Filter, UserPlus, Download, ChevronRight } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';

const Candidates = () => {
  const candidates = [
    { id: 1, name: 'Sarah Wilson', role: 'Senior UI Designer', location: 'New York', status: 'Screening', date: '2 days ago' },
    { id: 2, name: 'John Doe', role: 'Software Engineer', location: 'San Francisco', status: 'Interview', date: '3 days ago' },
    { id: 3, name: 'Emma Thompson', role: 'Product Manager', location: 'Boston', status: 'Offer', date: '1 week ago' },
    { id: 4, name: 'Michael Brown', role: 'Data Scientist', location: 'Austin', status: 'Rejected', date: '1 week ago' },
    { id: 5, name: 'Jessica Lee', role: 'Frontend Developer', location: 'Chicago', status: 'Screening', date: '2 weeks ago' },
  ];

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
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
                  <Input placeholder="Search candidates..." className="pl-10" />
                </div>
                
                <div className="flex gap-3">
                  <Button variant="outline" size="sm" className="flex gap-2">
                    <Filter size={16} />
                    Filter
                  </Button>
                  <Button variant="primary" size="sm" className="flex gap-2">
                    <UserPlus size={16} />
                    Add Candidate
                  </Button>
                </div>
              </div>

              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Role</TableHead>
                    <TableHead>Location</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Applied</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {candidates.map((candidate) => (
                    <TableRow key={candidate.id}>
                      <TableCell className="font-medium">{candidate.name}</TableCell>
                      <TableCell>{candidate.role}</TableCell>
                      <TableCell>{candidate.location}</TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
                          ${candidate.status === 'Screening' ? 'bg-blue-100 text-blue-800' : ''}
                          ${candidate.status === 'Interview' ? 'bg-yellow-100 text-yellow-800' : ''}
                          ${candidate.status === 'Offer' ? 'bg-green-100 text-green-800' : ''}
                          ${candidate.status === 'Rejected' ? 'bg-red-100 text-red-800' : ''}
                        `}>
                          {candidate.status}
                        </span>
                      </TableCell>
                      <TableCell>{candidate.date}</TableCell>
                      <TableCell className="text-right">
                        <Button variant="ghost" size="sm">
                          <ChevronRight size={16} />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </Container>
      </main>
    </div>
  );
};

export default Candidates;
