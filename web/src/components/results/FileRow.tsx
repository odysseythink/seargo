import type { FileResult } from '../../types/search';

interface Props { result: FileResult }

export function FileRow({ result }: Props) {
  const extra = result.extra;
  const fileSize = extra?.file_size !== undefined
    ? extra.file_size > 1_000_000_000
      ? `${(extra.file_size / 1_000_000_000).toFixed(1)} GB`
      : extra.file_size > 1_000_000
        ? `${(extra.file_size / 1_000_000).toFixed(1)} MB`
        : extra.file_size > 1_000
          ? `${(extra.file_size / 1_000).toFixed(1)} KB`
          : `${extra.file_size} B`
    : null;

  return (
    <div className="p-4 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl
                    hover:border-[rgba(255,255,255,0.15)] transition-all duration-200">
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 w-10 h-10 bg-[#2d1b69] rounded-lg flex items-center justify-center text-[#a78bfa] text-lg">
          📄
        </div>
        <div className="flex-1 min-w-0">
          <a href={result.url} target="_blank" rel="noopener noreferrer"
             className="text-base font-medium text-[#60a5fa] hover:text-[#93c5fd] hover:underline block truncate">
            {extra?.filename || result.title}
          </a>
          <div className="flex items-center gap-3 mt-1 text-xs text-[#6b7280]">
            {extra?.file_type && <span>{extra.file_type.toUpperCase()}</span>}
            {fileSize && <span>{fileSize}</span>}
            {extra?.seeders !== undefined && <span>Seeders: {extra.seeders}</span>}
            {extra?.leechers !== undefined && <span>Leechers: {extra.leechers}</span>}
          </div>
          {extra?.magnet_uri && (
            <a href={extra.magnet_uri}
               className="mt-1 inline-block text-xs text-[#22c55e] hover:underline truncate max-w-full">
              Magnet link
            </a>
          )}
        </div>
      </div>
    </div>
  );
}
