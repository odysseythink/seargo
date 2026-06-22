import type { NewsResult as NewsResultType } from '../../types/search';

interface Props {
  result: NewsResultType;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function NewsResult({ result, index: _index, resultsOnNewTab }: Props) {
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
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.75rem', color: 'var(--color-result-published-date)', marginBottom: '0.5rem' }}>
        {result.published_at && <span>{result.published_at}</span>}
        {result.engine && <span>· {result.engine}</span>}
      </div>
      {result.content && (
        <p style={{ color: 'var(--color-base-font)', fontSize: '0.875rem', lineHeight: 1.6, margin: 0 }}>{result.content}</p>
      )}
      <div style={{ marginTop: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <span style={{ fontSize: '0.75rem', color: 'var(--color-result-url)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {result.url}
        </span>
      </div>
    </article>
  );
}
