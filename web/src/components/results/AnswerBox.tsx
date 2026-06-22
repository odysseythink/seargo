import type { AnswerResult, Answer } from '../../types/search';

interface Props {
  result?: AnswerResult;
  answer?: Answer;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function AnswerBox({ result, answer, index: _index, resultsOnNewTab }: Props) {
  const text = result?.extra?.answer || answer?.answer || result?.title || '';
  const content = result?.content || answer?.content || '';
  const sourceUrl = result?.url || answer?.url;
  const engine = result?.engine || answer?.engine;

  return (
    <div style={{
      padding: '1rem',
      backgroundColor: 'var(--color-answer-background)',
      border: '1px solid var(--color-answer-border)',
      borderRadius: '8px',
      color: 'var(--color-answer-font)',
    }}>
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.75rem' }}>
        <div style={{ flexShrink: 0, marginTop: '0.25rem' }}>
          <svg style={{ width: '1.25rem', height: '1.25rem', color: 'var(--color-answer-border)' }} fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
          </svg>
        </div>
        <div style={{ flex: 1 }}>
          {text && (
            <p style={{ fontSize: '1rem', fontWeight: 500, color: 'var(--color-answer-font)', margin: '0 0 0.25rem' }}>{text}</p>
          )}
          {content && (
            <p style={{ fontSize: '0.875rem', color: 'var(--color-answer-font)', lineHeight: 1.6, margin: 0 }}>{content}</p>
          )}
          <div style={{ marginTop: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            {sourceUrl && (
              <a href={sourceUrl}
                target={resultsOnNewTab ? '_blank' : '_self'}
                rel="noopener noreferrer"
                style={{ fontSize: '0.75rem', color: 'var(--color-result-link)', textDecoration: 'none' }}
                onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
                onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
              >Source</a>
            )}
            {engine && (
              <span style={{ fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>{engine}</span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
