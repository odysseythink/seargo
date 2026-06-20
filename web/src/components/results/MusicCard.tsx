import type { MusicResult } from '../../types/search';

interface Props { result: MusicResult }

export function MusicCard({ result }: Props) {
  const extra = result.extra;
  return (
    <div className="p-4 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl
                    hover:border-[rgba(255,255,255,0.15)] transition-all duration-200">
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 w-10 h-10 bg-[#1e3a5f] rounded-lg flex items-center justify-center text-[#60a5fa] text-lg">
          🎵
        </div>
        <div className="flex-1 min-w-0">
          <a href={result.url} target="_blank" rel="noopener noreferrer"
             className="text-base font-medium text-[#60a5fa] hover:text-[#93c5fd] hover:underline block truncate">
            {result.title}
          </a>
          <div className="flex items-center gap-3 mt-1 text-xs text-[#6b7280]">
            {extra?.artist && <span>{extra.artist}</span>}
            {extra?.album && <span>{extra.album}</span>}
            {extra?.duration && <span>{extra.duration}</span>}
          </div>
          {result.content && (
            <p className="text-[#9ca3af] text-sm leading-relaxed mt-1">{result.content}</p>
          )}
        </div>
      </div>
    </div>
  );
}
