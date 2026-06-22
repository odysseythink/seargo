import { useState, useEffect } from 'react';
import { api } from '../../services/api';

interface EnginesTabProps {
  settings: Record<string, any>;
  locked: string[];
  onChange: (key: string, value: any) => void;
}

export default function EnginesTab({ settings, locked, onChange }: EnginesTabProps) {
  const [engines, setEngines] = useState<any[]>([]);
  const [categories, setCategories] = useState<any[]>([]);

  useEffect(() => {
    api.getEngines().then(r => setEngines((r.data as any).engines || [])).catch(() => {});
    api.getCategories().then(r => setCategories((r.data as any).categories || [])).catch(() => {});
  }, []);

  const disabled = new Set<string>(settings.disabled_engines || []);
  const isLocked = (k: string) => locked.includes(k);

  const toggle = (name: string, cat: string) => {
    if (isLocked('disabled_engines')) return;
    const key = `${name}__${cat}`;
    const next = new Set(disabled);
    if (next.has(key)) next.delete(key); else next.add(key);
    onChange('disabled_engines', Array.from(next));
  };

  const catNames = categories.map((c: any) => typeof c === 'string' ? c : c.name);

  return (
    <div>
      {catNames.map(cat => {
        const catEngines = engines.filter((e: any) => e.categories?.includes(cat));
        if (!catEngines.length) return null;
        return (
          <div key={cat} style={{ marginBottom: '1rem' }}>
            <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '0.5rem', textTransform: 'capitalize', color: 'var(--color-base-font)' }}>{cat}</h3>
            {catEngines.map((eng: any) => {
              const key = `${eng.name}|${cat}`;
              return (
                <label key={key} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', padding: '0.3rem 0' }}>
                  <input type="checkbox" checked={!disabled.has(key)} onChange={() => toggle(eng.name, cat)}
                    disabled={isLocked('disabled_engines')} />
                  <span style={{ color: 'var(--color-base-font)', fontWeight: 500 }}>{eng.name}</span>
                  {eng.shortcut && <code style={{ fontSize: '0.75rem', color: 'var(--color-result-url)' }}>!{eng.shortcut}</code>}
                </label>
              );
            })}
          </div>
        );
      })}
    </div>
  );
}
