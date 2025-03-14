
import React from 'react';
import { cn } from '@/lib/utils';

interface ContainerProps extends React.HTMLAttributes<HTMLDivElement> {
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  centered?: boolean;
  as?: React.ElementType;
}

const Container = React.forwardRef<HTMLDivElement, ContainerProps>(
  ({ className, size = 'lg', centered = true, as = 'div', children, ...props }, ref) => {
    const Component = as;
    
    // Size classes
    const sizeClasses = {
      sm: 'max-w-3xl',
      md: 'max-w-4xl',
      lg: 'max-w-6xl',
      xl: 'max-w-7xl',
      full: 'max-w-full'
    };
    
    // Centered
    const centeredClasses = centered ? 'mx-auto' : '';
    
    return (
      <Component
        ref={ref}
        className={cn(
          'w-full px-4 sm:px-6 lg:px-8',
          sizeClasses[size],
          centeredClasses,
          className
        )}
        {...props}
      >
        {children}
      </Component>
    );
  }
);

Container.displayName = 'Container';

export default Container;
