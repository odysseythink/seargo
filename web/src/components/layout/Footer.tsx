export default function Footer() {
  return (
    <footer style={{
      backgroundColor: 'var(--color-footer-background)',
      borderTop: '1px solid var(--color-footer-border)',
      padding: '1rem',
      textAlign: 'center',
      fontSize: '0.8rem',
      color: 'var(--color-base-font)',
      opacity: 0.7,
    }}>
      <p>
        Powered by <strong>SearGo</strong>
      </p>
    </footer>
  );
}
