
import { useEffect } from 'react';

/**
 * A hook to add enhanced API debugging features to the application.
 * This will monitor network requests and help diagnose issues.
 */
export const useApiDebug = () => {
  useEffect(() => {
    console.log('API Debug Mode Enabled');
    
    // Function to monitor failed network requests
    const monitorErrors = (event: PromiseRejectionEvent | ErrorEvent) => {
      const error = event instanceof PromiseRejectionEvent ? event.reason : event.error;
      
      if (error && error.isAxiosError) {
        console.group('API Request Failed');
        console.log('URL:', error.config?.url);
        console.log('Method:', error.config?.method?.toUpperCase());
        console.log('Status:', error.response?.status);
        console.log('Status Text:', error.response?.statusText);
        console.log('Headers:', error.config?.headers);
        console.log('Request Data:', error.config?.data);
        console.log('Response Data:', error.response?.data);
        console.groupEnd();
        
        // Provide helpful debugging messages based on status code
        switch (error.response?.status) {
          case 401:
            console.warn('Authentication error - Check your token');
            break;
          case 403:
            console.warn('Permission denied - Check your user role');
            break;
          case 404:
            console.warn('Resource not found - Check the URL path');
            break;
          case 405:
            console.warn('Method not allowed - Check if you\'re using the correct HTTP method (GET, POST, PUT, DELETE)');
            // Log method and URL to help diagnose the issue
            console.warn(`Attempted ${error.config?.method?.toUpperCase()} on ${error.config?.url}`);
            break;
          case 422:
            console.warn('Validation error - Check your request data');
            break;
          case 500:
            console.warn('Server error - Check server logs');
            break;
        }
      }
    };

    // Add event listeners
    window.addEventListener('unhandledrejection', monitorErrors);
    window.addEventListener('error', monitorErrors);

    // Monitor TypeError specifically for array handling errors
    const originalErrorHandler = console.error;
    console.error = function(...args) {
      // Check for specific filter is not a function error
      if (args[0] && typeof args[0] === 'string' && args[0].includes('TypeError') && args[0].includes('filter is not a function')) {
        console.warn('DEBUG: Detected TypeError: filter is not a function. This usually means a component expected an array but received something else.');
        console.trace('Stack trace for filter error:');
      }
      
      // Check for specific map is not a function error
      if (args[0] && typeof args[0] === 'string' && args[0].includes('TypeError') && args[0].includes('map is not a function')) {
        console.warn('DEBUG: Detected TypeError: map is not a function. This usually means a component expected an array but received something else.');
        console.trace('Stack trace for map error:');
      }
      
      originalErrorHandler.apply(console, args);
    };
    
    return () => {
      window.removeEventListener('unhandledrejection', monitorErrors);
      window.removeEventListener('error', monitorErrors);
      console.error = originalErrorHandler;
    };
  }, []);
};

export default useApiDebug;
