import type { MapResult } from '../../types/search';

interface Props {
  result: MapResult;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function MapCard({ result, index: _index, resultsOnNewTab }: Props) {
  const extra = result.extra;
  return (
    <article style={{
      padding: '0.75rem 0',
      borderBottom: '1px solid var(--color-result-border)',
    }}>
      <h3 style={{ margin: '0 0 0.25rem', fontSize: '1rem', fontWeight: 600 }}>
        <a href={result.url}
          target={resultsOnNewTab ? '_blank' : '_self'}
          rel="noopener noreferrer"
          style={{ color: 'var(--color-result-link)', textDecoration: 'none' }}
          onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
          onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
        >
          {result.title}
        </a>
      </h3>
      {extra?.address && (
        <p style={{ fontSize: '0.875rem', color: 'var(--color-base-font)', margin: '0 0 0.25rem' }}>{extra.address}</p>
      )}
      {(extra?.latitude !== undefined || extra?.longitude !== undefined) && (
        <p style={{ fontSize: '0.75rem', color: 'var(--color-result-published-date)', margin: '0 0 0.25rem' }}>
          {extra.latitude !== undefined && `${extra.latitude.toFixed(4)}`}
          {extra.latitude !== undefined && extra.longitude !== undefined && ', '}
          {extra.longitude !== undefined && `${extra.longitude.toFixed(4)}`}
        </p>
      )}
      {result.content && (
        <p style={{ color: 'var(--color-base-font)', fontSize: '0.875rem', lineHeight: 1.6, margin: 0 }}>{result.content}</p>
      )}
      <div style={{ marginTop: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        {extra?.map_url && (
          <a href={extra.map_url}
            target={resultsOnNewTab ? '_blank' : '_self'}
            rel="noopener noreferrer"
            style={{ fontSize: '0.75rem', color: 'var(--color-result-url)', textDecoration: 'none' }}
            onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
            onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
          >View on map</a>
        )}
        {result.engine && (
          <span style={{ fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>{result.engine}</span>
        )}
      </div>
    </article>
  );
}
