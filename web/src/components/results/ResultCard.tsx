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

interface Props {
  result: Result;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function ResultCard({ result, index, resultsOnNewTab }: Props) {
  switch (result.kind) {
    case 'image':
      return <ImageCard result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'video':
      return <VideoCard result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'news':
      return <NewsResult result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'paper':
      return <PaperCard result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'code':
      return <CodeBlock result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'file':
      return <FileRow result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'map':
      return <MapCard result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'music':
      return <MusicCard result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'answer':
      return <AnswerBox result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'keyvalue':
      return <KeyValueTable result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'infobox':
      return <InfoboxPanel result={result} index={index} resultsOnNewTab={resultsOnNewTab} />;
    case 'main':
    default:
      return <MainResult result={result as import('../../types/search').MainResult} index={index} resultsOnNewTab={resultsOnNewTab} />;
  }
}
