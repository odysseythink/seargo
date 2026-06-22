import { useSearchStore } from '../../stores/searchStore';
import { InfoboxPanel } from './InfoboxPanel';

export default function ResultsSidebar() {
  const { infoboxes, suggestions } = useSearchStore();

  if (!infoboxes?.length && !suggestions?.length) return null;

  return (
    <aside style={{
      width: '16rem',
      flexShrink: 0,
      fontSize: '0.85rem',
    }}>
      {infoboxes?.map((ib: any, i: number) => (
        <div key={i} style={{ marginBottom: '0.75rem' }}>
          <InfoboxPanel infobox={ib} />
        </div>
      ))}
      {suggestions?.length > 0 && (
        <div style={{
          padding: '0.75rem',
          backgroundColor: 'var(--color-sidebar-background)',
          border: '1px solid var(--color-sidebar-border)',
          borderRadius: '0.25rem',
          marginBottom: '0.75rem',
        }}>
          <h4 style={{ fontWeight: 600, marginBottom: '0.5rem', color: 'var(--color-base-font)' }}>Suggestions</h4>
          {suggestions.map((s: string, i: number) => (
            <a key={i} href={`/?q=${encodeURIComponent(s)}`}
              style={{ display: 'block', color: 'var(--color-result-link)', marginBottom: '0.25rem', fontSize: '0.85rem' }}>
              {s}
            </a>
          ))}
        </div>
      )}
    </aside>
  );
}
