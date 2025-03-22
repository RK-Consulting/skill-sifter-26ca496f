
import React from 'react';
import Navbar from '@/components/layout/Navbar';
import Dashboard from '@/components/dashboard/Dashboard';
import Footer from '@/components/layout/Footer';

const Index = () => {
  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 flex-grow">
        <Dashboard />
      </main>
      <Footer />
    </div>
  );
};

export default Index;
