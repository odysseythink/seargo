import type { ImageResult } from '../../types/search';

interface Props {
  result: ImageResult;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function ImageCard({ result, index: _index, resultsOnNewTab }: Props) {
  const imgSrc = result.extra?.img_src || result.thumbnail_url || '';
  const title = result.title || 'Image';
  return (
    <a href={result.url}
      target={resultsOnNewTab ? '_blank' : '_self'}
      rel="noopener noreferrer"
      style={{ display: 'block', cursor: 'pointer', textDecoration: 'none' }}>
      <div style={{
        aspectRatio: '4/3',
        backgroundColor: 'var(--color-result-background)',
        borderRadius: '8px',
        overflow: 'hidden',
        border: '1px solid var(--color-result-border)',
      }}>
        {imgSrc ? (
          <img src={imgSrc} alt={title}
            style={{
              width: '100%',
              height: '100%',
              objectFit: 'cover',
              transition: 'transform 0.3s',
            }}
            onMouseEnter={e => (e.target as HTMLElement).style.transform = 'scale(1.05)'}
            onMouseLeave={e => (e.target as HTMLElement).style.transform = 'scale(1)'}
          />
        ) : (
          <div style={{
            width: '100%',
            height: '100%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--color-result-published-date)',
            fontSize: '0.875rem',
          }}>No image</div>
        )}
      </div>
      {result.extra?.resolution && (
        <p style={{ marginTop: '0.25rem', fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>
          {result.extra.resolution}
        </p>
      )}
      <p style={{
        marginTop: '0.25rem',
        fontSize: '0.875rem',
        color: 'var(--color-result-url)',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
      }}>{title}</p>
    </a>
  );
}
