import type { VideoResult } from '../../types/search';

interface Props {
  result: VideoResult;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function VideoCard({ result, index: _index, resultsOnNewTab }: Props) {
  const duration = result.extra?.duration || result.extra?.length;
  return (
    <a href={result.url}
      target={resultsOnNewTab ? '_blank' : '_self'}
      rel="noopener noreferrer"
      style={{ display: 'block', cursor: 'pointer', textDecoration: 'none' }}>
      <div style={{
        position: 'relative',
        aspectRatio: '16/9',
        backgroundColor: 'var(--color-result-background)',
        borderRadius: '8px',
        overflow: 'hidden',
        border: '1px solid var(--color-result-border)',
      }}>
        {result.extra?.thumbnail ? (
          <>
            <img src={result.extra.thumbnail} alt={result.title}
              style={{
                width: '100%',
                height: '100%',
                objectFit: 'cover',
                transition: 'transform 0.3s',
              }}
              onMouseEnter={e => (e.target as HTMLElement).style.transform = 'scale(1.05)'}
              onMouseLeave={e => (e.target as HTMLElement).style.transform = 'scale(1)'}
            />
            {duration && (
              <span style={{
                position: 'absolute',
                bottom: '0.5rem',
                right: '0.5rem',
                padding: '0.125rem 0.5rem',
                backgroundColor: 'rgba(0,0,0,0.7)',
                color: '#fff',
                fontSize: '0.75rem',
                borderRadius: '4px',
              }}>
                {duration}
              </span>
            )}
          </>
        ) : (
          <div style={{
            width: '100%',
            height: '100%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--color-result-published-date)',
            fontSize: '0.875rem',
          }}>No thumbnail</div>
        )}
      </div>
      <p style={{
        marginTop: '0.25rem',
        fontSize: '0.875rem',
        color: 'var(--color-base-font)',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
        fontWeight: 500,
      }}>{result.title}</p>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>
        {result.extra?.author && <span>{result.extra.author}</span>}
        {result.extra?.upload_date && <span>{result.extra.upload_date}</span>}
        {result.extra?.view_count !== undefined && (
          <span>{result.extra.view_count.toLocaleString()} views</span>
        )}
      </div>
    </a>
  );
}
