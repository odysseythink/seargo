export default function CookiesTab() {
  const getCookie = () => {
    if (typeof document === 'undefined') return 'No preferences cookie found.';
    const match = document.cookie.split(';').find(c => c.trim().startsWith('preferences='));
    return match?.trim() || 'No preferences cookie found.';
  };

  return (
    <div style={{ padding: '1rem 0' }}>
      <p style={{ color: 'var(--color-base-font)', marginBottom: '0.5rem' }}>
        SearGo uses a cookie to store your preferences. The encoded preferences string:
      </p>
      <pre style={{
        padding: '0.75rem',
        backgroundColor: 'var(--color-result-background)',
        border: '1px solid var(--color-result-border)',
        borderRadius: '0.25rem',
        fontSize: '0.75rem',
        wordBreak: 'break-all',
        color: 'var(--color-base-font)',
        maxHeight: '10rem',
        overflowY: 'auto',
      }}>
        {getCookie()}
      </pre>
      <p style={{ marginTop: '0.5rem', fontSize: '0.8rem', color: 'var(--color-result-url)' }}>
        Copy this string to back up your preferences, or paste to restore.
      </p>
    </div>
  );
}
