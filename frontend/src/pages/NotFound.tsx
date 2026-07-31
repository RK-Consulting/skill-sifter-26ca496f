
import { useLocation } from "react-router-dom";
import { useEffect } from "react";
import Footer from "@/components/layout/Footer";

const NotFound = () => {
  const location = useLocation();

  useEffect(() => {
    console.error(
      "404 Error: User attempted to access non-existent route:",
      location.pathname
    );
  }, [location.pathname]);

  return (
    <div className="min-h-screen flex flex-col bg-gray-100">
      <div className="flex-grow flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-4xl font-bold text-ats-blue-500 mb-2">404</h1>
          <p className="text-xl text-gray-600 mb-2">Oops! Page not found</p>
          <p className="text-lg font-semibold text-ats-blue-500 mb-1">SkillSifter ATS</p>
          <p className="text-sm text-gray-600 mb-4">R K Consulting</p>
          <a href="/" className="text-blue-500 hover:text-blue-700 underline">
            Return to Home
          </a>
        </div>
      </div>
      <Footer />
    </div>
  );
};

export default NotFound;
