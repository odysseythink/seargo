import type { NewsResult as NewsResultType } from '../../types/search';

interface Props { result: NewsResultType }

export function NewsResult({ result }: Props) {
  return (
    <div className="p-5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl
                    hover:border-[rgba(255,255,255,0.15)] transition-all duration-200">
      <a href={result.url} target="_blank" rel="noopener noreferrer"
         className="text-lg font-medium text-[#60a5fa] hover:text-[#93c5fd] hover:underline block mb-1">
        {result.title}
      </a>
      <div className="flex items-center gap-2 text-xs text-[#6b7280] mb-2">
        {result.published_at && <span>{result.published_at}</span>}
        {result.engine && <span>· {result.engine}</span>}
      </div>
      {result.content && (
        <p className="text-[#9ca3af] text-sm leading-relaxed">{result.content}</p>
      )}
      <div className="mt-3 flex items-center gap-2">
        <span className="text-xs text-[#22c55e] truncate">{result.url}</span>
      </div>
    </div>
  );
}
