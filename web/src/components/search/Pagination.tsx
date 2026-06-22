import { useSearchParams, useNavigate } from 'react-router-dom';
import { useSearchStore } from '../../stores/searchStore';

export default function Pagination() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { total, page, pageSize, results } = useSearchStore();

  if (results.length === 0 && total === 0) return null;

  const actualTotal = total > 0 ? total : results.length;
  const pageSize_ = pageSize > 0 ? pageSize : 10;
  const currentPage = page > 0 ? page : parseInt(searchParams.get('page') || '1');
  const totalPages = Math.max(1, Math.ceil(actualTotal / pageSize_));

  if (totalPages <= 1) return null;

  const maxVisible = 7;
  let start = Math.max(1, currentPage - Math.floor(maxVisible / 2));
  let end = Math.min(totalPages, start + maxVisible - 1);
  if (end - start < maxVisible - 1) {
    start = Math.max(1, end - maxVisible + 1);
  }

  const pages: (number | string)[] = [];
  if (start > 1) pages.push(1);
  if (start > 2) pages.push('...');
  for (let i = start; i <= end; i++) pages.push(i);
  if (end < totalPages - 1) pages.push('...');
  if (end < totalPages) pages.push(totalPages);

  const goTo = (p: number) => {
    const params = new URLSearchParams(searchParams);
    params.set('page', String(p));
    navigate(`/?${params.toString()}`);
  };

  const btnBase: React.CSSProperties = {
    padding: '0.25rem 0.6rem',
    border: '1px solid var(--color-pagination-font)',
    borderRadius: '0.25rem',
    background: 'var(--color-pagination-background)',
    color: 'var(--color-pagination-font)',
    cursor: 'pointer',
    fontSize: '0.85rem',
    fontFamily: 'inherit',
  };

  return (
    <nav style={{ display: 'flex', justifyContent: 'center', gap: '0.25rem', padding: '1rem 0' }}>
      {pages.map((p, i) => {
        if (p === '...') {
          return <span key={i} style={{ ...btnBase, border: 'none', cursor: 'default', opacity: 0.6 }}>...</span>;
        }
        const isActive = p === currentPage;
        return (
          <button
            key={i}
            onClick={() => goTo(p as number)}
            style={{
              ...btnBase,
              backgroundColor: isActive ? 'var(--color-pagination-active)' : 'var(--color-pagination-background)',
              color: isActive ? '#fff' : 'var(--color-pagination-font)',
              fontWeight: isActive ? 700 : 400,
            }}
          >
            {p}
          </button>
        );
      })}
    </nav>
  );
}
