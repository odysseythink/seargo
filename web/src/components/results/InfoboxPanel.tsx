import type { InfoboxResult, Infobox } from '../../types/search';

interface Props {
  result?: InfoboxResult;
  infobox?: Infobox;
}

export function InfoboxPanel({ result, infobox }: Props) {
  const title = infobox?.title || result?.title || '';
  const content = infobox?.content || result?.content || '';
  const engine = infobox?.engine || result?.engine || '';
  const imgSrc = infobox?.img_src || result?.extra?.img_src;
  const urls = infobox?.urls || result?.extra?.urls;
  const attributes = infobox?.attributes || result?.extra?.attributes;
  const relatedTopics = infobox?.related_topics || result?.extra?.related_topics;
  const srcUrl = infobox?.url || result?.url;

  return (
    <div className="p-5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl">
      <div className="flex items-start gap-4">
        {imgSrc && (
          <img src={imgSrc} alt={title}
               className="w-20 h-20 object-cover rounded-lg flex-shrink-0" />
        )}
        <div className="flex-1 min-w-0">
          {title && (
            <h3 className="text-lg font-semibold text-[#e5e5e5] mb-1">{title}</h3>
          )}
          {content && (
            <p className="text-sm text-[#9ca3af] leading-relaxed mb-3">{content}</p>
          )}
          {attributes && attributes.length > 0 && (
            <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm mb-3">
              {attributes.map((attr, i) => (
                <div key={i} className="contents">
                  <dt className="text-[#6b7280]">{attr.label}</dt>
                  <dd className="text-[#e5e5e5]">
                    {attr.url ? (
                      <a href={attr.url} target="_blank" rel="noopener noreferrer"
                         className="text-[#60a5fa] hover:underline">{attr.value}</a>
                    ) : attr.value}
                  </dd>
                </div>
              ))}
            </dl>
          )}
          {urls && urls.length > 0 && (
            <div className="mb-3">
              <p className="text-xs text-[#6b7280] mb-1 font-medium">Links</p>
              <div className="flex flex-wrap gap-2">
                {urls.map((u, i) => (
                  <a key={i} href={u.url} target="_blank" rel="noopener noreferrer"
                     className="text-xs text-[#60a5fa] hover:underline">
                    {u.title}
                  </a>
                ))}
              </div>
            </div>
          )}
          {relatedTopics && relatedTopics.length > 0 && (
            <div>
              <p className="text-xs text-[#6b7280] mb-1 font-medium">Related</p>
              <div className="flex flex-wrap gap-1">
                {relatedTopics.map((topic, i) => (
                  <span key={i}
                        className="inline-flex items-center px-2 py-0.5 rounded-full text-xs bg-[#2d1b69] text-[#a78bfa]">
                    {topic}
                  </span>
                ))}
              </div>
            </div>
          )}
          {engine && (
            <div className="mt-3 text-xs text-[#6b7280]">{engine}</div>
          )}
          {srcUrl && (
            <a href={srcUrl} target="_blank" rel="noopener noreferrer"
               className="mt-1 inline-block text-xs text-[#22c55e] hover:underline">
              Learn more
            </a>
          )}
        </div>
      </div>
    </div>
  );
}
