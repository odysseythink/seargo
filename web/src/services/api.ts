import axios from 'axios';
import type { SearchRequest, SearchResponse, AutocompleteResponse, StatsEnginesResponse, StatsErrorsResponse } from '../types/search';
import type { EngineInfo } from '../types/engine';
import type { Config } from '../types/config';

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
});

export const api = {
  search: (params: SearchRequest) =>
    client.get<SearchResponse>('/search', { params }),

  getEngines: () =>
    client.get<{ engines: EngineInfo[] }>('/engines'),

  getCategories: () =>
    client.get<{ categories: string[] }>('/categories'),

  getConfig: () =>
    client.get<Config>('/config'),

  autocomplete: (q: string, backend?: string, format?: string) =>
    client.get<AutocompleteResponse>('/autocomplete', {
      params: { q, backend, format },
    }),
};

export const fetchStatsEngines = async (): Promise<StatsEnginesResponse> => {
  const res = await fetch("/api/stats/engines");
  if (!res.ok) throw new Error(`Failed to fetch stats: ${res.status}`);
  return res.json();
};

export const fetchStatsErrors = async (): Promise<StatsErrorsResponse> => {
  const res = await fetch("/api/stats/errors");
  if (!res.ok) throw new Error(`Failed to fetch stats errors: ${res.status}`);
  return res.json();
};
