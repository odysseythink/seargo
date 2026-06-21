import { useEffect, useRef, useCallback, useState } from 'react';
import type { AutocompleteSuggestion } from '../../types/search';
import { api } from '../../services/api';

const DEBOUNCE_MS = 300;
const MIN_LENGTH = 2;

interface AutocompleteDropdownProps {
  query: string;
  onSelect: (value: string) => void;
  onClose: () => void;
  visible: boolean;
}

export default function AutocompleteDropdown({
  query,
  onSelect,
  onClose,
  visible,
}: AutocompleteDropdownProps) {
  const [suggestions, setSuggestions] = useState<AutocompleteSuggestion[]>([]);
  const [loading, setLoading] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const fetchSuggestions = useCallback(async (q: string) => {
    if (q.length < MIN_LENGTH) {
      setSuggestions([]);
      return;
    }
    if (abortRef.current) abortRef.current.abort();
    abortRef.current = new AbortController();

    setLoading(true);
    try {
      const response = await api.autocomplete(q);
      setSuggestions(response.data.suggestions || []);
      setActiveIndex(-1);
    } catch {
      setSuggestions([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!visible) return;
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => fetchSuggestions(query), DEBOUNCE_MS);
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [query, visible, fetchSuggestions]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setActiveIndex((prev) =>
          prev < suggestions.length - 1 ? prev + 1 : 0
        );
        break;
      case 'ArrowUp':
        e.preventDefault();
        setActiveIndex((prev) =>
          prev > 0 ? prev - 1 : suggestions.length - 1
        );
        break;
      case 'Enter':
        e.preventDefault();
        if (activeIndex >= 0 && activeIndex < suggestions.length) {
          onSelect(suggestions[activeIndex].value);
        }
        break;
      case 'Escape':
        onClose();
        break;
    }
  };

  if (!visible || (suggestions.length === 0 && !loading)) return null;

  return (
    <div
      tabIndex={0}
      className="absolute left-0 right-0 top-full mt-1 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)]
                 rounded-xl shadow-lg overflow-hidden z-50"
      onKeyDown={handleKeyDown}
    >
      {loading && suggestions.length === 0 && (
        <div className="px-4 py-3 text-[#6b7280] text-sm">Loading...</div>
      )}
      {suggestions.map((s, i) => (
        <button
          key={s.value}
          type="button"
          className={`w-full text-left px-4 py-2.5 text-[#e5e5e5] hover:bg-[#3b82f6]/20
                     transition-colors duration-100 text-sm
                     ${i === activeIndex ? 'bg-[#3b82f6]/20' : ''}`}
          onClick={() => onSelect(s.value)}
          onMouseEnter={() => setActiveIndex(i)}
        >
          {s.label}
        </button>
      ))}
    </div>
  );
}
