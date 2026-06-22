import { describe, it, expect } from 'vitest';
import { resolveEffectiveTheme, blackThemeOverride } from './ThemeProvider';

describe('resolveEffectiveTheme', () => {
  it('returns dark when auto and prefersDark is true', () => {
    expect(resolveEffectiveTheme('auto', true)).toBe('dark');
  });
  it('returns light when auto and prefersDark is false', () => {
    expect(resolveEffectiveTheme('auto', false)).toBe('light');
  });
  it('returns explicit theme regardless of preference', () => {
    expect(resolveEffectiveTheme('light', true)).toBe('light');
    expect(resolveEffectiveTheme('dark', false)).toBe('dark');
    expect(resolveEffectiveTheme('black', true)).toBe('black');
  });
});

describe('blackThemeOverride', () => {
  it('overrides key backgrounds to #000', () => {
    const dark = {
      '--color-base-background': '#222',
      '--color-header-background': '#1a1a1a',
      '--color-footer-background': '#1a1a1a',
      '--color-sidebar-background': '#1a1a1a',
      '--color-result-background': '#1a1a1a',
    };
    const black = blackThemeOverride(dark);
    expect(black['--color-base-background']).toBe('#000');
    expect(black['--color-header-background']).toBe('#000');
    expect(black['--color-result-background']).toBe('#1a1a1a'); // not overridden
  });
});
