import type { MusicResult } from '../../types/search';

interface Props {
  result: MusicResult;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function MusicCard({ result, index: _index, resultsOnNewTab }: Props) {
  const extra = result.extra;
  return (
    <div style={{
      padding: '0.75rem',
      borderBottom: '1px solid var(--color-result-border)',
    }}>
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.75rem' }}>
        <div style={{
          flexShrink: 0,
          width: '2.5rem',
          height: '2.5rem',
          backgroundColor: 'var(--color-result-background)',
          borderRadius: '8px',
          border: '1px solid var(--color-result-border)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: '1.125rem',
          color: 'var(--color-result-link)',
        }}>
          🎵
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <a href={result.url}
            target={resultsOnNewTab ? '_blank' : '_self'}
            rel="noopener noreferrer"
            style={{
              fontSize: '1rem',
              fontWeight: 500,
              color: 'var(--color-result-link)',
              textDecoration: 'none',
              display: 'block',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
            onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
            onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
          >
            {result.title}
          </a>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginTop: '0.25rem', fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>
            {extra?.artist && <span>{extra.artist}</span>}
            {extra?.album && <span>{extra.album}</span>}
            {extra?.duration && <span>{extra.duration}</span>}
          </div>
          {result.content && (
            <p style={{ color: 'var(--color-base-font)', fontSize: '0.875rem', lineHeight: 1.6, margin: '0.25rem 0 0' }}>{result.content}</p>
          )}
        </div>
      </div>
    </div>
  );
}
