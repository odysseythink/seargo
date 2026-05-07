import { useState } from 'react';
import { useSearchStore } from '../stores/searchStore';

export default function SearchPage() {
  const [input, setInput] = useState('');
  const { results, isLoading, enginesUsed, enginesFailed, responseTimeMs, error, search } = useSearchStore();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim()) {
      search({ q: input.trim() });
    }
  };

  return (
    <div style={{ maxWidth: 800, margin: '0 auto', padding: 20 }}>
      <h1>SearGo</h1>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Search..."
          style={{ width: '100%', padding: 10, fontSize: 16 }}
        />
        <button type="submit" disabled={isLoading} style={{ marginTop: 10, padding: '8px 20px' }}>
          {isLoading ? 'Searching...' : 'Search'}
        </button>
      </form>

      {error && <p style={{ color: 'red' }}>{error}</p>}

      {results.length > 0 && (
        <div style={{ marginTop: 20 }}>
          <p style={{ color: '#666', fontSize: 14 }}>
            Found {results.length} results in {responseTimeMs}ms
            {enginesUsed.length > 0 && ` · Engines: ${enginesUsed.join(', ')}`}
            {enginesFailed.length > 0 && ` · Failed: ${enginesFailed.join(', ')}`}
          </p>
          {results.map((r, i) => (
            <div key={i} style={{ marginTop: 16, padding: 12, border: '1px solid #eee', borderRadius: 8 }}>
              <a href={r.url} target="_blank" rel="noopener noreferrer" style={{ fontSize: 18, color: '#1a0dab' }}>
                {r.title}
              </a>
              <p style={{ color: '#006621', fontSize: 14, margin: '4px 0' }}>{r.url}</p>
              <p style={{ color: '#545454', fontSize: 14 }}>{r.content}</p>
              <span style={{ color: '#999', fontSize: 12 }}>Source: {r.engine}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
