interface UITabProps {
  settings: Record<string, any>;
  locked: string[];
  onChange: (key: string, value: any) => void;
}

export default function UITab({ settings, locked, onChange }: UITabProps) {
  const isLocked = (k: string) => locked.includes(k);
  const toggleField = (key: string, label: string) => (
    <CheckField key={key} label={label} checked={!!settings[key]} locked={isLocked(key)}
      onChange={v => onChange(key, v)} />
  );

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
      {toggleField('results_on_new_tab', 'Open results in new tab')}
      {toggleField('center_alignment', 'Center alignment')}
      {toggleField('query_in_title', 'Show query in title')}
      {toggleField('search_on_category_select', 'Search on category select')}
      {toggleField('image_proxy', 'Image proxy')}
    </div>
  );
}

function CheckField({ label, checked, locked, onChange }: { label: string; checked: boolean; locked: boolean; onChange: (v: boolean) => void }) {
  return (
    <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', padding: '0.4rem 0', opacity: locked ? 0.5 : 1, cursor: locked ? 'not-allowed' : 'pointer' }}>
      <input type="checkbox" checked={checked} onChange={e => onChange(e.target.checked)} disabled={locked} />
      <span style={{ color: 'var(--color-base-font)' }}>{label}</span>
      {locked && <span style={{ fontSize: '0.75rem', color: 'var(--color-engine-warning)' }}> 🔒</span>}
    </label>
  );
}
