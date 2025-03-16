
import React from 'react';
import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Card, CardContent } from '../ui-custom/Card';

interface StatsCardProps {
  title: string;
  value: string;
  trend: string;
  icon: React.ReactElement;
  trendType: 'up' | 'down';
  positive?: boolean;
}

const StatsCard = ({ title, value, trend, icon, trendType, positive = false }: StatsCardProps) => (
  <Card hover>
    <CardContent className="p-6">
      <div className="flex justify-between items-start mb-4">
        <div className="p-2 rounded-lg bg-ats-gray-100/50">
          {React.cloneElement(icon, { className: "w-5 h-5 text-ats-gray-600" })}
        </div>
        <div className={cn(
          "text-sm font-medium px-2 py-1 rounded-full",
          positive ? "bg-green-100 text-green-700" : "bg-ats-blue/10 text-ats-blue"
        )}>
          {trend}
        </div>
      </div>
      <h3 className="text-2xl font-semibold mb-1">{value}</h3>
      <p className="text-sm text-ats-gray-500">{title}</p>
    </CardContent>
  </Card>
);

interface StatsCardsProps {
  stats: {
    title: string;
    value: string;
    trend: string;
    icon: React.ReactElement;
    trendType: 'up' | 'down';
    link: string;
  }[];
}

const StatsCards = ({ stats }: StatsCardsProps) => {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      {stats.map((stat, index) => (
        <Link key={index} to={stat.link}>
          <StatsCard
            title={stat.title}
            value={stat.value}
            trend={stat.trend}
            icon={stat.icon}
            trendType={stat.trendType}
          />
        </Link>
      ))}
    </div>
  );
};

export default StatsCards;
