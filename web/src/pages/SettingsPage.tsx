import { useState, useEffect, useRef, useCallback } from 'react';
import { fetchPreferences, updatePreferences } from '../api/client';
import GeneralTab from './settings/GeneralTab';
import UITab from './settings/UITab';
import EnginesTab from './settings/EnginesTab';
import CookiesTab from './settings/CookiesTab';

const tabs = [
  { key: 'general', label: 'General' },
  { key: 'ui', label: 'UI' },
  { key: 'privacy', label: 'Privacy' },
  { key: 'engines', label: 'Engines' },
  { key: 'cookies', label: 'Cookies' },
];

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState('general');
  const [settings, setSettings] = useState<Record<string, any>>({});
  const [locked, setLocked] = useState<string[]>([]);
  const [saving, setSaving] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const saveTimer = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => {
    fetchPreferences().then(resp => {
      setSettings(resp.settings || {});
      setLocked(resp.locked || []);
      setLoaded(true);
    }).catch(() => setLoaded(true));
  }, []);

  const debouncedSave = useCallback((updates: Record<string, any>) => {
    setSettings(prev => ({ ...prev, ...updates }));
    setSaving(true);
    clearTimeout(saveTimer.current);
    saveTimer.current = setTimeout(async () => {
      try {
        await updatePreferences(updates);
      } catch {}
      setSaving(false);
    }, 500);
  }, []);

  const handleChange = (key: string, value: any) => {
    debouncedSave({ [key]: value });
  };

  if (!loaded) return <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--color-base-font)' }}>Loading...</div>;

  return (
    <div style={{ maxWidth: '48rem', margin: '0 auto', padding: '1rem' }}>
      <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '1rem', color: 'var(--color-base-font)' }}>Preferences</h1>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: '0.25rem', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
        {tabs.map(tab => (
          <button key={tab.key} onClick={() => setActiveTab(tab.key)} style={{
            padding: '0.4rem 1rem',
            border: '1px solid var(--color-settings-tabs-border)',
            borderRadius: '0.25rem 0.25rem 0 0',
            backgroundColor: activeTab === tab.key ? 'var(--color-settings-tabs-active)' : 'var(--color-settings-tabs-background)',
            color: activeTab === tab.key ? 'var(--color-settings-tabs-active-font)' : 'var(--color-settings-tabs-font)',
            cursor: 'pointer',
            fontSize: '0.9rem',
            borderBottom: activeTab === tab.key ? 'none' : undefined,
          }}>
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div style={{ padding: '1rem', backgroundColor: 'var(--color-result-background)', border: '1px solid var(--color-result-border)', borderRadius: '0 0 0.25rem 0.25rem' }}>
        {activeTab === 'general' && <GeneralTab settings={settings} locked={locked} onChange={handleChange} />}
        {activeTab === 'ui' && <UITab settings={settings} locked={locked} onChange={handleChange} />}
        {activeTab === 'privacy' && <UITab settings={settings} locked={locked} onChange={handleChange} />}
        {activeTab === 'engines' && <EnginesTab settings={settings} locked={locked} onChange={handleChange} />}
        {activeTab === 'cookies' && <CookiesTab />}
      </div>

      {saving && (
        <div style={{ textAlign: 'right', marginTop: '0.5rem', fontSize: '0.75rem', color: 'var(--color-result-url)' }}>
          Saving...
        </div>
      )}
    </div>
  );
}
