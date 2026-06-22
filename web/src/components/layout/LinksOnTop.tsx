import { Link } from 'react-router-dom';

export default function LinksOnTop() {
  return (
    <nav style={{
      display: 'flex',
      justifyContent: 'flex-end',
      gap: '1rem',
      padding: '0.25rem 1rem',
      backgroundColor: 'var(--color-header-background)',
      borderBottom: '1px solid var(--color-header-border)',
      fontSize: '0.8rem',
    }}>
      <Link to="/about" style={{ color: 'var(--color-result-link)' }}>About</Link>
      <Link to="/preferences" style={{ color: 'var(--color-result-link)' }}>Preferences</Link>
    </nav>
  );
}
