
import React from 'react';
import { ChevronRight } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui-custom/Card';

interface PipelineItemProps {
  label: string;
  count: number;
  icon: React.ReactNode;
}

const PipelineItem = ({ label, count, icon }: PipelineItemProps) => (
  <div className="flex items-center justify-between p-4 rounded-lg border border-ats-gray-200 hover:border-ats-gray-300 transition-colors">
    <div className="flex items-center space-x-3">
      {icon}
      <span className="font-medium">{label}</span>
    </div>
    <div className="flex items-center">
      <span className="text-ats-gray-500 mr-2">{count} candidates</span>
      <ChevronRight className="w-4 h-4 text-ats-gray-400" />
    </div>
  </div>
);

interface PipelineStatusProps {
  items: {
    label: string;
    count: number;
    icon: React.ReactNode;
  }[];
}

const PipelineStatus = ({ items }: PipelineStatusProps) => {
  return (
    <Card className="col-span-full lg:col-span-2">
      <CardHeader className="p-6">
        <CardTitle>Recruitment Pipeline</CardTitle>
      </CardHeader>
      <CardContent className="px-6">
        <div className="space-y-4">
          {items.map((item, index) => (
            <PipelineItem
              key={index}
              label={item.label}
              count={item.count}
              icon={item.icon}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
};

export default PipelineStatus;
