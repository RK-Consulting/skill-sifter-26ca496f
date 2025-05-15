
import api from './api';

export const reportsService = {
  // Get hiring statistics
  getHiringStats: async () => {
    try {
      const response = await api.get('/reports/hiring');
      return response.data.data || [];
    } catch (error) {
      console.error('Error fetching hiring statistics:', error);
      // Return fallback data when API fails
      return [
        { name: 'Jan', candidates: 4 },
        { name: 'Feb', candidates: 7 },
        { name: 'Mar', candidates: 5 },
        { name: 'Apr', candidates: 10 },
        { name: 'May', candidates: 8 },
        { name: 'Jun', candidates: 12 },
      ];
    }
  },

  // Get candidate source statistics
  getSourceStats: async () => {
    try {
      const response = await api.get('/reports/sources');
      return response.data.data || [];
    } catch (error) {
      console.error('Error fetching candidate sources:', error);
      // Return fallback data when API fails
      return [
        { name: 'LinkedIn', value: 40 },
        { name: 'Referrals', value: 25 },
        { name: 'Job Boards', value: 20 },
        { name: 'Direct', value: 15 },
      ];
    }
  },

  // Get candidates pipeline stats
  getPipelineStats: async () => {
    try {
      const response = await api.get('/reports/pipeline');
      return response.data.data || [];
    } catch (error) {
      console.error('Error fetching pipeline statistics:', error);
      // Return fallback data
      return [];
    }
  },
};

export default reportsService;
