import { useState } from 'react';
import { useSearchStore } from '../stores/searchStore';
import { ResultCard } from '../components/results/ResultCard';
import { ImageGrid } from '../components/results/ImageGrid';
import { AnswerBox } from '../components/results/AnswerBox';
import { InfoboxPanel } from '../components/results/InfoboxPanel';

export default function SearchPage() {
  const [input, setInput] = useState('');
  const { results, answers, corrections, infoboxes, isLoading, enginesUsed, enginesFailed, responseTimeMs, error, search } = useSearchStore();
  const hasSearched = results.length > 0 || error !== null || enginesUsed.length > 0;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim()) {
      search({ q: input.trim() });
    }
  };

  // Separate image results for grid layout
  const imageResults = results.filter(r => r.kind === 'image') as any[];
  const nonImageResults = results.filter(r => r.kind !== 'image');

  return (
    <div className="min-h-screen bg-[#0f0f0f] text-[#e5e5e5]">
      <div className="max-w-3xl mx-auto px-4 py-12">
        {/* Logo / Title */}
        <div className={`text-center transition-all duration-500 ${hasSearched ? 'mb-6' : 'mb-12 mt-20'}`}>
          <h1 className="text-5xl font-bold tracking-tight mb-2">
            <span className="text-[#3b82f6]">Sear</span>Go
          </h1>
          <p className="text-[#9ca3af] text-sm">Privacy-respecting meta search</p>
        </div>

        {/* Search Box */}
        <form onSubmit={handleSubmit} className="relative mb-8">
          <div className="flex gap-2">
            <div className="flex-1 relative">
              <input
                type="text" value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="Search the web..."
                className="w-full px-5 py-3.5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl
                         text-[#e5e5e5] placeholder-[#6b7280] outline-none
                         focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/30
                         transition-all duration-200 text-base"
              />
              {input && (
                <button type="button" onClick={() => setInput('')}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-[#6b7280] hover:text-[#e5e5e5]">
                  ✕
                </button>
              )}
            </div>
            <button type="submit" disabled={isLoading}
              className="px-6 py-3.5 bg-[#3b82f6] hover:bg-[#2563eb] disabled:bg-[#1e3a5f]
                       rounded-xl font-medium transition-all duration-200
                       flex items-center gap-2 min-w-[100px] justify-center">
              {isLoading ? (
                <span className="inline-block w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : 'Search'}
            </button>
          </div>
        </form>

        {/* Error */}
        {error && (
          <div className="mb-6 p-4 bg-red-900/20 border border-red-500/30 rounded-xl text-red-300">{error}</div>
        )}

        {/* Results Stats */}
        {(results.length > 0 || enginesFailed.length > 0) && (
          <div className="mb-4 text-sm text-[#9ca3af]">
            Found <span className="text-[#e5e5e5] font-medium">{results.length}</span> results
            {responseTimeMs > 0 && ` in ${responseTimeMs}ms`}
            {enginesUsed.length > 0 && <span> · Engines: {enginesUsed.join(', ')}</span>}
            {enginesFailed.length > 0 && <span className="text-red-400"> · Failed: {enginesFailed.join(', ')}</span>}
          </div>
        )}

        {/* Answers */}
        {answers.length > 0 && (
          <div className="mb-6 space-y-2">
            {answers.map((a, i) => <AnswerBox key={i} answer={a} />)}
          </div>
        )}

        {/* Corrections */}
        {corrections.length > 0 && (
          <div className="mb-6 p-3 bg-yellow-900/20 border border-yellow-500/30 rounded-xl text-yellow-300 text-sm">
            Did you mean: {corrections.join(', ')}?
          </div>
        )}

        {/* Infoboxes */}
        {infoboxes.length > 0 && (
          <div className="mb-6 space-y-4">
            {infoboxes.map((inf, i) => <InfoboxPanel key={i} infobox={inf} />)}
          </div>
        )}

        {/* Image Grid (when images present) */}
        {imageResults.length > 0 && <ImageGrid results={imageResults} />}

        {/* Non-image Results */}
        <div className="space-y-3">
          {nonImageResults.map((r, i) => (
            <div key={r.url + i}
                 className="animate-fade-in"
                 style={{ animationDelay: `${i * 60}ms` }}>
              <ResultCard result={r} />
            </div>
          ))}
        </div>

        {/* Empty state */}
        {hasSearched && results.length === 0 && !isLoading && !error && (
          <div className="text-center py-12 text-[#6b7280]">
            <p className="text-lg mb-2">No results found</p>
            <p className="text-sm">Try a different query or check your engine configuration</p>
          </div>
        )}
      </div>
    </div>
  );
}
