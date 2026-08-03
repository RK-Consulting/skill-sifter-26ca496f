import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { useEffect, useState } from "react";
import useApiDebug from "./hooks/useApiDebug";
import Index from "./pages/Index";
import Candidates from "./pages/Candidates";
import Jobs from "./pages/Jobs";
import JobDetails from "./pages/JobDetails";
import Reports from "./pages/Reports";
import NotFound from "./pages/NotFound";
import Login from "./pages/Login";
import Register from "./pages/Register";
import AddCandidate from "./pages/AddCandidate";
import AddJob from "./pages/AddJob";
import DailyJobs from "./pages/DailyJobs";
import AddDailyJob from "./pages/AddDailyJob";
import BusinessDev from "./pages/BusinessDev";
import AddBusinessDev from "./pages/AddBusinessDev";
import Interviews from "./pages/Interviews";
import InterviewDetails from "./pages/InterviewDetails";
import ScheduleInterview from "./pages/ScheduleInterview";

// Protected route component
const ProtectedRoute = ({ children }: { children: JSX.Element }) => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);
  
  useEffect(() => {
    const checkAuth = () => {
      const token = localStorage.getItem('token');
      const user = localStorage.getItem('user');
      const isAuthed = !!(token && user);
      setIsAuthenticated(isAuthed);
      
      console.log('Authentication check:', isAuthed ? 'Authenticated' : 'Not authenticated');
    };
    
    // Check auth immediately
    checkAuth();
    
    // Listen for storage events (for when user logs in/out in another tab)
    window.addEventListener('storage', checkAuth);
    return () => {
      window.removeEventListener('storage', checkAuth);
    };
  }, []);
  
  // Show loading or nothing while checking authentication
  if (isAuthenticated === null) {
    return <div className="flex justify-center items-center h-screen">
      <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-ats-blue-500"></div>
    </div>;
  }
  
  return isAuthenticated ? children : <Navigate to="/login" replace />;
};

// Create a QueryClient with retry settings
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1, // Only retry once
      retryDelay: 1000, // Wait 1 second before retry
      staleTime: 1000 * 60 * 5, // 5 minutes
    },
  },
});

// App component with API debugging
const App = () => {
  // Enable API debugging
  useApiDebug();

  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <Toaster />
        <Sonner />
        <BrowserRouter>
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            
            {/* Protected Routes */}
            <Route path="/" element={<ProtectedRoute><Index /></ProtectedRoute>} />
            <Route path="/candidates" element={<ProtectedRoute><Candidates /></ProtectedRoute>} />
            <Route path="/candidates/add" element={<ProtectedRoute><AddCandidate /></ProtectedRoute>} />
            <Route path="/jobs" element={<ProtectedRoute><Jobs /></ProtectedRoute>} />
            <Route path="/jobs/:id" element={<ProtectedRoute><JobDetails /></ProtectedRoute>} />
            <Route path="/jobs/add" element={<ProtectedRoute><AddJob /></ProtectedRoute>} />
            <Route path="/daily-jobs" element={<ProtectedRoute><DailyJobs /></ProtectedRoute>} />
            <Route path="/daily-jobs/add" element={<ProtectedRoute><AddDailyJob /></ProtectedRoute>} />
            <Route path="/business-dev" element={<ProtectedRoute><BusinessDev /></ProtectedRoute>} />
            <Route path="/business-dev/add" element={<ProtectedRoute><AddBusinessDev /></ProtectedRoute>} />
            <Route path="/interviews" element={<ProtectedRoute><Interviews /></ProtectedRoute>} />
            <Route path="/interviews/:id" element={<ProtectedRoute><InterviewDetails /></ProtectedRoute>} />
            <Route path="/interviews/schedule" element={<ProtectedRoute><ScheduleInterview /></ProtectedRoute>} />
            <Route path="/reports" element={<ProtectedRoute><Reports /></ProtectedRoute>} />
            
            {/* Catch-all route */}
            <Route path="*" element={<NotFound />} />
          </Routes>
        </BrowserRouter>
      </TooltipProvider>
    </QueryClientProvider>
  );
};

export default App;