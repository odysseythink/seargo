import type { CodeResult } from '../../types/search';

interface Props {
  result: CodeResult;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function CodeBlock({ result, index: _index, resultsOnNewTab }: Props) {
  const extra = result.extra;
  const lines = extra?.code_lines?.slice(0, 20) ?? [];
  return (
    <div style={{
      padding: '0.75rem 0',
      borderBottom: '1px solid var(--color-result-border)',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem', flexWrap: 'wrap' }}>
        {extra?.filename && (
          <span style={{ fontSize: '0.875rem', fontFamily: 'monospace', color: 'var(--color-result-link)' }}>{extra.filename}</span>
        )}
        {extra?.code_language && (
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
            {extra.code_language}
          </span>
        )}
        {extra?.repository && (
          <a href={extra.repository}
            target={resultsOnNewTab ? '_blank' : '_self'}
            rel="noopener noreferrer"
            style={{
              fontSize: '0.75rem',
              color: 'var(--color-result-published-date)',
              textDecoration: 'none',
              marginLeft: 'auto',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              maxWidth: '200px',
            }}
            onMouseEnter={e => (e.target as HTMLElement).style.color = 'var(--color-base-font)'}
            onMouseLeave={e => (e.target as HTMLElement).style.color = 'var(--color-result-published-date)'}
          >
            {extra.repository}
          </a>
        )}
      </div>
      {lines.length > 0 ? (
        <pre style={{
          backgroundColor: 'var(--color-result-background)',
          borderRadius: '8px',
          padding: '1rem',
          overflowX: 'auto',
          fontSize: '0.875rem',
          fontFamily: 'monospace',
          lineHeight: 1.6,
          border: '1px solid var(--color-result-border)',
        }}>
          {lines.map((l, i) => (
            <code key={i} style={{ color: 'var(--color-base-font)' }}>
              <span style={{ color: 'var(--color-result-published-date)', userSelect: 'none', marginRight: '1rem' }}>{l.line}</span>
              <span>{l.text}</span>
              {'\n'}
            </code>
          ))}
        </pre>
      ) : (
        <p style={{ color: 'var(--color-base-font)', fontSize: '0.875rem', margin: 0 }}>{result.content}</p>
      )}
      <a href={result.url}
        target={resultsOnNewTab ? '_blank' : '_self'}
        rel="noopener noreferrer"
        style={{ marginTop: '0.5rem', display: 'inline-block', fontSize: '0.75rem', color: 'var(--color-result-link)', textDecoration: 'none' }}
        onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
        onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
      >
        View source
      </a>
    </div>
  );
}
