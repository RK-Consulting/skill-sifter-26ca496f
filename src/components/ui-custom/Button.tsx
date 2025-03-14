
import React from 'react';
import { cn } from '@/lib/utils';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'link';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
  icon?: React.ReactNode;
  iconPosition?: 'left' | 'right';
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ 
    className, 
    children, 
    variant = 'primary', 
    size = 'md', 
    isLoading = false,
    icon,
    iconPosition = 'left',
    disabled,
    ...props 
  }, ref) => {
    // Button sizes
    const sizeClasses = {
      sm: 'px-3 py-1.5 text-xs',
      md: 'px-4 py-2 text-sm',
      lg: 'px-5 py-2.5 text-base'
    };

    // Button variants
    const variantClasses = {
      primary: 'bg-ats-blue text-white hover:bg-ats-blue/90 shadow-sm',
      secondary: 'bg-ats-gray-100 text-ats-gray-700 hover:bg-ats-gray-200 border border-ats-gray-200',
      outline: 'bg-transparent border border-ats-gray-300 text-ats-gray-700 hover:bg-ats-gray-50',
      ghost: 'bg-transparent text-ats-gray-700 hover:bg-ats-gray-100 hover:text-ats-gray-900',
      link: 'bg-transparent text-ats-blue hover:underline p-0 h-auto shadow-none'
    };

    // Loading state
    const loadingClasses = isLoading ? 'opacity-70 cursor-wait' : '';
    const disabledClasses = disabled ? 'opacity-50 cursor-not-allowed' : '';

    return (
      <button
        className={cn(
          'relative inline-flex items-center justify-center whitespace-nowrap rounded-lg font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ats-blue focus-visible:ring-offset-2 active:scale-[0.98]',
          sizeClasses[size],
          variantClasses[variant],
          loadingClasses,
          disabledClasses,
          variant !== 'link' && 'h-10',
          className
        )}
        disabled={disabled || isLoading}
        ref={ref}
        {...props}
      >
        {isLoading && (
          <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        )}
        
        {!isLoading && icon && iconPosition === 'left' && (
          <span className="mr-2">{icon}</span>
        )}
        
        <span>{children}</span>
        
        {!isLoading && icon && iconPosition === 'right' && (
          <span className="ml-2">{icon}</span>
        )}
      </button>
    );
  }
);

Button.displayName = 'Button';

export default Button;
