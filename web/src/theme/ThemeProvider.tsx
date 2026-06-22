import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import type { ThemeName, CSSVariableMap } from './types';

const VALID_THEMES: ThemeName[] = ['auto', 'light', 'dark', 'black'];

function isValidTheme(t: string): t is ThemeName {
  return (VALID_THEMES as string[]).includes(t);
}

function safeMatchMedia(query: string): { matches: boolean; addEventListener?: (type: string, handler: (e: MediaQueryListEvent) => void) => void; removeEventListener?: (type: string, handler: (e: MediaQueryListEvent) => void) => void } {
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia(query);
  }
  return { matches: false };
}

function safeLocalStorageGet(key: string): string | null {
  try {
    if (typeof localStorage !== 'undefined') {
      return localStorage.getItem(key);
    }
  } catch {
    // Ignore localStorage errors (e.g., private mode)
  }
  return null;
}

function safeLocalStorageSet(key: string, value: string): void {
  try {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(key, value);
    }
  } catch {
    // Ignore localStorage errors
  }
}

// --- Helpers (exported for testing) ---

export function resolveEffectiveTheme(theme: ThemeName, prefersDark: boolean): ThemeName {
  if (theme === 'auto') return prefersDark ? 'dark' : 'light';
  return theme;
}

export function blackThemeOverride(darkVariables: CSSVariableMap): CSSVariableMap {
  const overrides: CSSVariableMap = {};
  for (const key of Object.keys(darkVariables)) {
    overrides[key] = darkVariables[key];
  }
  const blackBgKeys = [
    '--color-base-background', '--color-base-background-mobile',
    '--color-header-background', '--color-footer-background',
    '--color-sidebar-background',
    '--color-search-background', '--color-categories-background',
    '--color-autocomplete-background', '--color-toolkit-background',
    '--color-pagination-background', '--color-settings-tabs-background',
  ];
  for (const k of blackBgKeys) {
    overrides[k] = '#000';
  }
  return overrides;
}

function applyThemeClass(effective: ThemeName): void {
  const root = document.documentElement;
  root.classList.remove('theme-auto', 'theme-light', 'theme-dark', 'theme-black');
  root.classList.add(`theme-${effective}`);
}

// --- Context ---

interface ThemeContextValue {
  theme: ThemeName;
  effective: ThemeName;
  setTheme: (t: ThemeName) => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error('useTheme must be used within ThemeProvider');
  return ctx;
}

// --- Provider ---

interface ThemeProviderProps {
  defaultTheme?: ThemeName;
  children: ReactNode;
}

export function ThemeProvider({ defaultTheme = 'auto', children }: ThemeProviderProps) {
  const [theme, setThemeState] = useState<ThemeName>(() => {
    const stored = safeLocalStorageGet('sxng-theme');
    if (stored && isValidTheme(stored)) {
      return stored;
    }
    return defaultTheme;
  });
  const [effective, setEffective] = useState<ThemeName>(() => {
    const mq = safeMatchMedia('(prefers-color-scheme: dark)');
    return resolveEffectiveTheme(theme, mq.matches);
  });

  // Listen for system preference changes
  useEffect(() => {
    const mq = safeMatchMedia('(prefers-color-scheme: dark)');
    const handler = (e: MediaQueryListEvent) => {
      if (theme === 'auto') {
        setEffective(e.matches ? 'dark' : 'light');
        applyThemeClass(e.matches ? 'dark' : 'light');
      }
    };
    mq.addEventListener?.('change', handler);
    return () => mq.removeEventListener?.('change', handler);
  }, [theme]);

  const setTheme = useCallback((t: ThemeName) => {
    setThemeState(t);
    safeLocalStorageSet('sxng-theme', t);
    const mq = safeMatchMedia('(prefers-color-scheme: dark)');
    const eff = resolveEffectiveTheme(t, mq.matches);
    setEffective(eff);
    applyThemeClass(eff);
  }, []);

  // Apply class on mount and theme change
  useEffect(() => {
    applyThemeClass(effective);
  }, [effective]);

  return (
    <ThemeContext.Provider value={{ theme, effective, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}
