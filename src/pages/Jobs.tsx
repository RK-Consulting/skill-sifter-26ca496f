import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, Filter, Plus, Users, Clock, MapPin, Calendar } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toast } from 'sonner';
import { jobService } from '@/services/api';

interface Job {
  id: number;
  title: string;
  department: string;
  location: string;
  type: string;
  applicants: number;
  postedDate: string;
  status: string;
  isRemote?: boolean;
}

const Jobs = () => {
  const navigate = useNavigate();
  const [jobs, setJobs] = useState<Job[]>([]);
  const [filteredJobs, setFilteredJobs] = useState<Job[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [activeTab, setActiveTab] = useState('all');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Load jobs from API using JWT auth
    const fetchJobs = async () => {
      try {
        setIsLoading(true);
        const response = await jobService.getAllJobs();
        
        if (response.data.success) {
          setJobs(response.data.data || []);
        } else {
          // If API returns success: false
          toast.error(response.data.message || "Failed to load jobs");
          // Fall back to localStorage for demo purposes
          const storedJobs = localStorage.getItem('jobs');
          if (storedJobs) {
            setJobs(JSON.parse(storedJobs));
          }
        }
      } catch (error: any) {
        console.error("Error fetching jobs:", error);
        toast.error("Failed to load jobs. Using demo data.");
        
        // Fall back to localStorage for demo purposes
        const storedJobs = localStorage.getItem('jobs');
        if (storedJobs) {
          setJobs(JSON.parse(storedJobs));
        }
      } finally {
        setIsLoading(false);
      }
    };
    
    fetchJobs();
  }, []);

  useEffect(() => {
    let filtered = jobs;
    
    // Filter by status
    if (activeTab !== 'all') {
      filtered = filtered.filter(job => job.status.toLowerCase() === activeTab);
    }
    
    // Filter by search term
    if (searchTerm) {
      filtered = filtered.filter(job => 
        job.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
        job.department.toLowerCase().includes(searchTerm.toLowerCase()) ||
        job.location.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }
    
    setFilteredJobs(filtered);
  }, [searchTerm, activeTab, jobs]);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const handleTabChange = (value: string) => {
    setActiveTab(value);
  };

  const addJob = () => {
    navigate('/jobs/add');
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Jobs</h1>
            <p className="text-ats-gray-500">Create and manage job postings for your organization.</p>
          </div>

          <Card className="mb-8">
            <CardContent className="p-6">
              <div className="flex flex-col md:flex-row justify-between gap-4 mb-6">
                <div className="flex gap-3 items-center">
                  <Tabs defaultValue="all" className="w-full" onValueChange={handleTabChange}>
                    <TabsList>
                      <TabsTrigger value="all">All Jobs</TabsTrigger>
                      <TabsTrigger value="active">Active</TabsTrigger>
                      <TabsTrigger value="draft">Draft</TabsTrigger>
                      <TabsTrigger value="closed">Closed</TabsTrigger>
                    </TabsList>
                  </Tabs>
                </div>
                
                <div className="flex gap-3">
                  <div className="relative w-64">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-ats-gray-400" size={18} />
                    <Input 
                      placeholder="Search jobs..." 
                      className="pl-10" 
                      value={searchTerm}
                      onChange={handleSearch}
                    />
                  </div>
                  <Button 
                    variant="primary" 
                    size="sm" 
                    className="flex gap-2"
                    onClick={addJob}
                  >
                    <Plus size={16} />
                    Post New Job
                  </Button>
                </div>
              </div>

              {isLoading ? (
                <div className="text-center py-12">
                  <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-current border-e-transparent align-[-0.125em] text-ats-blue motion-reduce:animate-[spin_1.5s_linear_infinite]"></div>
                  <p className="mt-4 text-ats-gray-500">Loading jobs...</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                  {filteredJobs.length > 0 ? (
                    filteredJobs.map((job) => (
                      <Card key={job.id} className="overflow-hidden border border-ats-gray-200 hover:border-ats-gray-300 transition-colors">
                        <div className="p-6">
                          <div className="flex justify-between items-start mb-4">
                            <h3 className="text-lg font-semibold">{job.title}</h3>
                            <span className={`text-xs px-2 py-1 rounded-full font-medium
                              ${job.status === 'Active' ? 'bg-green-100 text-green-800' : ''}
                              ${job.status === 'Draft' ? 'bg-gray-100 text-gray-800' : ''}
                              ${job.status === 'Closed' ? 'bg-red-100 text-red-800' : ''}
                            `}>
                              {job.status}
                            </span>
                          </div>
                          
                          <p className="text-ats-gray-500 text-sm mb-4">{job.department}</p>
                          
                          <div className="space-y-2 mb-5">
                            <div className="flex items-center text-sm text-ats-gray-600">
                              <MapPin size={16} className="mr-2" />
                              <span>{job.location}</span>
                            </div>
                            <div className="flex items-center text-sm text-ats-gray-600">
                              <Clock size={16} className="mr-2" />
                              <span>{job.type}</span>
                            </div>
                            <div className="flex items-center text-sm text-ats-gray-600">
                              <Calendar size={16} className="mr-2" />
                              <span>Posted on {job.postedDate}</span>
                            </div>
                          </div>
                          
                          <div className="border-t border-ats-gray-200 pt-4 flex justify-between items-center">
                            <div className="flex items-center text-ats-gray-600">
                              <Users size={16} className="mr-2" />
                              <span className="text-sm">{job.applicants} Applicants</span>
                            </div>
                            <Button variant="ghost" size="sm">View Details</Button>
                          </div>
                        </div>
                      </Card>
                    ))
                  ) : (
                    <div className="col-span-full text-center py-8 text-gray-500">
                      {searchTerm ? 'No jobs found matching your search.' : 'No jobs found.'}
                    </div>
                  )}
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

export default Jobs;
