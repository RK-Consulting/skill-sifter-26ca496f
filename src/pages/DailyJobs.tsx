
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, Filter, PlusCircle, ChevronRight } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';

interface DailyJob {
  id: number;
  jdNo: number;
  instructions: string;
  assignedUser: number;
  assignedDate: string; // Added for display purposes
}

const DailyJobs = () => {
  const navigate = useNavigate();
  const [dailyJobs, setDailyJobs] = useState<DailyJob[]>([]);
  const [filteredJobs, setFilteredJobs] = useState<DailyJob[]>([]);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    // Load daily jobs from localStorage or use defaults if none exist
    const storedJobs = localStorage.getItem('dailyJobs');
    const defaultJobs = [
      { id: 1, jdNo: 1001, instructions: "Contact all candidates and schedule interviews", assignedUser: 3, assignedDate: "Today" },
      { id: 2, jdNo: 1002, instructions: "Follow up with ClientX for feedback", assignedUser: 2, assignedDate: "Yesterday" },
      { id: 3, jdNo: 1003, instructions: "Prepare job descriptions for new positions", assignedUser: 1, assignedDate: "2 days ago" },
      { id: 4, jdNo: 1005, instructions: "Send offer letters to selected candidates", assignedUser: 3, assignedDate: "3 days ago" },
      { id: 5, jdNo: 1008, instructions: "Update candidate profiles in the database", assignedUser: 1, assignedDate: "1 week ago" },
    ];
    
    if (storedJobs) {
      setDailyJobs(JSON.parse(storedJobs));
    } else {
      setDailyJobs(defaultJobs);
      localStorage.setItem('dailyJobs', JSON.stringify(defaultJobs));
    }
  }, []);

  useEffect(() => {
    if (searchTerm) {
      const filtered = dailyJobs.filter(job => 
        job.instructions.toLowerCase().includes(searchTerm.toLowerCase()) ||
        job.jdNo.toString().includes(searchTerm)
      );
      setFilteredJobs(filtered);
    } else {
      setFilteredJobs(dailyJobs);
    }
  }, [searchTerm, dailyJobs]);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const addDailyJob = () => {
    navigate('/daily-jobs/add');
  };

  const viewJobDetails = (id: number) => {
    // In a real application, this would navigate to a job details page
    console.log(`View daily job ${id}`);
  };

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Daily Job Assignments</h1>
            <p className="text-ats-gray-500">Manage and track daily tasks assigned to team members.</p>
          </div>

          <Card className="mb-8">
            <CardContent className="p-6">
              <div className="flex flex-col md:flex-row justify-between gap-4 mb-6">
                <div className="relative w-full md:w-80">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-ats-gray-400" size={18} />
                  <Input 
                    placeholder="Search assignments..." 
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
                    onClick={addDailyJob}
                  >
                    <PlusCircle size={16} />
                    Add Assignment
                  </Button>
                </div>
              </div>

              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>JD No</TableHead>
                    <TableHead>Instructions</TableHead>
                    <TableHead>Assigned User</TableHead>
                    <TableHead>Assigned Date</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredJobs.length > 0 ? (
                    filteredJobs.map((job) => (
                      <TableRow key={job.id}>
                        <TableCell className="font-medium">{job.jdNo}</TableCell>
                        <TableCell>{job.instructions}</TableCell>
                        <TableCell>User #{job.assignedUser}</TableCell>
                        <TableCell>{job.assignedDate}</TableCell>
                        <TableCell className="text-right">
                          <Button 
                            variant="ghost" 
                            size="sm"
                            onClick={() => viewJobDetails(job.id)}
                          >
                            <ChevronRight size={16} />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-8 text-gray-500">
                        {searchTerm ? 'No assignments found matching your search.' : 'No assignments found.'}
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </Container>
      </main>
    </div>
  );
};

export default DailyJobs;
