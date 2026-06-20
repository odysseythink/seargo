import type { MapResult } from '../../types/search';

interface Props { result: MapResult }

export function MapCard({ result }: Props) {
  const extra = result.extra;
  return (
    <div className="p-5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl
                    hover:border-[rgba(255,255,255,0.15)] transition-all duration-200">
      <a href={result.url} target="_blank" rel="noopener noreferrer"
         className="text-lg font-medium text-[#60a5fa] hover:text-[#93c5fd] hover:underline block mb-1">
        {result.title}
      </a>
      {extra?.address && (
        <p className="text-sm text-[#9ca3af] mb-1">{extra.address}</p>
      )}
      {(extra?.latitude !== undefined || extra?.longitude !== undefined) && (
        <p className="text-xs text-[#6b7280] mb-1">
          {extra.latitude !== undefined && `${extra.latitude.toFixed(4)}`}
          {extra.latitude !== undefined && extra.longitude !== undefined && ', '}
          {extra.longitude !== undefined && `${extra.longitude.toFixed(4)}`}
        </p>
      )}
      {result.content && (
        <p className="text-[#9ca3af] text-sm leading-relaxed">{result.content}</p>
      )}
      <div className="mt-3 flex items-center gap-2">
        {extra?.map_url && (
          <a href={extra.map_url} target="_blank" rel="noopener noreferrer"
             className="text-xs text-[#22c55e] hover:underline">View on map</a>
        )}
        {result.engine && (
          <span className="text-xs text-[#6b7280]">{result.engine}</span>
        )}
      </div>
    </div>
  );
}
