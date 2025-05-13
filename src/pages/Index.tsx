
import React, { useEffect, useState } from 'react';
import Navbar from '@/components/layout/Navbar';
import Dashboard from '@/components/dashboard/Dashboard';
import Footer from '@/components/layout/Footer';

const Index = () => {
  const [username, setUsername] = useState('User');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Get the username from localStorage
    const userStr = localStorage.getItem('user');
    if (userStr) {
      try {
        const userData = JSON.parse(userStr);
        if (userData && userData.username) {
          setUsername(userData.username);
          console.log('Username set to:', userData.username);
        }
      } catch (error) {
        console.error('Error parsing user data:', error);
      }
    }
    
    // Set loading to false after checking user data
    setIsLoading(false);
    
    // Log for debugging
    console.log('Index page loaded, username:', username);
  }, []);

  if (isLoading) {
    return (
      <div className="min-h-screen flex justify-center items-center">
        <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-ats-blue-500"></div>
      </div>
    );
  }

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
