import { useState, useCallback } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useSearchStore } from '../../stores/searchStore';
import AutocompleteDropdown from './AutocompleteDropdown';
import CategorySelector from './CategorySelector';

export default function SearchHeader() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { search, isLoading } = useSearchStore();
  const [query, setQuery] = useState(searchParams.get('q') || '');
  const hasQuery = searchParams.has('q');

  const handleSubmit = useCallback((e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;
    const category = searchParams.get('category') || 'general';
    navigate(`/?q=${encodeURIComponent(query.trim())}&category=${category}`);
    search({ q: query.trim(), category });
  }, [query, searchParams, navigate, search]);

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
    }}>
      {/* Logo */}
      {!hasQuery ? (
        <div style={{ padding: '4rem 1rem 2rem' }}>
          <img src="/searxng.png" alt="SearGo" style={{ height: '4rem' }} />
        </div>
      ) : (
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', padding: '0.75rem 1rem', width: '100%', maxWidth: '60rem' }}>
          <a href="/">
            <img src="/searxng.png" alt="SearGo" style={{ height: '2rem' }} />
          </a>
        </div>
      )}

      {/* Search form */}
      {!hasQuery ? (
        // Home page: centered large search
        <form onSubmit={handleSubmit} style={{ width: '100%', maxWidth: '40rem', padding: '0 1rem 2rem' }}>
          <div style={{ position: 'relative', display: 'flex' }}>
            <input
              type="text"
              value={query}
              onChange={e => setQuery(e.target.value)}
              placeholder="Search..."
              autoFocus
              style={{
                flex: 1,
                padding: '0.75rem 1rem',
                border: '1px solid var(--color-search-border)',
                borderRadius: '0.5rem 0 0 0.5rem',
                backgroundColor: 'var(--color-search-background)',
                color: 'var(--color-search-font)',
                fontSize: '1.1rem',
                outline: 'none',
                boxShadow: 'var(--color-search-shadow)',
              }}
            />
            <AutocompleteDropdown query={query} onSelect={(q) => { setQuery(q); }} />
            <button type="submit" disabled={isLoading} style={{
              padding: '0.75rem 1.5rem',
              backgroundColor: 'var(--color-button-background)',
              color: 'var(--color-button-font)',
              border: 'none',
              borderRadius: '0 0.5rem 0.5rem 0',
              cursor: 'pointer',
              fontWeight: 600,
              fontSize: '1.1rem',
            }}>
              Search
            </button>
          </div>
        </form>
      ) : (
        // Results page: compact search in header bar
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', padding: '0 0 0.75rem 1rem', width: '100%', maxWidth: '60rem' }}>
          <form onSubmit={handleSubmit} style={{ flex: 1 }}>
            <div style={{ position: 'relative', display: 'flex' }}>
              <input
                type="text"
                value={query}
                onChange={e => setQuery(e.target.value)}
                placeholder="Search..."
                style={{
                  flex: 1,
                  padding: '0.4rem 0.75rem',
                  border: '1px solid var(--color-search-border)',
                  borderRadius: '0.25rem 0 0 0.25rem',
                  backgroundColor: 'var(--color-search-background)',
                  color: 'var(--color-search-font)',
                  fontSize: '0.9rem',
                  outline: 'none',
                }}
              />
              <AutocompleteDropdown query={query} onSelect={(q) => { setQuery(q); }} />
              <button type="submit" disabled={isLoading} style={{
                padding: '0.4rem 1rem',
                backgroundColor: 'var(--color-button-background)',
                color: 'var(--color-button-font)',
                border: 'none',
                borderRadius: '0 0.25rem 0.25rem 0',
                cursor: 'pointer',
                fontWeight: 600,
              }}>
                {isLoading ? '...' : 'Search'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Category selector (results page only, below search bar) */}
      {hasQuery && <CategorySelector />}
    </div>
  );
}
