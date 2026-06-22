import { create } from 'zustand';
import type { SearchState } from '../types/search';
import { api } from '../services/api';

export const useSearchStore = create<SearchState>((set) => ({
  query: '',
  results: [],
  answers: [],
  corrections: [],
  infoboxes: [],
  suggestions: [],
  isLoading: false,
  enginesUsed: [],
  enginesFailed: [],
  responseTimeMs: 0,
  total: 0,
  page: 1,
  pageSize: 10,
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
        suggestions: resp.data.suggestions || [],
        enginesUsed: resp.data.engines_used || [],
        enginesFailed: resp.data.engines_failed || [],
        responseTimeMs: resp.data.response_time_ms || 0,
        total: resp.data.total || 0,
        page: resp.data.page || 1,
        pageSize: resp.data.page_size || 10,
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
