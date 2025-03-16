
import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Index from "./pages/Index";
import Candidates from "./pages/Candidates";
import Jobs from "./pages/Jobs";
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
import ScheduleInterview from "./pages/ScheduleInterview";

const queryClient = new QueryClient();

const App = () => (
  <QueryClientProvider client={queryClient}>
    <TooltipProvider>
      <Toaster />
      <Sonner />
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Index />} />
          <Route path="/candidates" element={<Candidates />} />
          <Route path="/candidates/add" element={<AddCandidate />} />
          <Route path="/jobs" element={<Jobs />} />
          <Route path="/jobs/add" element={<AddJob />} />
          <Route path="/daily-jobs" element={<DailyJobs />} />
          <Route path="/daily-jobs/add" element={<AddDailyJob />} />
          <Route path="/business-dev" element={<BusinessDev />} />
          <Route path="/business-dev/add" element={<AddBusinessDev />} />
          <Route path="/reports" element={<Reports />} />
          <Route path="/interviews" element={<ScheduleInterview />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          {/* ADD ALL CUSTOM ROUTES ABOVE THE CATCH-ALL "*" ROUTE */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </BrowserRouter>
    </TooltipProvider>
  </QueryClientProvider>
);

export default App;
