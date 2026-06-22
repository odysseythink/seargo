import { type ReactNode, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import Footer from './Footer';
import LinksOnTop from './LinksOnTop';

interface LayoutShellProps {
  header: ReactNode;
  children: ReactNode;
}

export default function LayoutShell({ header, children }: LayoutShellProps) {
  const location = useLocation();

  useEffect(() => {
    document.body.className = '';
    if (location.pathname === '/' && location.search.includes('q=')) {
      document.body.classList.add('results_endpoint');
    }
    if (location.pathname === '/' && !location.search.includes('q=')) {
      document.body.classList.add('index_endpoint');
    }
  }, [location]);

  return (
    <div style={{
      display: 'grid',
      gridTemplateRows: 'auto 1fr auto',
      minHeight: '100vh',
      backgroundColor: 'var(--color-base-background)',
      color: 'var(--color-base-font)',
    }}>
      <header style={{
        backgroundColor: 'var(--color-header-background)',
        borderBottom: '1px solid var(--color-header-border)',
      }}>
        <LinksOnTop />
        {header}
      </header>
      <main style={{ padding: '1rem' }}>
        {children}
      </main>
      <Footer />
    </div>
  );
}
