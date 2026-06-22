export default function InfoPage() {
  return (
    <div style={{ maxWidth: '48rem', margin: '0 auto', padding: '3rem 1.5rem' }}>
      <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '1rem', color: 'var(--color-base-font)' }}>About SearGo</h1>
      <p style={{ color: 'var(--color-base-font)', lineHeight: 1.6, marginBottom: '1rem' }}>
        SearGo is a privacy-focused metasearch engine. It aggregates results from multiple search engines
        while preserving your privacy — no tracking, no profiling, no data collection.
      </p>
      <p style={{ color: 'var(--color-base-font)', lineHeight: 1.6, marginBottom: '1rem' }}>
        Built with Go and React. Inspired by SearXNG.
      </p>
    </div>
  );
}
