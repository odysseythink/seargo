import type { ImageResult } from '../../types/search';

interface Props { result: ImageResult }

export function ImageCard({ result }: Props) {
  const imgSrc = result.extra?.img_src || result.thumbnail_url || '';
  const title = result.title || 'Image';
  return (
    <a href={result.url} target="_blank" rel="noopener noreferrer"
       className="block group cursor-pointer">
      <div className="aspect-[4/3] bg-[#1a1a1a] rounded-xl overflow-hidden border border-[rgba(255,255,255,0.08)]">
        {imgSrc ? (
          <img src={imgSrc} alt={title}
               className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-[#6b7280] text-sm">No image</div>
        )}
      </div>
      {result.extra?.resolution && (
        <p className="mt-1 text-xs text-[#6b7280]">{result.extra.resolution}</p>
      )}
      <p className="mt-1 text-sm text-[#9ca3af] truncate">{title}</p>
    </a>
  );
}
