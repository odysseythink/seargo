import type { VideoResult } from '../../types/search';

interface Props { result: VideoResult }

export function VideoCard({ result }: Props) {
  const duration = result.extra?.duration || result.extra?.length;
  return (
    <a href={result.url} target="_blank" rel="noopener noreferrer"
       className="block group cursor-pointer">
      <div className="relative aspect-video bg-[#1a1a1a] rounded-xl overflow-hidden border border-[rgba(255,255,255,0.08)]">
        {result.extra?.thumbnail ? (
          <>
            <img src={result.extra.thumbnail} alt={result.title}
                 className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
            {duration && (
              <span className="absolute bottom-2 right-2 px-2 py-0.5 bg-black/70 text-white text-xs rounded">
                {duration}
              </span>
            )}
          </>
        ) : (
          <div className="w-full h-full flex items-center justify-center text-[#6b7280] text-sm">No thumbnail</div>
        )}
      </div>
      <p className="mt-1 text-sm text-[#e5e5e5] truncate font-medium">{result.title}</p>
      <div className="flex items-center gap-2 text-xs text-[#6b7280]">
        {result.extra?.author && <span>{result.extra.author}</span>}
        {result.extra?.upload_date && <span>{result.extra.upload_date}</span>}
        {result.extra?.view_count !== undefined && (
          <span>{result.extra.view_count.toLocaleString()} views</span>
        )}
      </div>
    </a>
  );
}
