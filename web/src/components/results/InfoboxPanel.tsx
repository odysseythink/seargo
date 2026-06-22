import type { InfoboxResult, Infobox } from '../../types/search';

interface Props {
  result?: InfoboxResult;
  infobox?: Infobox;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function InfoboxPanel({ result, infobox, index: _index, resultsOnNewTab }: Props) {
  const title = infobox?.title || result?.title || '';
  const content = infobox?.content || result?.content || '';
  const engine = infobox?.engine || result?.engine || '';
  const imgSrc = infobox?.img_src || result?.extra?.img_src;
  const urls = infobox?.urls || result?.extra?.urls;
  const attributes = infobox?.attributes || result?.extra?.attributes;
  const relatedTopics = infobox?.related_topics || result?.extra?.related_topics;
  const srcUrl = infobox?.url || result?.url;

  return (
    <div style={{
      padding: '1rem',
      backgroundColor: 'var(--color-sidebar-background)',
      border: '1px solid var(--color-sidebar-border)',
      borderRadius: '8px',
    }}>
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: '1rem' }}>
        {imgSrc && (
          <img src={imgSrc} alt={title}
            style={{ width: '5rem', height: '5rem', objectFit: 'cover', borderRadius: '8px', flexShrink: 0 }} />
        )}
        <div style={{ flex: 1, minWidth: 0 }}>
          {title && (
            <h3 style={{ fontSize: '1.125rem', fontWeight: 600, color: 'var(--color-base-font)', margin: '0 0 0.25rem' }}>{title}</h3>
          )}
          {content && (
            <p style={{ fontSize: '0.875rem', color: 'var(--color-base-font)', lineHeight: 1.6, margin: '0 0 0.75rem' }}>{content}</p>
          )}
          {attributes && attributes.length > 0 && (
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'auto 1fr',
              gap: '0.25rem 1rem',
              fontSize: '0.875rem',
              marginBottom: '0.75rem',
            }}>
              {attributes.map((attr, i) => (
                <div key={i} style={{ display: 'contents' }}>
                  <dt style={{ color: 'var(--color-result-published-date)' }}>{attr.label}</dt>
                  <dd style={{ color: 'var(--color-base-font)', margin: 0 }}>
                    {attr.url ? (
                      <a href={attr.url}
                        target={resultsOnNewTab ? '_blank' : '_self'}
                        rel="noopener noreferrer"
                        style={{ color: 'var(--color-result-link)', textDecoration: 'none' }}
                        onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
                        onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
                      >{attr.value}</a>
                    ) : attr.value}
                  </dd>
                </div>
              ))}
            </div>
          )}
          {urls && urls.length > 0 && (
            <div style={{ marginBottom: '0.75rem' }}>
              <p style={{ fontSize: '0.75rem', color: 'var(--color-result-published-date)', marginBottom: '0.25rem', fontWeight: 500 }}>Links</p>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
                {urls.map((u, i) => (
                  <a key={i} href={u.url}
                    target={resultsOnNewTab ? '_blank' : '_self'}
                    rel="noopener noreferrer"
                    style={{ fontSize: '0.75rem', color: 'var(--color-result-link)', textDecoration: 'none' }}
                    onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
                    onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
                  >
                    {u.title}
                  </a>
                ))}
              </div>
            </div>
          )}
          {relatedTopics && relatedTopics.length > 0 && (
            <div>
              <p style={{ fontSize: '0.75rem', color: 'var(--color-result-published-date)', marginBottom: '0.25rem', fontWeight: 500 }}>Related</p>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.25rem' }}>
                {relatedTopics.map((topic, i) => (
                  <span key={i} style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    padding: '0.125rem 0.5rem',
                    borderRadius: '999px',
                    fontSize: '0.75rem',
                    color: 'var(--color-result-link)',
                    backgroundColor: 'var(--color-sidebar-background)',
                    border: '1px solid var(--color-sidebar-border)',
                  }}>
                    {topic}
                  </span>
                ))}
              </div>
            </div>
          )}
          {engine && (
            <div style={{ marginTop: '0.75rem', fontSize: '0.75rem', color: 'var(--color-result-published-date)' }}>{engine}</div>
          )}
          {srcUrl && (
            <a href={srcUrl}
              target={resultsOnNewTab ? '_blank' : '_self'}
              rel="noopener noreferrer"
              style={{ marginTop: '0.25rem', display: 'inline-block', fontSize: '0.75rem', color: 'var(--color-result-url)', textDecoration: 'none' }}
              onMouseEnter={e => (e.target as HTMLElement).style.textDecoration = 'underline'}
              onMouseLeave={e => (e.target as HTMLElement).style.textDecoration = 'none'}
            >
              Learn more
            </a>
          )}
        </div>
      </div>
    </div>
  );
}
