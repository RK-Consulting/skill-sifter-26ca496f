import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui-custom/Card';
import { Search, Filter, UserPlus, ChevronRight } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Button from '@/components/ui-custom/Button';
import { businessDevService } from '@/services/api';
import { toast } from 'sonner';

interface BusinessDev {
  id: number;
  clientName: string;
  partnerName: string;
  contactPerson: string;
  contactNumber: string;
  contactEmail: string;
  createdAt: string;
}

const BusinessDev = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState('');

  // Safely extract the array from the API's {success, message, data}
  // envelope. This never falls back to fabricated data — an unrecognized
  // response shape is treated the same as "no data", not "fake data",
  // so a user can never mistake placeholder content for real records.
  const safeGetData = (response: unknown): BusinessDev[] => {
    if (!response) return [];

    if (Array.isArray(response)) return response as BusinessDev[];
    if (typeof response === 'object' && response !== null && 'data' in response) {
      const inner = (response as { data: unknown }).data;
      if (Array.isArray(inner)) return inner as BusinessDev[];
      if (typeof inner === 'object' && inner !== null && 'data' in inner) {
        const innerInner = (inner as { data: unknown }).data;
        if (Array.isArray(innerInner)) return innerInner as BusinessDev[];
      }
    }

    console.warn('Expected array data but received:', response);
    return [];
  };

  // Fetch business dev data using React Query
  const { data: rawBusinessDevs, isLoading, isError } = useQuery({
    queryKey: ['businessDevs'],
    queryFn: async () => {
      try {
        const response = await businessDevService.getAllBusinessDevs();
        console.log('BusinessDev API response:', response);
        return response;
      } catch (error) {
        console.error('Error fetching business dev contacts:', error);
        toast.error('Failed to load business contacts');
        throw error;
      }
    },
    staleTime: 60000, // 1 minute
    retry: 1,
    retryDelay: 1000,
  });

  // Extract safe data
  const businessDevs = safeGetData(rawBusinessDevs);

  // Filter business devs based on search term - ensure businessDevs is an array
  const filteredDevs = React.useMemo(() => {
    if (searchTerm) {
      return businessDevs.filter((dev: BusinessDev) => 
        dev.clientName.toLowerCase().includes(searchTerm.toLowerCase()) ||
        dev.contactPerson.toLowerCase().includes(searchTerm.toLowerCase()) ||
        dev.contactEmail.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }
    return businessDevs;
  }, [searchTerm, businessDevs]);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const addBusinessDev = () => {
    navigate('/business-dev/add');
  };

  const viewBusinessDevDetails = (id: number) => {
    console.log(`View business dev ${id}`);
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-3">Business Development</h1>
            <p className="text-ats-gray-500">Manage client relationships and business contacts.</p>
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
                    onClick={addBusinessDev}
                  >
                    <UserPlus size={16} />
                    Add Client
                  </Button>
                </div>
              </div>

              {isLoading ? (
                <div className="h-64 flex items-center justify-center">
                  <p className="text-ats-gray-500">Loading business contacts...</p>
                </div>
              ) : isError ? (
                <div className="h-64 flex items-center justify-center">
                  <p className="text-red-500">Failed to load business contacts. Please refresh the page or try again.</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Client Name</TableHead>
                      <TableHead>Partner</TableHead>
                      <TableHead>Contact Person</TableHead>
                      <TableHead>Contact Info</TableHead>
                      <TableHead>Added</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredDevs.length > 0 ? (
                      filteredDevs.map((dev: BusinessDev) => (
                        <TableRow key={dev.id}>
                          <TableCell className="font-medium">{dev.clientName}</TableCell>
                          <TableCell>{dev.partnerName || '-'}</TableCell>
                          <TableCell>{dev.contactPerson}</TableCell>
                          <TableCell>
                            <div className="text-sm">
                              <div>{dev.contactEmail}</div>
                              <div className="text-ats-gray-500">{dev.contactNumber}</div>
                            </div>
                          </TableCell>
                          <TableCell>{dev.createdAt}</TableCell>
                          <TableCell className="text-right">
                            <Button 
                              variant="ghost" 
                              size="sm"
                              onClick={() => viewBusinessDevDetails(dev.id)}
                            >
                              <ChevronRight size={16} />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                          {searchTerm ? 'No clients found matching your search.' : 'No clients found.'}
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

export default BusinessDev;