
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

interface BusinessDev {
  id: number;
  clientName: string;
  partnerName: string;
  contactPerson: string;
  contactNumber: string;
  contactEmail: string;
  createdAt: string; // Added for display purposes
}

const BusinessDev = () => {
  const navigate = useNavigate();
  const [businessDevs, setBusinessDevs] = useState<BusinessDev[]>([]);
  const [filteredDevs, setFilteredDevs] = useState<BusinessDev[]>([]);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    // Load business dev records from localStorage or use defaults if none exist
    const storedDevs = localStorage.getItem('businessDevs');
    const defaultDevs = [
      { 
        id: 1, 
        clientName: "TechSolutions Inc", 
        partnerName: "Innovate Partners", 
        contactPerson: "John Smith", 
        contactNumber: "+1 (555) 123-4567", 
        contactEmail: "john.smith@techsolutions.com",
        createdAt: "1 day ago" 
      },
      { 
        id: 2, 
        clientName: "Global Finance Group", 
        partnerName: "Capital Ventures", 
        contactPerson: "Sarah Johnson", 
        contactNumber: "+1 (555) 987-6543", 
        contactEmail: "sjohnson@globalfinance.com",
        createdAt: "3 days ago" 
      },
      { 
        id: 3, 
        clientName: "Healthcare Systems", 
        partnerName: "", 
        contactPerson: "Michael Brown", 
        contactNumber: "+1 (555) 456-7890", 
        contactEmail: "mbrown@healthsystems.com",
        createdAt: "1 week ago" 
      },
      { 
        id: 4, 
        clientName: "Retail Solutions", 
        partnerName: "Shop Partners", 
        contactPerson: "Emily Davis", 
        contactNumber: "+1 (555) 234-5678", 
        contactEmail: "edavis@retailsolutions.com",
        createdAt: "2 weeks ago" 
      },
    ];
    
    if (storedDevs) {
      setBusinessDevs(JSON.parse(storedDevs));
    } else {
      setBusinessDevs(defaultDevs);
      localStorage.setItem('businessDevs', JSON.stringify(defaultDevs));
    }
  }, []);

  useEffect(() => {
    if (searchTerm) {
      const filtered = businessDevs.filter(dev => 
        dev.clientName.toLowerCase().includes(searchTerm.toLowerCase()) ||
        dev.contactPerson.toLowerCase().includes(searchTerm.toLowerCase()) ||
        dev.contactEmail.toLowerCase().includes(searchTerm.toLowerCase())
      );
      setFilteredDevs(filtered);
    } else {
      setFilteredDevs(businessDevs);
    }
  }, [searchTerm, businessDevs]);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const addBusinessDev = () => {
    navigate('/business-dev/add');
  };

  const viewBusinessDevDetails = (id: number) => {
    // In a real application, this would navigate to a business dev details page
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

          <Card className="mb-8">
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
                    filteredDevs.map((dev) => (
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
            </CardContent>
          </Card>
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default BusinessDev;
