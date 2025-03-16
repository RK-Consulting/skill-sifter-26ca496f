
import React, { useState, useEffect } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { cn } from '@/lib/utils';
import Container from './Container';
import { Search, User, Menu, X, BarChart3, Users, FileText, Briefcase, Calendar, Store } from 'lucide-react';
import Button from '../ui-custom/Button';

interface NavbarProps extends React.HTMLAttributes<HTMLElement> {
  transparent?: boolean;
}

const Navbar = React.forwardRef<HTMLElement, NavbarProps>(
  ({ className, transparent = false, ...props }, ref) => {
    const [isScrolled, setIsScrolled] = useState(false);
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
    const location = useLocation();
    
    // Detect scroll position
    useEffect(() => {
      const handleScroll = () => {
        setIsScrolled(window.scrollY > 10);
      };
      
      window.addEventListener('scroll', handleScroll);
      handleScroll(); // Check initial position
      
      return () => {
        window.removeEventListener('scroll', handleScroll);
      };
    }, []);
    
    const navbarClass = cn(
      'fixed top-0 left-0 right-0 z-50 transition-all duration-300 py-4',
      isScrolled || !transparent ? 'glass-nav py-3' : 'bg-transparent',
      className
    );
    
    const linkClass = "relative px-3 py-2 text-ats-gray-600 hover:text-ats-gray-900 font-medium rounded-lg transition-colors hover:bg-ats-gray-100/50";
    const activeLinkClass = "text-ats-blue font-semibold";
    
    const isActive = (path: string) => {
      return location.pathname.startsWith(path);
    };
    
    return (
      <nav ref={ref} className={navbarClass} {...props}>
        <Container>
          <div className="flex items-center justify-between">
            {/* Logo */}
            <Link to="/" className="flex items-center space-x-2">
              <div className="w-8 h-8 rounded-lg bg-ats-blue text-white flex items-center justify-center">
                <BarChart3 className="w-5 h-5" />
              </div>
              <span className="text-xl font-bold tracking-tight">SkillSifter</span>
            </Link>
            
            {/* Desktop Navigation */}
            <div className="hidden md:flex items-center space-x-1">
              <Link to="/" className={cn(linkClass, isActive('/') && location.pathname === '/' && activeLinkClass)}>Dashboard</Link>
              <Link to="/candidates" className={cn(linkClass, isActive('/candidates') && activeLinkClass)}>Candidates</Link>
              <Link to="/jobs" className={cn(linkClass, isActive('/jobs') && activeLinkClass)}>Jobs</Link>
              <Link to="/daily-jobs" className={cn(linkClass, isActive('/daily-jobs') && activeLinkClass)}>Daily Tasks</Link>
              <Link to="/business-dev" className={cn(linkClass, isActive('/business-dev') && activeLinkClass)}>Business Dev</Link>
              <Link to="/reports" className={cn(linkClass, isActive('/reports') && activeLinkClass)}>Reports</Link>
            </div>
            
            {/* Right actions */}
            <div className="hidden md:flex items-center space-x-4">
              <button className="p-2 text-ats-gray-600 hover:text-ats-gray-900 rounded-lg hover:bg-ats-gray-100/50">
                <Search className="w-5 h-5" />
              </button>
              
              <Link to="/candidates/add">
                <Button 
                  variant="primary" 
                  size="sm" 
                  className="animate-fade-in"
                >
                  Add Candidate
                </Button>
              </Link>
              
              <button className="p-1.5 text-ats-gray-600 hover:text-ats-gray-900 rounded-full hover:bg-ats-gray-100/50 border border-ats-gray-200">
                <User className="w-5 h-5" />
              </button>
            </div>
            
            {/* Mobile menu button */}
            <button 
              className="md:hidden p-2 text-ats-gray-600 hover:text-ats-gray-900 rounded-lg hover:bg-ats-gray-100/50"
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            >
              {isMobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
          
          {/* Mobile menu */}
          {isMobileMenuOpen && (
            <div className="md:hidden pt-4 pb-2 animate-fade-in">
              <div className="flex flex-col space-y-2">
                <Link to="/" className={cn(linkClass, isActive('/') && location.pathname === '/' && activeLinkClass)}>
                  <div className="flex items-center space-x-2">
                    <BarChart3 className="w-5 h-5" />
                    <span>Dashboard</span>
                  </div>
                </Link>
                <Link to="/candidates" className={cn(linkClass, isActive('/candidates') && activeLinkClass)}>
                  <div className="flex items-center space-x-2">
                    <Users className="w-5 h-5" />
                    <span>Candidates</span>
                  </div>
                </Link>
                <Link to="/jobs" className={cn(linkClass, isActive('/jobs') && activeLinkClass)}>
                  <div className="flex items-center space-x-2">
                    <Briefcase className="w-5 h-5" />
                    <span>Jobs</span>
                  </div>
                </Link>
                <Link to="/daily-jobs" className={cn(linkClass, isActive('/daily-jobs') && activeLinkClass)}>
                  <div className="flex items-center space-x-2">
                    <Calendar className="w-5 h-5" />
                    <span>Daily Tasks</span>
                  </div>
                </Link>
                <Link to="/business-dev" className={cn(linkClass, isActive('/business-dev') && activeLinkClass)}>
                  <div className="flex items-center space-x-2">
                    <Store className="w-5 h-5" />
                    <span>Business Dev</span>
                  </div>
                </Link>
                <Link to="/reports" className={cn(linkClass, isActive('/reports') && activeLinkClass)}>
                  <div className="flex items-center space-x-2">
                    <FileText className="w-5 h-5" />
                    <span>Reports</span>
                  </div>
                </Link>
              </div>
              <div className="mt-4 pt-4 border-t border-ats-gray-200">
                <Link to="/candidates/add">
                  <Button 
                    variant="primary" 
                    className="w-full justify-center"
                  >
                    Add Candidate
                  </Button>
                </Link>
              </div>
            </div>
          )}
        </Container>
      </nav>
    );
  }
);

Navbar.displayName = 'Navbar';

export default Navbar;
