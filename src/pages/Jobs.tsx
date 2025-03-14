
import React from 'react';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { Search, Filter, Plus, Users, Clock, MapPin, Calendar } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const Jobs = () => {
  const jobs = [
    { 
      id: 1,
      title: 'Senior UI Designer',
      department: 'Design',
      location: 'New York',
      type: 'Full-time',
      applicants: 24,
      postedDate: 'Sep 14, 2023',
      status: 'Active'
    },
    { 
      id: 2,
      title: 'Software Engineer',
      department: 'Engineering',
      location: 'San Francisco',
      type: 'Full-time',
      applicants: 38,
      postedDate: 'Sep 10, 2023',
      status: 'Active'
    },
    { 
      id: 3,
      title: 'Product Manager',
      department: 'Product',
      location: 'Remote',
      type: 'Full-time',
      applicants: 12,
      postedDate: 'Sep 5, 2023',
      status: 'Closed'
    },
    { 
      id: 4,
      title: 'Data Scientist',
      department: 'Data',
      location: 'Boston',
      type: 'Contract',
      applicants: 8,
      postedDate: 'Aug 28, 2023',
      status: 'Draft'
    },
  ];

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-10">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Jobs</h1>
            <p className="text-ats-gray-500">Create and manage job postings for your organization.</p>
          </div>

          <Card className="mb-8">
            <CardContent className="p-6">
              <div className="flex flex-col md:flex-row justify-between gap-4 mb-6">
                <div className="flex gap-3 items-center">
                  <Tabs defaultValue="all" className="w-full">
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
                    <Input placeholder="Search jobs..." className="pl-10" />
                  </div>
                  <Button variant="primary" size="sm" className="flex gap-2">
                    <Plus size={16} />
                    Post New Job
                  </Button>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {jobs.map((job) => (
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
                ))}
              </div>
            </CardContent>
          </Card>
        </Container>
      </main>
    </div>
  );
};

export default Jobs;
