export type ThemeName = 'auto' | 'light' | 'dark' | 'black';
export type IconName =
  | 'search' | 'close' | 'settings' | 'information-circle' | 'heart'
  | 'navigate-up' | 'navigate-left' | 'navigate-right' | 'alert'
  | 'exclamation-sign';
export type CSSVariableMap = Record<string, string>;
export interface ThemeVariableSet {
  light: CSSVariableMap;
  dark: CSSVariableMap;
  black: CSSVariableMap;
}
