import { create } from 'zustand';
import type { Result, SearchRequest } from '../types/search';
import { api } from '../services/api';

interface SearchState {
  query: string;
  results: Result[];
  isLoading: boolean;
  enginesUsed: string[];
  enginesFailed: string[];
  responseTimeMs: number;
  error: string | null;
  setQuery: (q: string) => void;
  search: (req: SearchRequest) => Promise<void>;
}

export const useSearchStore = create<SearchState>((set) => ({
  query: '',
  results: [],
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
        enginesUsed: resp.data.engines_used,
        enginesFailed: resp.data.engines_failed,
        responseTimeMs: resp.data.response_time_ms,
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
