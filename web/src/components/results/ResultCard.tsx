import type { Result } from '../../types/search';
import { MainResult } from './MainResult';
import { ImageCard } from './ImageCard';
import { VideoCard } from './VideoCard';
import { NewsResult } from './NewsResult';
import { PaperCard } from './PaperCard';
import { CodeBlock } from './CodeBlock';
import { FileRow } from './FileRow';
import { MapCard } from './MapCard';
import { MusicCard } from './MusicCard';
import { AnswerBox } from './AnswerBox';
import { KeyValueTable } from './KeyValueTable';
import { InfoboxPanel } from './InfoboxPanel';

interface Props { result: Result }

export function ResultCard({ result }: Props) {
  switch (result.kind) {
    case 'image':
      return <ImageCard result={result} />;
    case 'video':
      return <VideoCard result={result} />;
    case 'news':
      return <NewsResult result={result} />;
    case 'paper':
      return <PaperCard result={result} />;
    case 'code':
      return <CodeBlock result={result} />;
    case 'file':
      return <FileRow result={result} />;
    case 'map':
      return <MapCard result={result} />;
    case 'music':
      return <MusicCard result={result} />;
    case 'answer':
      return <AnswerBox result={result} />;
    case 'keyvalue':
      return <KeyValueTable result={result} />;
    case 'infobox':
      return <InfoboxPanel result={result} />;
    case 'main':
    default:
      return <MainResult result={result as import('../../types/search').MainResult} />;
  }
}
