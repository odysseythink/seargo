import type { ReactNode } from 'react';

interface GeneralTabProps {
  settings: Record<string, any>;
  locked: string[];
  onChange: (key: string, value: any) => void;
}

export default function GeneralTab({ settings, locked, onChange }: GeneralTabProps) {
  const isLocked = (k: string) => locked.includes(k);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
      <Field label="Theme" locked={isLocked('theme')}>
        <select value={settings.theme || 'auto'} onChange={e => onChange('theme', e.target.value)}
          disabled={isLocked('theme')}
          style={selectStyle}>
          {['auto','light','dark','black'].map(t => <option key={t} value={t}>{t.charAt(0).toUpperCase() + t.slice(1)}</option>)}
        </select>
      </Field>
      <Field label="Language" locked={isLocked('language')}>
        <select value={settings.language || 'en'} onChange={e => onChange('language', e.target.value)}
          disabled={isLocked('language')} style={selectStyle}>
          <option value="en">English</option>
          <option value="zh-CN">简体中文</option>
        </select>
      </Field>
      <Field label="SafeSearch" locked={isLocked('safesearch')}>
        <select value={String(settings.safesearch ?? 1)} onChange={e => onChange('safesearch', parseInt(e.target.value))}
          disabled={isLocked('safesearch')} style={selectStyle}>
          <option value="0">Off</option>
          <option value="1">Moderate</option>
          <option value="2">Strict</option>
        </select>
      </Field>
      <Field label="Autocomplete" locked={isLocked('autocomplete')}>
        <select value={settings.autocomplete || ''} onChange={e => onChange('autocomplete', e.target.value)}
          disabled={isLocked('autocomplete')} style={selectStyle}>
          <option value="">Disabled</option>
          <option value="google">Google</option>
          <option value="duckduckgo">DuckDuckGo</option>
          <option value="bing">Bing</option>
        </select>
      </Field>
      <Field label="HTTP Method" locked={isLocked('method')}>
        <select value={settings.method || 'POST'} onChange={e => onChange('method', e.target.value)}
          disabled={isLocked('method')} style={selectStyle}>
          <option value="POST">POST</option>
          <option value="GET">GET</option>
        </select>
      </Field>
    </div>
  );
}

function Field({ label, locked, children }: { label: string; locked: boolean; children: ReactNode }) {
  return (
    <label style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid var(--color-result-border)', opacity: locked ? 0.5 : 1 }}>
      <span style={{ fontWeight: 500, color: 'var(--color-base-font)' }}>
        {label}
        {locked && <span style={{ marginLeft: '0.5rem', fontSize: '0.75rem', color: 'var(--color-engine-warning)' }}> 🔒</span>}
      </span>
      {children}
    </label>
  );
}

const selectStyle: React.CSSProperties = {
  padding: '0.3rem 0.5rem',
  backgroundColor: 'var(--color-search-background)',
  color: 'var(--color-base-font)',
  border: '1px solid var(--color-search-border)',
  borderRadius: '0.25rem',
  minWidth: '8rem',
};
