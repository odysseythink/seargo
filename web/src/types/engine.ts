export interface Capabilities {
  supports_safe_search: boolean;
  supports_language: boolean;
  supports_time_range: boolean;
  supports_pagination: boolean;
  requires_api_key: boolean;
}

export interface EngineInfo {
  name: string;
  categories: string[];
  capabilities: Capabilities;
  enabled: boolean;
}
