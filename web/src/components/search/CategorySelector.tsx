import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { api } from '../../services/api';

export default function CategorySelector() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [categories, setCategories] = useState<string[]>([]);
  const current = searchParams.get('category') || 'general';

  useEffect(() => {
    api.getCategories().then(resp => {
      const cats = resp.data.categories;
      if (Array.isArray(cats)) {
        setCategories(cats);
      }
    }).catch(() => {});
  }, []);

  if (categories.length === 0) return null;

  const handleSelect = (cat: string) => {
    const params = new URLSearchParams(searchParams);
    params.set('category', cat);
    navigate(`/?${params.toString()}`);
  };

  return (
    <div style={{
      display: 'flex',
      gap: '0.25rem',
      flexWrap: 'wrap',
      justifyContent: 'center',
      padding: '0.5rem 1rem',
    }}>
      {categories.map(cat => (
        <button
          key={cat}
          onClick={() => handleSelect(cat)}
          style={{
            padding: '0.25rem 0.75rem',
            borderRadius: '0.25rem',
            border: '1px solid var(--color-categories-border)',
            backgroundColor: cat === current
              ? 'var(--color-categories-selected)'
              : 'var(--color-categories-background)',
            color: 'var(--color-categories-font)',
            fontSize: '0.8rem',
            cursor: 'pointer',
            fontWeight: cat === current ? 600 : 400,
          }}
        >
          {cat}
        </button>
      ))}
    </div>
  );
}
