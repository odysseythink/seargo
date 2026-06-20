import type { AnswerResult, Answer } from '../../types/search';

interface Props {
  result?: AnswerResult;
  answer?: Answer;
}

export function AnswerBox({ result, answer }: Props) {
  const text = result?.extra?.answer || answer?.answer || result?.title || '';
  const content = result?.content || answer?.content || '';
  const sourceUrl = result?.url || answer?.url;
  const engine = result?.engine || answer?.engine;

  return (
    <div className="p-5 bg-[#1a1a1a] border border-[#3b82f6]/20 rounded-xl">
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 mt-1">
          <svg className="w-5 h-5 text-[#3b82f6]" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
          </svg>
        </div>
        <div className="flex-1">
          {text && (
            <p className="text-base font-medium text-[#e5e5e5] mb-1">{text}</p>
          )}
          {content && (
            <p className="text-sm text-[#9ca3af] leading-relaxed">{content}</p>
          )}
          <div className="mt-2 flex items-center gap-3">
            {sourceUrl && (
              <a href={sourceUrl} target="_blank" rel="noopener noreferrer"
                 className="text-xs text-[#60a5fa] hover:underline">Source</a>
            )}
            {engine && (
              <span className="text-xs text-[#6b7280]">{engine}</span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
