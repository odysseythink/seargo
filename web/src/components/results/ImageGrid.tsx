import type { ImageResult } from '../../types/search';
import { ImageCard } from './ImageCard';

interface Props { results: ImageResult[] }

export function ImageGrid({ results }: Props) {
  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-8">
      {results.map((r, i) => <ImageCard key={r.url + i} result={r} />)}
    </div>
  );
}
