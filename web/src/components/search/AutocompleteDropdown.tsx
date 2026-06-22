import { useEffect, useRef, useCallback, useState } from 'react';
import type { AutocompleteSuggestion } from '../../types/search';
import { api } from '../../services/api';

const DEBOUNCE_MS = 300;
const MIN_LENGTH = 2;

interface AutocompleteDropdownProps {
  query: string;
  onSelect: (value: string) => void;
  onClose?: () => void;
  visible?: boolean;
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
    if (visible === false) return;
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
        onClose?.();
        break;
    }
  };

  const isVisible = visible !== undefined ? visible : (suggestions.length > 0 || loading);
  if (!isVisible || (suggestions.length === 0 && !loading)) return null;

  return (
    <div
      tabIndex={0}
      style={{
        position: 'absolute',
        left: 0,
        right: 0,
        top: '100%',
        marginTop: '0.25rem',
        backgroundColor: 'var(--color-autocomplete-background)',
        border: '1px solid var(--color-autocomplete-border)',
        borderRadius: '0.75rem',
        boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -2px rgba(0,0,0,0.05)',
        overflow: 'hidden',
        zIndex: 50,
      }}
      onKeyDown={handleKeyDown}
    >
      {loading && suggestions.length === 0 && (
        <div style={{
          padding: '0.75rem 1rem',
          color: 'var(--color-autocomplete-font)',
          fontSize: '0.875rem',
          opacity: 0.6,
        }}>
          Loading...
        </div>
      )}
      {suggestions.map((s, i) => (
        <button
          key={s.value}
          type="button"
          style={{
            width: '100%',
            textAlign: 'left',
            padding: '0.625rem 1rem',
            color: 'var(--color-autocomplete-font)',
            backgroundColor: i === activeIndex ? 'var(--color-autocomplete-selected)' : 'transparent',
            transition: 'background-color 0.1s',
            fontSize: '0.875rem',
            border: 'none',
            cursor: 'pointer',
          }}
          onClick={() => onSelect(s.value)}
          onMouseEnter={() => setActiveIndex(i)}
        >
          {s.label}
        </button>
      ))}
    </div>
  );
}
