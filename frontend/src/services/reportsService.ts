import { reportService } from './api';

export interface HiringReportEntry {
  date: string;
  totalInterviews: number;
}

export interface SourceReportEntry {
  source: string;
  count: number;
}

export interface ActivityEntry {
  type: string;
  title: string;
  description: string;
  timestamp: string;
}

export interface PipelineData {
  screening: number;
  interview: number;
  rejected: number;
}

export const reportsService = {
  async getHiringStats(): Promise<HiringReportEntry[]> {
    try {
      const response = await reportService.getHiringReport();
      return response.data?.data || [];
    } catch (error) {
      console.error('Error fetching hiring statistics:', error);
      return [];
    }
  },

  async getSourceStats(): Promise<SourceReportEntry[]> {
    try {
      const response = await reportService.getSourceReport();
      return response.data?.data || [];
    } catch (error) {
      console.error('Error fetching candidate sources:', error);
      return [];
    }
  },

  async getRecentActivity(): Promise<ActivityEntry[]> {
    try {
      const response = await reportService.getRecentActivity();
      return response.data?.data || [];
    } catch (error) {
      console.error('Error fetching recent activity:', error);
      return [];
    }
  },

  async getPipeline(): Promise<PipelineData> {
    try {
      const response = await reportService.getPipeline();
      return response.data?.data || { screening: 0, interview: 0, rejected: 0 };
    } catch (error) {
      console.error('Error fetching pipeline:', error);
      return { screening: 0, interview: 0, rejected: 0 };
    }
  },

  // Keep the old method names for backward compatibility
  async getHiringStatistics(): Promise<HiringReportEntry[]> {
    return this.getHiringStats();
  },

  async getCandidateSources(): Promise<SourceReportEntry[]> {
    return this.getSourceStats();
  },
};