import React, { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { cn } from '@/lib/utils';
import Container from './Container';
import { Search, User, Menu, X, BarChart3, Users, FileText, Briefcase, Calendar, Store, Video, LogOut } from 'lucide-react';
import Button from '../ui-custom/Button';

interface NavbarProps extends React.HTMLAttributes<HTMLElement> {
  transparent?: boolean;
}

const Navbar = React.forwardRef<HTMLElement, NavbarProps>(
  ({ className, transparent = false, ...props }, ref) => {
    const [isScrolled, setIsScrolled] = useState(false);
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [userData, setUserData] = useState<any>(null);
    const location = useLocation();
    const navigate = useNavigate();
    
    // Check if user is logged in
    useEffect(() => {
      const storedUser = localStorage.getItem('user');
      if (storedUser) {
        try {
          const parsedUser = JSON.parse(storedUser);
          setIsLoggedIn(!!parsedUser.isLoggedIn);
          setUserData(parsedUser);
        } catch (e) {
          console.error("Error parsing user data:", e);
          setIsLoggedIn(false);
        }
      } else {
        setIsLoggedIn(false);
      }
    }, [location]);
    
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
    
    const handleLogout = () => {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      setIsLoggedIn(false);
      setUserData(null);
      navigate('/login');
    };
    
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
              <div className="h-10 w-10 rounded-full overflow-hidden">
                <img 
                  src="/lovable-uploads/35d9a32a-9b4d-4be7-a93d-03a036a4ab8a.png" 
                  alt="R K Consulting Logo" 
                  className="h-full w-full object-cover"
                />
              </div>
              <div className="flex flex-col">
                <span className="text-xl font-bold tracking-tight whitespace-nowrap text-ats-blue-500">SkillSifter ATS</span>
                <span className="text-sm text-ats-gray-700">R K Consulting</span>
              </div>
            </Link>
            
            {/* Desktop Navigation */}
            <div className="hidden md:flex items-center space-x-1">
              {isLoggedIn ? (
                <>
                  <Link to="/" className={cn(linkClass, isActive('/') && location.pathname === '/' && activeLinkClass)}>Dashboard</Link>
                  <Link to="/candidates" className={cn(linkClass, isActive('/candidates') && activeLinkClass)}>Candidates</Link>
                  <Link to="/jobs" className={cn(linkClass, isActive('/jobs') && activeLinkClass)}>Jobs</Link>
                  <Link to="/daily-jobs" className={cn(linkClass, isActive('/daily-jobs') && activeLinkClass)}>Daily Tasks</Link>
                  <Link to="/business-dev" className={cn(linkClass, isActive('/business-dev') && activeLinkClass)}>Business Dev</Link>
                  <Link to="/interviews" className={cn(linkClass, isActive('/interviews') && activeLinkClass)}>Interviews</Link>
                  <Link to="/reports" className={cn(linkClass, isActive('/reports') && activeLinkClass)}>Reports</Link>
                </>
              ) : (
                <>
                  <Link to="/login" className={cn(linkClass, isActive('/login') && activeLinkClass)}>Login</Link>
                  <Link to="/register" className={cn(linkClass, isActive('/register') && activeLinkClass)}>Register</Link>
                </>
              )}
            </div>
            
            {/* Right actions */}
            <div className="hidden md:flex items-center space-x-4">
              {isLoggedIn ? (
                <>
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
                  
                  <div className="flex items-center space-x-2">
                    <div className="p-1.5 text-ats-gray-600 hover:text-ats-gray-900 rounded-full hover:bg-ats-gray-100/50 border border-ats-gray-200 cursor-pointer">
                      <User className="w-5 h-5" />
                    </div>
                    <div className="hidden sm:block">
                      <div className="text-sm font-medium">{userData?.username || 'User'}</div>
                      <div className="text-xs text-ats-gray-500">{userData?.email || ''}</div>
                    </div>
                    <button 
                      onClick={handleLogout}
                      className="p-1.5 text-ats-gray-600 hover:text-red-500 rounded-full hover:bg-ats-gray-100/50"
                      title="Logout"
                    >
                      <LogOut className="w-5 h-5" />
                    </button>
                  </div>
                </>
              ) : (
                <Link to="/login">
                  <Button 
                    variant="primary" 
                    size="sm" 
                    className="animate-fade-in"
                  >
                    Login
                  </Button>
                </Link>
              )}
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
                {isLoggedIn ? (
                  <>
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
                    <Link to="/interviews" className={cn(linkClass, isActive('/interviews') && activeLinkClass)}>
                      <div className="flex items-center space-x-2">
                        <Video className="w-5 h-5" />
                        <span>Interviews</span>
                      </div>
                    </Link>
                    <Link to="/reports" className={cn(linkClass, isActive('/reports') && activeLinkClass)}>
                      <div className="flex items-center space-x-2">
                        <FileText className="w-5 h-5" />
                        <span>Reports</span>
                      </div>
                    </Link>
                    <button 
                      onClick={handleLogout}
                      className={cn(linkClass, "text-red-500 hover:text-red-600")}
                    >
                      <div className="flex items-center space-x-2">
                        <LogOut className="w-5 h-5" />
                        <span>Logout</span>
                      </div>
                    </button>
                  </>
                ) : (
                  <>
                    <Link to="/login" className={cn(linkClass, isActive('/login') && activeLinkClass)}>
                      <div className="flex items-center space-x-2">
                        <User className="w-5 h-5" />
                        <span>Login</span>
                      </div>
                    </Link>
                    <Link to="/register" className={cn(linkClass, isActive('/register') && activeLinkClass)}>
                      <div className="flex items-center space-x-2">
                        <Users className="w-5 h-5" />
                        <span>Register</span>
                      </div>
                    </Link>
                  </>
                )}
              </div>
              <div className="mt-4 pt-4 border-t border-ats-gray-200 space-y-4">
                {isLoggedIn && (
                  <Link to="/candidates/add">
                    <Button 
                      variant="primary" 
                      className="w-full justify-center"
                    >
                      Add Candidate
                    </Button>
                  </Link>
                )}
                <div className="text-xs text-ats-gray-500 text-center">
                  <p>R K Consulting - SkillSifter</p>
                  <p>Bangalore</p>
                  <a href="http://www.rkconsulting.co.in" className="text-ats-blue hover:underline" target="_blank" rel="noopener noreferrer">www.rkconsulting.co.in</a>
                  <p>
                    <a href="mailto:harishnagaraju@rkconsulting.co.in" className="text-ats-blue hover:underline">harishnagaraju@rkconsulting.co.in</a>
                  </p>
                </div>
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
