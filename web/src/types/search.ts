export type ResultKind =
  | 'main' | 'image' | 'video' | 'news' | 'paper' | 'code'
  | 'file' | 'map' | 'music' | 'answer' | 'keyvalue' | 'infobox';

export interface BaseSearchResult {
  kind: ResultKind;
  template?: string;
  title: string;
  url: string;
  content?: string;
  engine: string;
  engines?: string[];
  category: string;
  score: number;
  thumbnail_url?: string;
  published_at?: string;
  domain?: string;
  favicon?: string;
}

export interface MainResult extends BaseSearchResult {
  kind: 'main';
  extra?: Record<string, unknown>;
}

export interface ImageResult extends BaseSearchResult {
  kind: 'image';
  extra?: {
    img_src: string;
    thumbnail_src?: string;
    resolution?: string;
    img_format?: string;
    source?: string;
    width?: number;
    height?: number;
    file_size?: string;
  };
}

export interface VideoResult extends BaseSearchResult {
  kind: 'video';
  extra?: {
    thumbnail?: string;
    iframe_src?: string;
    length?: string;
    duration?: string;
    author?: string;
    upload_date?: string;
    view_count?: number;
  };
}

export interface NewsResult extends BaseSearchResult {
  kind: 'news';
  extra?: Record<string, unknown>;
}

export interface PaperResult extends BaseSearchResult {
  kind: 'paper';
  extra?: {
    doi?: string;
    journal?: string;
    authors?: string[];
    publisher?: string;
    type?: string;
    pdf_url?: string;
    html_url?: string;
    issn?: string[];
    isbn?: string[];
    pages?: string;
    tags?: string[];
  };
}

export interface CodeResult extends BaseSearchResult {
  kind: 'code';
  extra?: {
    repository?: string;
    code_language?: string;
    filename?: string;
    code_lines?: { line: number; text: string }[];
    hl_lines?: number[];
  };
}

export interface FileResult extends BaseSearchResult {
  kind: 'file';
  extra?: {
    file_type?: string;
    file_size?: number;
    filename?: string;
    magnet_uri?: string;
    seeders?: number;
    leechers?: number;
  };
}

export interface MapResult extends BaseSearchResult {
  kind: 'map';
  extra?: {
    latitude?: number;
    longitude?: number;
    map_url?: string;
    address?: string;
    bounding_box?: number[];
  };
}

export interface MusicResult extends BaseSearchResult {
  kind: 'music';
  extra?: {
    artist?: string;
    album?: string;
    duration?: string;
  };
}

export interface AnswerResult extends BaseSearchResult {
  kind: 'answer';
  extra?: {
    answer: string;
  };
}

export interface KeyValueResult extends BaseSearchResult {
  kind: 'keyvalue';
  extra?: {
    kv_map?: Record<string, string>;
    caption?: string;
    key_title?: string;
    value_title?: string;
  };
}

export interface InfoboxResult extends BaseSearchResult {
  kind: 'infobox';
  extra?: {
    infobox_id?: string;
    attributes?: { label: string; value: string; url?: string }[];
    urls?: { title: string; url: string }[];
    related_topics?: string[];
    img_src?: string;
    img_alt?: string;
  };
}

export type Result = MainResult | ImageResult | VideoResult | NewsResult
  | PaperResult | CodeResult | FileResult | MapResult
  | MusicResult | AnswerResult | KeyValueResult | InfoboxResult;

export interface SearchRequest {
  q: string;
  category?: string;
  language?: string;
  safesearch?: number;
  time_range?: string;
  page?: number;
}

export interface Answer {
  answer: string;
  url?: string;
  content: string;
  engine?: string;
}

export interface SearchResponse {
  query: string;
  category: string;
  results: Result[];
  suggestions: string[];
  answers: Answer[];
  corrections: string[];
  infoboxes: Infobox[];
  total: number;
  page: number;
  page_size: number;
  engines_used: string[];
  engines_failed: string[];
  response_time_ms: number;
  engine_data?: Record<string, unknown>;
}

export interface Infobox {
  title: string;
  url?: string;
  content?: string;
  engine?: string;
  engines?: string[];
  img_src?: string;
  urls?: { title: string; url: string }[];
  attributes?: { label: string; value: string; url?: string }[];
  related_topics?: string[];
}

export interface SearchState {
  query: string;
  results: Result[];
  answers: Answer[];
  corrections: string[];
  infoboxes: Infobox[];
  isLoading: boolean;
  enginesUsed: string[];
  enginesFailed: string[];
  responseTimeMs: number;
  error: string | null;
  setQuery: (q: string) => void;
  search: (req: SearchRequest) => Promise<void>;
}

export interface PluginPrefItem {
  id: string;
  name: string;
  description: string;
  active: boolean;
  preference_section: "general" | "ui" | "privacy" | "query";
  examples?: string[];
}

export interface AnswererPrefItem {
  id: string;
  name: string;
  description: string;
  active: boolean;
  keywords: string[];
  examples?: string[];
}

export interface AutocompleteSuggestion {
  label: string;
  value: string;
}

export interface AutocompleteResponse {
  query: string;
  suggestions: AutocompleteSuggestion[];
}

export interface UserPreferences {
  categories: string[];
  language: string;
  locale: string;
  autocomplete: string;
  favicon_resolver: string;
  image_proxy: boolean;
  method: string;
  safesearch: number;
  theme: string;
  results_on_new_tab: boolean;
  doi_resolver: string;
  simple_style: string;
  center_alignment: boolean;
  query_in_title: boolean;
  search_on_category_select: boolean;
  hotkeys: string;
  url_formatting: string;
  disabled_engines: string[];
  enabled_engines: string[];
  disabled_plugins: string[];
  enabled_plugins: string[];
  tokens: string[];
  locked: string[];
}

export interface PreferencesResponse {
  plugins: PluginPrefItem[];
  answerers: AnswererPrefItem[];
  autocomplete: string;
  settings: UserPreferences;
  categories: string[];
  themes: string[];
  locales: { tag: string; name: string }[];
  doi_resolvers: string[];
}

export interface TimeStats {
  p50: number;
  p80: number;
  p95: number;
}

export interface CountStats {
  min: number;
  max: number;
  avg: number;
}

export interface EngineTimeSnapshot {
  total: TimeStats;
  http: TimeStats;
  processing: TimeStats;
}

export interface EngineSnapshot {
  engine: string;
  reliability: number;
  score: number;
  time: EngineTimeSnapshot;
  result_count: CountStats;
  request_count: number;
  success_count: number;
  error_counts: Record<string, number>;
  suspended: boolean;
  last_error_at?: string;
}

export interface StatsEnginesResponse {
  engines: EngineSnapshot[];
}

export interface ErrorEntry {
  engine: string;
  total_errors: number;
  by_class: Record<string, number>;
}

export interface StatsErrorsResponse {
  errors: ErrorEntry[];
}

export interface PreferencesUpdate {
  language?: string;
  locale?: string;
  autocomplete?: string;
  favicon_resolver?: string;
  image_proxy?: boolean;
  method?: string;
  safesearch?: number;
  theme?: string;
  results_on_new_tab?: boolean;
  doi_resolver?: string;
  simple_style?: string;
  center_alignment?: boolean;
  query_in_title?: boolean;
  search_on_category_select?: boolean;
  hotkeys?: string;
  url_formatting?: string;
  disabled_engines?: string[];
  enabled_engines?: string[];
  disabled_plugins?: string[];
  enabled_plugins?: string[];
  tokens?: string[];
  categories?: string[];
  plugins?: Record<string, boolean>;
  answerers?: Record<string, boolean>;
}
