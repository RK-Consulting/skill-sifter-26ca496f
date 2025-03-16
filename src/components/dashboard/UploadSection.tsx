
import React from 'react';
import { Upload } from 'lucide-react';
import { Card, CardContent } from '../ui-custom/Card';
import Button from '../ui-custom/Button';

const UploadSection = () => {
  return (
    <Card hover className="col-span-full lg:col-span-1">
      <CardContent className="p-6">
        <div className="flex flex-col items-center text-center">
          <div className="w-12 h-12 rounded-full bg-ats-blue/10 text-ats-blue flex items-center justify-center mb-4">
            <Upload className="w-6 h-6" />
          </div>
          <h3 className="text-lg font-semibold mb-2">Upload Resumes</h3>
          <p className="text-sm text-ats-gray-500 mb-4">
            Batch upload resumes to process multiple candidates at once
          </p>
          <Button variant="primary" className="w-full">
            Upload Files
          </Button>
        </div>
      </CardContent>
    </Card>
  );
};

export default UploadSection;
