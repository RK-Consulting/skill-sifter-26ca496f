import React, { useEffect, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { cn } from '@/lib/utils';
import Container from './Container';
import { User, Menu, X, LogOut } from 'lucide-react';
import Button from '../ui-custom/Button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuLabel,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';

interface NavbarUserData {
  username?: string;
  email?: string;
  role?: string;
  companyName?: string;
  subscriptionStatus?: string;
  planCode?: string;
}

interface NavbarProps extends React.HTMLAttributes<HTMLElement> {
  transparent?: boolean;
}

const Navbar = React.forwardRef<HTMLElement, NavbarProps>(
  ({ className, transparent = false, ...props }, ref) => {
    const [isScrolled, setIsScrolled] = useState(false);
    const [mobile, setMobile] = useState(false);
    const [logged, setLogged] = useState(false);
    const [user, setUser] = useState<NavbarUserData | null>(null);
    const location = useLocation();
    const navigate = useNavigate();

    useEffect(() => {
      const syncAuth = () => {
        const token = localStorage.getItem('token');
        const storedUser = localStorage.getItem('user');

        if (!token || !storedUser) {
          setLogged(false);
          setUser(null);
          return;
        }

        try {
          setUser(JSON.parse(storedUser) as NavbarUserData);
          setLogged(true);
        } catch {
          setLogged(false);
          setUser(null);
        }
      };

      syncAuth();
      window.addEventListener('storage', syncAuth);
      window.addEventListener('skillsifter-auth-changed', syncAuth);
      return () => {
        window.removeEventListener('storage', syncAuth);
        window.removeEventListener('skillsifter-auth-changed', syncAuth);
      };
    }, [location]);

    useEffect(() => {
      const updateScroll = () => setIsScrolled(window.scrollY > 10);
      window.addEventListener('scroll', updateScroll);
      updateScroll();
      return () => window.removeEventListener('scroll', updateScroll);
    }, []);

    const logout = () => {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      setLogged(false);
      setUser(null);
      window.dispatchEvent(new Event('skillsifter-auth-changed'));
      navigate('/login', { replace: true });
    };

    const link =
      'relative px-3 py-2 text-ats-gray-600 hover:text-ats-gray-900 font-medium rounded-lg transition-colors hover:bg-ats-gray-100/50 whitespace-nowrap';
    const active = 'text-ats-blue font-semibold';
    const isActive = (path: string) =>
      path === '/dashboard'
        ? location.pathname === '/dashboard'
        : location.pathname === path || location.pathname.startsWith(`${path}/`);

    const baseNavItems: [string, string][] = [
      ['/dashboard', 'Dashboard'],
      ['/candidates', 'Candidates'],
      ['/clients', 'Clients'],
      ['/requirements', 'Requirements'],
      ['/interviews', 'Interviews'],
      ['/resume-ai', 'Resume AI'],
      ['/reports', 'Reports'],
    ];
    const navItems: [string, string][] =
      user?.role === 'recruiter'
        ? baseNavItems
        : [
            ...baseNavItems.slice(0, 5),
            ['/billing', 'Billing'],
            ...baseNavItems.slice(5),
          ];

    return (
      <nav
        ref={ref}
        className={cn(
          'fixed top-0 left-0 right-0 z-50 transition-all duration-300 py-4',
          isScrolled || !transparent ? 'glass-nav py-3' : 'bg-transparent',
          className,
        )}
        {...props}
      >
        <Container>
          <div className="flex items-center justify-between gap-2">
            <Link
              to={logged ? '/dashboard' : '/'}
              className="flex items-center space-x-2 flex-shrink-0"
            >
              <div className="h-10 w-10 rounded-full overflow-hidden">
                <img
                  src="/lovable-uploads/35d9a32a-9b4d-4be7-a93d-03a036a4ab8a.png"
                  alt="R K Consulting Logo"
                  className="h-full w-full object-cover"
                />
              </div>
              <div className="flex flex-col">
                <span className="text-xl font-bold tracking-tight whitespace-nowrap text-ats-blue-500">
                  SkillSifter ATS
                </span>
                <span className="text-sm text-ats-gray-700 whitespace-nowrap">
                  R K Consulting
                </span>
              </div>
            </Link>

            <div className="hidden md:flex items-center space-x-1">
              {logged ? (
                navItems.map(([path, label]) => (
                  <Link
                    key={path}
                    to={path}
                    className={cn(link, isActive(path) && active)}
                  >
                    {label}
                  </Link>
                ))
              ) : (
                <>
                  <Link to="/login" className={cn(link, isActive('/login') && active)}>
                    Login
                  </Link>
                  <Link to="/register" className={cn(link, isActive('/register') && active)}>
                    Register
                  </Link>
                </>
              )}
            </div>

            <div className="hidden md:flex items-center space-x-4 flex-shrink-0">
              {logged ? (
                <>
                  <Link to="/candidates/add">
                    <Button variant="primary" size="sm">Add Candidate</Button>
                  </Link>
                  <DropdownMenu>
                    <DropdownMenuTrigger className="p-1.5 border border-ats-gray-200 rounded-full hover:bg-ats-gray-100/50 transition-colors">
                      <User className="w-5 h-5" />
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-56">
                      <DropdownMenuLabel>
                        <div className="text-sm font-medium">{user?.username || 'User'}</div>
                        <div className="text-xs text-ats-gray-500 font-normal">{user?.email || ''}</div>
                        <div className="text-xs text-ats-gray-500 font-normal capitalize">
                          {user?.role?.replace('_', ' ') || ''}
                        </div>
                        <div className="text-xs text-ats-gray-500 font-normal">{user?.companyName || ''}</div>
                        <div className="text-xs text-ats-gray-500 font-normal">
                          {user?.planCode ? `${user.planCode} · ${user.subscriptionStatus || ''}` : ''}
                        </div>
                      </DropdownMenuLabel>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={() => navigate('/account')}>
                        Account & Subscription
                      </DropdownMenuItem>
                      {user?.role === 'admin' && (
                        <DropdownMenuItem onClick={() => navigate('/admin/users')}>
                          User Management
                        </DropdownMenuItem>
                      )}
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={logout} className="text-red-500 focus:text-red-500">
                        <LogOut className="w-4 h-4 mr-2" />
                        Logout
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </>
              ) : (
                <Link to="/login">
                  <Button variant="primary" size="sm">Login</Button>
                </Link>
              )}
            </div>

            <button
              className="md:hidden p-2 text-ats-gray-600 flex-shrink-0"
              onClick={() => setMobile(!mobile)}
              aria-label={mobile ? 'Close navigation menu' : 'Open navigation menu'}
              aria-expanded={mobile}
            >
              {mobile ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>

          {mobile && (
            <div className="md:hidden pt-4 pb-2">
              <div className="flex flex-col space-y-2">
                {logged ? (
                  navItems.map(([path, label]) => (
                    <Link
                      key={path}
                      to={path}
                      onClick={() => setMobile(false)}
                      className={cn(link, isActive(path) && active)}
                    >
                      {label}
                    </Link>
                  ))
                ) : (
                  <>
                    <Link to="/login" onClick={() => setMobile(false)} className={link}>Login</Link>
                    <Link to="/register" onClick={() => setMobile(false)} className={link}>Register</Link>
                  </>
                )}
              </div>
            </div>
          )}
        </Container>
      </nav>
    );
  },
);

Navbar.displayName = 'Navbar';
export default Navbar;
