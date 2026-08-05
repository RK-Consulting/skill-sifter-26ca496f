import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import { Building2, User, Phone, Mail, Calendar, ArrowLeft, Pencil } from 'lucide-react';
import Button from '@/components/ui-custom/Button';
import { toast } from 'sonner';
import { businessDevService } from '@/services/api';
import { Skeleton } from '@/components/ui/skeleton';

interface BusinessDev {
  id: number;
  clientName: string;
  partnerName?: string;
  contactPerson: string;
  contactNumber?: string;
  contactEmail: string;
  createdAt: string;
}

const BusinessDevDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const { data: contact, isLoading, isError } = useQuery({
    queryKey: ['businessDev', id],
    queryFn: async () => {
      try {
        const response = await businessDevService.getBusinessDevById(Number(id));
        if (response.data && response.data.data) {
          return response.data.data as BusinessDev;
        }
        throw new Error('Invalid business dev data format');
      } catch (error) {
        console.error('Error fetching business dev details:', error);
        toast.error('Failed to load client details');
        throw error;
      }
    },
  });

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="pt-24 pb-16">
        <Container>
          <Button
            variant="ghost"
            className="mb-6 flex items-center gap-2"
            onClick={() => navigate('/business-dev')}
          >
            <ArrowLeft size={16} />
            Back to Business Development
          </Button>

          {isLoading && (
            <div className="space-y-4">
              <Skeleton className="h-8 w-1/3" />
              <Skeleton className="h-32 w-full" />
            </div>
          )}

          {isError && (
            <Card>
              <CardContent className="p-6 text-center text-gray-500">
                Could not load client details.
              </CardContent>
            </Card>
          )}

          {contact && (
            <>
              <div className="flex items-start justify-between mb-6">
                <div>
                  <h1 className="text-3xl font-bold">{contact.clientName}</h1>
                  {contact.partnerName && <p className="text-gray-500 mt-1">{contact.partnerName}</p>}
                </div>
                <Button
                  variant="secondary"
                  className="flex items-center gap-2"
                  onClick={() => navigate(`/business-dev/add?editId=${contact.id}`)}
                >
                  <Pencil size={16} />
                  Edit
                </Button>
              </div>

              <Card>
                <CardHeader>
                  <CardTitle>Contact Details</CardTitle>
                </CardHeader>
                <CardContent className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="flex items-center gap-2">
                    <User size={16} className="text-gray-400" />
                    <span>{contact.contactPerson}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Building2 size={16} className="text-gray-400" />
                    <span>{contact.partnerName || 'No partner listed'}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Mail size={16} className="text-gray-400" />
                    <span>{contact.contactEmail}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Phone size={16} className="text-gray-400" />
                    <span>{contact.contactNumber || 'No number listed'}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Calendar size={16} className="text-gray-400" />
                    <span>Added {new Date(contact.createdAt).toLocaleDateString()}</span>
                  </div>
                </CardContent>
              </Card>
            </>
          )}
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default BusinessDevDetails;