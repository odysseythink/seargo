import type { CodeResult } from '../../types/search';

interface Props { result: CodeResult }

export function CodeBlock({ result }: Props) {
  const extra = result.extra;
  const lines = extra?.code_lines?.slice(0, 20) ?? [];
  return (
    <div className="p-5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl
                    hover:border-[rgba(255,255,255,0.15)] transition-all duration-200">
      <div className="flex items-center gap-2 mb-2 flex-wrap">
        {extra?.filename && (
          <span className="text-sm font-mono text-[#60a5fa]">{extra.filename}</span>
        )}
        {extra?.code_language && (
          <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-[#1e3a5f] text-[#60a5fa]">
            {extra.code_language}
          </span>
        )}
        {extra?.repository && (
          <a href={extra.repository} target="_blank" rel="noopener noreferrer"
             className="text-xs text-[#6b7280] hover:text-[#9ca3af] ml-auto truncate max-w-[200px]">
            {extra.repository}
          </a>
        )}
      </div>
      {lines.length > 0 ? (
        <pre className="bg-[#0f0f0f] rounded-lg p-4 overflow-x-auto text-sm font-mono leading-relaxed">
          {lines.map((l, i) => (
            <code key={i} className="text-[#9ca3af]">
              <span className="text-[#6b7280] select-none mr-4">{l.line}</span>
              <span>{l.text}</span>
              {'\n'}
            </code>
          ))}
        </pre>
      ) : (
        <p className="text-[#9ca3af] text-sm">{result.content}</p>
      )}
      <a href={result.url} target="_blank" rel="noopener noreferrer"
         className="mt-2 inline-block text-xs text-[#60a5fa] hover:underline">
        View source
      </a>
    </div>
  );
}
