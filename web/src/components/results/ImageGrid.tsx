import type { ImageResult } from '../../types/search';
import { ImageCard } from './ImageCard';

interface Props { results: ImageResult[]; resultsOnNewTab?: boolean; }

export function ImageGrid({ results, resultsOnNewTab }: Props) {
  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fill, minmax(10rem, 1fr))',
      gap: '1rem',
      marginBottom: '2rem',
    }}>
      {results.map((r, i) => <ImageCard key={r.url + i} result={r} index={i} resultsOnNewTab={resultsOnNewTab} />)}
    </div>
  );
}
