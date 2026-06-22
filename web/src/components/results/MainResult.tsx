import type { MainResult as MainResultType } from '../../types/search';

interface Props {
  result: MainResultType;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function MainResult({ result, index: _index, resultsOnNewTab }: Props) {
  const engineColors: Record<string, string> = {
    google: '#ea4335', bing: '#00809d', duckduckgo: '#de5833',
    brave: '#fb542b', wikipedia: '#3366cc', yahoo: '#6001d2',
  };
  const color = engineColors[result.engine?.toLowerCase()] || '#6b7280';

  return (
    <article style={{ padding: '0.75rem 0', borderBottom: '1px solid var(--color-result-border)' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', marginBottom: '0.25rem' }}>
        {result.favicon && (
          <img src={result.favicon} alt=""
            style={{ width: '1rem', height: '1rem' }}
            onError={e => { (e.target as HTMLImageElement).style.display = 'none'; }} />
        )}
        <span style={{
          color: 'var(--color-result-url)',
          fontSize: '0.8rem',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap'
        }}>
          {result.url}
        </span>
      </div>
      <h3 style={{ margin: '0.25rem 0', fontSize: '1rem', fontWeight: 600 }}>
        <a href={result.url}
          target={resultsOnNewTab ? '_blank' : '_self'}
          rel="noreferrer"
          style={{ color: 'var(--color-result-link)', textDecoration: 'none' }}
          onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
          onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
        >
          {result.title}
        </a>
      </h3>
      {result.content && (
        <p style={{
          margin: '0.25rem 0',
          fontSize: '0.85rem',
          color: 'var(--color-base-font)',
          lineHeight: 1.5
        }}>
          {result.content}
        </p>
      )}
      {result.published_at && (
        <time style={{ fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>
          {new Date(result.published_at).toLocaleDateString()}
        </time>
      )}
      <div style={{ marginTop: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <span style={{
          display: 'inline-flex',
          alignItems: 'center',
          padding: '0.125rem 0.5rem',
          borderRadius: '999px',
          fontSize: '0.75rem',
          fontWeight: 500,
          color: '#fff',
          backgroundColor: color,
        }}>
          {result.engine}
        </span>
        {result.score > 0 && (
          <span style={{ fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>
            Score: {result.score.toFixed(2)}
          </span>
        )}
      </div>
    </article>
  );
}
