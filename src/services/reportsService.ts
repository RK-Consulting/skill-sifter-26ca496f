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

  // Keep the old method names for backward compatibility
  async getHiringStatistics(): Promise<HiringReportEntry[]> {
    return this.getHiringStats();
  },

  async getCandidateSources(): Promise<SourceReportEntry[]> {
    return this.getSourceStats();
  },
};
