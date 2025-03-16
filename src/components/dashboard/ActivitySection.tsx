
import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui-custom/Card';
import Button from '../ui-custom/Button';

interface ActivityItemProps {
  title: string;
  description: string;
  time: string;
}

const ActivityItem = ({ title, description, time }: ActivityItemProps) => (
  <div className="flex items-start space-x-4 p-4 rounded-lg hover:bg-ats-gray-50 transition-colors">
    <div className="flex-1">
      <h4 className="font-medium mb-1">{title}</h4>
      <p className="text-sm text-ats-gray-500">{description}</p>
    </div>
    <time className="text-xs text-ats-gray-400">{time}</time>
  </div>
);

interface ActivitySectionProps {
  activities: {
    title: string;
    description: string;
    time: string;
  }[];
}

const ActivitySection = ({ activities }: ActivitySectionProps) => {
  return (
    <Card>
      <CardHeader className="p-6 flex flex-row items-center justify-between">
        <CardTitle>Recent Activity</CardTitle>
        <Button variant="ghost" size="sm" className="text-ats-blue">
          View All
        </Button>
      </CardHeader>
      <CardContent className="px-6">
        <div className="space-y-4">
          {activities.map((activity, index) => (
            <ActivityItem
              key={index}
              title={activity.title}
              description={activity.description}
              time={activity.time}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
};

export default ActivitySection;
