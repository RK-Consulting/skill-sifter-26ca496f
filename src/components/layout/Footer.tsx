
import React from 'react';
import Container from './Container';

const Footer: React.FC = () => {
  return (
    <footer className="bg-white border-t border-ats-gray-100 py-6">
      <Container>
        <div className="flex flex-col md:flex-row items-center justify-between">
          <div className="flex items-center space-x-3 mb-4 md:mb-0">
            <img 
              src="/lovable-uploads/35d9a32a-9b4d-4be7-a93d-03a036a4ab8a.png" 
              alt="R K Consulting Logo" 
              className="h-10 w-10 rounded-full"
            />
            <div>
              <h3 className="font-bold text-lg">R K Consulting</h3>
              <p className="text-sm text-ats-gray-500">Talent Acquisition Experts</p>
            </div>
          </div>
          
          <div className="text-center md:text-right">
            <p className="text-sm text-ats-gray-500">Bangalore, India</p>
            <a 
              href="http://www.rkconsulting.co.in" 
              className="text-sm text-ats-blue hover:underline block"
              target="_blank"
              rel="noopener noreferrer"
            >
              www.rkconsulting.co.in
            </a>
            <a 
              href="mailto:harishnagaraju@rkconsulting.co.in" 
              className="text-sm text-ats-blue hover:underline block"
            >
              harishnagaraju@rkconsulting.co.in
            </a>
          </div>
        </div>
      </Container>
    </footer>
  );
};

export default Footer;
