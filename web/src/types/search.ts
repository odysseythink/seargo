export interface Result {
  title: string;
  url: string;
  content: string;
  engine: string;
  score: number;
  thumbnail_url?: string;
  published_at?: string;
}

export interface SearchRequest {
  q: string;
  category?: string;
  language?: string;
  safesearch?: number;
  time_range?: string;
  page?: number;
}

export interface SearchResponse {
  query: string;
  category: string;
  results: Result[];
  suggestions: string[];
  total: number;
  page: number;
  page_size: number;
  engines_used: string[];
  engines_failed: string[];
  response_time_ms: number;
}
