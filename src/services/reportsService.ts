
import { reportService } from './api';

export interface HiringReportEntry {
  date: string;
  totalInterviews: number;
}

export interface SourceReportEntry {
  source: string;
  count: number;
}

export const reportsService = {
  async getHiringStatistics(): Promise<HiringReportEntry[]> {
    try {
      const response = await reportService.getHiringReport();
      return response.data?.data || [];
    } catch (error) {
      console.error('Error fetching hiring statistics:', error);
      return [];
    }
  },

  async getCandidateSources(): Promise<SourceReportEntry[]> {
    try {
      const response = await reportService.getSourceReport();
      return response.data?.data || [];
    } catch (error) {
      console.error('Error fetching candidate sources:', error);
      return [];
    }
  },
};
