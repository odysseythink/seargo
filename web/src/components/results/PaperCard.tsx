import type { PaperResult } from '../../types/search';

interface Props {
  result: PaperResult;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function PaperCard({ result, index: _index, resultsOnNewTab }: Props) {
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
      {extra?.authors && extra.authors.length > 0 && (
        <p style={{ fontSize: '0.875rem', color: 'var(--color-base-font)', margin: '0 0 0.25rem' }}>{extra.authors.join(', ')}</p>
      )}
      {extra?.journal && (
        <p style={{ fontSize: '0.875rem', color: 'var(--color-result-published-date)', fontStyle: 'italic', margin: '0 0 0.25rem' }}>{extra.journal}</p>
      )}
      {result.content && (
        <p style={{ color: 'var(--color-base-font)', fontSize: '0.875rem', lineHeight: 1.6, margin: '0.25rem 0' }}>{result.content}</p>
      )}
      <div style={{ marginTop: '0.5rem', display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: '0.5rem' }}>
        {extra?.doi && (
          <a href={`https://doi.org/${extra.doi}`}
            target={resultsOnNewTab ? '_blank' : '_self'}
            rel="noopener noreferrer"
            style={{ fontSize: '0.75rem', color: 'var(--color-result-link)', textDecoration: 'none' }}
            onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
            onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
          >DOI: {extra.doi}</a>
        )}
        {extra?.publisher && (
          <span style={{ fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>{extra.publisher}</span>
        )}
        {extra?.type && (
          <span style={{
            display: 'inline-flex',
            alignItems: 'center',
            padding: '0.125rem 0.5rem',
            borderRadius: '4px',
            fontSize: '0.75rem',
            fontWeight: 500,
            backgroundColor: 'var(--color-result-background)',
            color: 'var(--color-result-link)',
            border: '1px solid var(--color-result-border)',
          }}>
            {extra.type}
          </span>
        )}
        {result.engine && (
          <span style={{ fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>{result.engine}</span>
        )}
      </div>
    </article>
  );
}
