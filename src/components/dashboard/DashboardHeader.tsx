
import React from 'react';
import { BarChart3 } from 'lucide-react';

interface DashboardHeaderProps {
  username: string;
}

const DashboardHeader = ({ username }: DashboardHeaderProps) => {
  return (
    <div className="mb-8">
      <div className="inline-flex items-center px-3 py-1 rounded-full bg-ats-blue/10 text-ats-blue text-sm font-medium mb-4">
        <BarChart3 className="w-4 h-4 mr-2" />
        Dashboard Overview
      </div>
      <h1 className="text-3xl font-semibold tracking-tight">Welcome back, {username}</h1>
      <p className="text-ats-gray-500 mt-2">Here's what's happening with your recruitment pipeline.</p>
    </div>
  );
};

export default DashboardHeader;
