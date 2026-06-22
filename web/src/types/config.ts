export interface Config {
  default_language: string;
  default_category: string;
  safe_search: number;
  autocomplete: string;
  max_results: number;
  ui?: {
    default_locale: string;
    rtl: boolean;
  };
}
