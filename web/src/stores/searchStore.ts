import { create } from 'zustand';
import type { SearchState } from '../types/search';
import { api } from '../services/api';

export const useSearchStore = create<SearchState>((set) => ({
  query: '',
  results: [],
  answers: [],
  corrections: [],
  infoboxes: [],
  isLoading: false,
  enginesUsed: [],
  enginesFailed: [],
  responseTimeMs: 0,
  error: null,

  setQuery: (q) => set({ query: q }),

  search: async (req) => {
    set({ isLoading: true, error: null });
    try {
      const resp = await api.search(req);
      set({
        query: resp.data.query,
        results: resp.data.results,
        answers: resp.data.answers || [],
        corrections: resp.data.corrections || [],
        infoboxes: resp.data.infoboxes || [],
        enginesUsed: resp.data.engines_used || [],
        enginesFailed: resp.data.engines_failed || [],
        responseTimeMs: resp.data.response_time_ms || 0,
        isLoading: false,
      });
    } catch (err: any) {
      set({
        isLoading: false,
        error: err.response?.data?.error?.message || err.message || 'Search failed',
      });
    }
  },
}));
