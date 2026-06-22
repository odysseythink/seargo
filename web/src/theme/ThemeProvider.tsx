import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import type { ThemeName, CSSVariableMap } from './types';

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
    const stored = localStorage.getItem('sxng-theme');
    return (stored as ThemeName) || defaultTheme;
  });
  const [effective, setEffective] = useState<ThemeName>(() => {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    return resolveEffectiveTheme(theme, prefersDark);
  });

  // Listen for system preference changes
  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = (e: MediaQueryListEvent) => {
      if (theme === 'auto') {
        setEffective(e.matches ? 'dark' : 'light');
        applyThemeClass(e.matches ? 'dark' : 'light');
      }
    };
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, [theme]);

  const setTheme = useCallback((t: ThemeName) => {
    setThemeState(t);
    localStorage.setItem('sxng-theme', t);
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const eff = resolveEffectiveTheme(t, prefersDark);
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
