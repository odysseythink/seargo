import type { FileResult } from '../../types/search';

interface Props {
  result: FileResult;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function FileRow({ result, index: _index, resultsOnNewTab }: Props) {
  const extra = result.extra;
  const fileSize = extra?.file_size !== undefined
    ? extra.file_size > 1_000_000_000
      ? `${(extra.file_size / 1_000_000_000).toFixed(1)} GB`
      : extra.file_size > 1_000_000
        ? `${(extra.file_size / 1_000_000).toFixed(1)} MB`
        : extra.file_size > 1_000
          ? `${(extra.file_size / 1_000).toFixed(1)} KB`
          : `${extra.file_size} B`
    : null;

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
          📄
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
            {extra?.filename || result.title}
          </a>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginTop: '0.25rem', fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>
            {extra?.file_type && <span>{extra.file_type.toUpperCase()}</span>}
            {fileSize && <span>{fileSize}</span>}
            {extra?.seeders !== undefined && <span>Seeders: {extra.seeders}</span>}
            {extra?.leechers !== undefined && <span>Leechers: {extra.leechers}</span>}
          </div>
          {extra?.magnet_uri && (
            <a href={extra.magnet_uri}
              style={{ marginTop: '0.25rem', display: 'inline-block', fontSize: '0.75rem', color: 'var(--color-result-url)', textDecoration: 'none', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '100%' }}
              onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
              onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
            >
              Magnet link
            </a>
          )}
        </div>
      </div>
    </div>
  );
}
