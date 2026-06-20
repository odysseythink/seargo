import type { MainResult as MainResultType } from '../../types/search';

interface Props {
  result: MainResultType;
}

export function MainResult({ result }: Props) {
  const engineColors: Record<string, string> = {
    google: '#ea4335', bing: '#00809d', duckduckgo: '#de5833',
    brave: '#fb542b', wikipedia: '#3366cc', yahoo: '#6001d2',
  };
  const color = engineColors[result.engine?.toLowerCase()] || '#6b7280';

  return (
    <div className="p-5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl
                    hover:border-[rgba(255,255,255,0.15)] transition-all duration-200">
      {result.thumbnail_url && (
        <img src={result.thumbnail_url} alt="" className="w-16 h-16 object-cover rounded mb-2 float-right ml-2" />
      )}
      <a href={result.url} target="_blank" rel="noopener noreferrer"
         className="text-lg font-medium text-[#60a5fa] hover:text-[#93c5fd] hover:underline block mb-1">
        {result.title}
      </a>
      <p className="text-[#22c55e] text-sm mb-2 truncate">{result.url}</p>
      {result.content && (
        <p className="text-[#9ca3af] text-sm leading-relaxed">{result.content}</p>
      )}
      <div className="mt-3 flex items-center gap-2">
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium text-white"
              style={{ backgroundColor: color }}>
          {result.engine}
        </span>
        {result.score > 0 && (
          <span className="text-xs text-[#6b7280]">Score: {result.score.toFixed(2)}</span>
        )}
      </div>
    </div>
  );
}
