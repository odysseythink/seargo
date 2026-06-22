import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export default function LinksOnTop() {
  const { t } = useTranslation();

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
      <Link to="/about" style={{ color: 'var(--color-result-link)' }}>{t('sidebar.about')}</Link>
      <Link to="/privacy" style={{ color: 'var(--color-result-link)' }}>{t('sidebar.privacy')}</Link>
      <Link to="/preferences" style={{ color: 'var(--color-result-link)' }}>{t('sidebar.settings')}</Link>
    </nav>
  );
}
