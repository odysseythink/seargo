import { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useSearchStore } from '../stores/searchStore';
import { AnswerBox } from '../components/results/AnswerBox';
import { ResultCard } from '../components/results/ResultCard';
import { ImageGrid } from '../components/results/ImageGrid';
import Pagination from '../components/search/Pagination';
import ResultsSidebar from '../components/results/Sidebar';

export default function SearchPage() {
  const [searchParams] = useSearchParams();
  const { results, answers, corrections, isLoading, enginesUsed, responseTimeMs, error, search } = useSearchStore();
  const hasQuery = searchParams.has('q');

  useEffect(() => {
    if (hasQuery) {
      search({
        q: searchParams.get('q') || '',
        category: searchParams.get('category') || 'general',
        page: parseInt(searchParams.get('page') || '1'),
      });
    }
  }, [searchParams, search, hasQuery]);

  const imageResults = results.filter(r => r.kind === 'image');
  const mainResults = results.filter(r => r.kind !== 'image');
  const resultsOnNewTab = false;

  if (!hasQuery) return null;

  return (
    <div>
      {/* Error */}
      {error && (
        <div style={{ padding: '0.75rem 1rem', color: 'var(--color-engine-error)', backgroundColor: 'var(--color-engine-error-background)', borderRadius: '0.25rem', marginBottom: '1rem', fontSize: '0.9rem' }}>
          {error}
        </div>
      )}

      {/* Loading */}
      {isLoading && (
        <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--color-base-font)' }}>Searching...</div>
      )}

      {/* Search stats */}
      {!isLoading && results.length > 0 && (
        <div style={{ marginBottom: '0.75rem', fontSize: '0.8rem', color: 'var(--color-base-font)', opacity: 0.7 }}>
          {results.length} results ({responseTimeMs}ms) &middot; Engines: {enginesUsed.join(', ')}
        </div>
      )}

      {/* Corrections */}
      {corrections?.length > 0 && (
        <div style={{ marginBottom: '0.75rem', fontSize: '0.9rem' }}>
          <span style={{ color: 'var(--color-base-font)' }}>Did you mean: </span>
          {corrections.map((c: string, i: number) => (
            <a key={i} href={`/?q=${encodeURIComponent(c)}`}
              style={{ color: 'var(--color-result-link)', marginRight: '0.5rem', fontWeight: 500 }}>{c}</a>
          ))}
        </div>
      )}

      {/* Answers */}
      {answers?.length > 0 && answers.map((a: any, i: number) => (
        <div key={i} style={{ marginBottom: '0.75rem' }}>
          <AnswerBox answer={a} />
        </div>
      ))}

      {/* Results + Sidebar grid */}
      <div style={{ display: 'flex', gap: '1.5rem' }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          {/* Image results */}
          {imageResults.length > 0 && <ImageGrid results={imageResults} resultsOnNewTab={resultsOnNewTab} />}

          {/* Main results */}
          {mainResults.map((r, i) => (
            <ResultCard key={i} result={r} index={i} resultsOnNewTab={resultsOnNewTab} />
          ))}

          {/* Pagination */}
          <Pagination />
        </div>

        {/* Sidebar */}
        {!isLoading && <ResultsSidebar />}
      </div>

      {/* Empty state */}
      {!isLoading && results.length === 0 && !error && (
        <div style={{ textAlign: 'center', padding: '4rem' }}>
          <p style={{ fontSize: '1.2rem', color: 'var(--color-base-font)', marginBottom: '1rem' }}>No results found.</p>
          <p style={{ color: 'var(--color-result-url)' }}>Try different keywords or check your search settings.</p>
        </div>
      )}
    </div>
  );
}
