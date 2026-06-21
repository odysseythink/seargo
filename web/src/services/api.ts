import axios from 'axios';
import type { SearchRequest, SearchResponse } from '../types/search';
import type { AutocompleteResponse } from '../types/search';
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
