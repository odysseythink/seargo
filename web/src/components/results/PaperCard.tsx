import type { PaperResult } from '../../types/search';

interface Props { result: PaperResult }

export function PaperCard({ result }: Props) {
  const extra = result.extra;
  return (
    <div className="p-5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl
                    hover:border-[rgba(255,255,255,0.15)] transition-all duration-200">
      <a href={result.url} target="_blank" rel="noopener noreferrer"
         className="text-lg font-medium text-[#60a5fa] hover:text-[#93c5fd] hover:underline block mb-1">
        {result.title}
      </a>
      {extra?.authors && extra.authors.length > 0 && (
        <p className="text-sm text-[#9ca3af] mb-1">{extra.authors.join(', ')}</p>
      )}
      {extra?.journal && (
        <p className="text-sm text-[#6b7280] italic mb-1">{extra.journal}</p>
      )}
      {result.content && (
        <p className="text-[#9ca3af] text-sm leading-relaxed mt-1">{result.content}</p>
      )}
      <div className="mt-3 flex flex-wrap items-center gap-2">
        {extra?.doi && (
          <a href={`https://doi.org/${extra.doi}`} target="_blank" rel="noopener noreferrer"
             className="text-xs text-[#60a5fa] hover:underline">DOI: {extra.doi}</a>
        )}
        {extra?.publisher && (
          <span className="text-xs text-[#6b7280]">{extra.publisher}</span>
        )}
        {extra?.type && (
          <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-[#2d1b69] text-[#a78bfa]">
            {extra.type}
          </span>
        )}
        {result.engine && (
          <span className="text-xs text-[#6b7280]">{result.engine}</span>
        )}
      </div>
    </div>
  );
}
