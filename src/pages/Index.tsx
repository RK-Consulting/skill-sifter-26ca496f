
import React, { useEffect, useState } from 'react';
import Navbar from '@/components/layout/Navbar';
import Dashboard from '@/components/dashboard/Dashboard';
import Footer from '@/components/layout/Footer';
import { useQuery } from '@tanstack/react-query';

const Index = () => {
  const [username, setUsername] = useState('User');

  useEffect(() => {
    // Get the username from localStorage
    const userStr = localStorage.getItem('user');
    if (userStr) {
      try {
        const userData = JSON.parse(userStr);
        if (userData && userData.username) {
          setUsername(userData.username);
        }
      } catch (error) {
        console.error('Error parsing user data:', error);
      }
    }
  }, []);

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 flex-grow">
        <Dashboard username={username} />
      </main>
      <Footer />
    </div>
  );
};

export default Index;
