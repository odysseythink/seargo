import type { KeyValueResult } from '../../types/search';

interface Props {
  result: KeyValueResult;
  index?: number;
  resultsOnNewTab?: boolean;
}

export function KeyValueTable({ result, index: _index, resultsOnNewTab: _resultsOnNewTab }: Props) {
  const extra = result.extra;
  const kvMap = extra?.kv_map;

  return (
    <div style={{
      padding: '1rem',
      backgroundColor: 'var(--color-result-background)',
      border: '1px solid var(--color-result-border)',
      borderRadius: '8px',
    }}>
      {extra?.caption && (
        <p style={{ fontSize: '0.875rem', fontWeight: 500, color: 'var(--color-base-font)', margin: '0 0 0.75rem' }}>{extra.caption}</p>
      )}
      {kvMap && Object.keys(kvMap).length > 0 ? (
        <table style={{ width: '100%', fontSize: '0.875rem', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--color-result-border)' }}>
              <th style={{ textAlign: 'left', padding: '0.5rem 1rem 0.5rem 0', color: 'var(--color-result-published-date)', fontWeight: 500 }}>
                {extra?.key_title || 'Key'}
              </th>
              <th style={{ textAlign: 'left', padding: '0.5rem 0', color: 'var(--color-result-published-date)', fontWeight: 500 }}>
                {extra?.value_title || 'Value'}
              </th>
            </tr>
          </thead>
          <tbody>
            {Object.entries(kvMap).map(([key, value]) => (
              <tr key={key} style={{ borderBottom: '1px solid var(--color-result-border)' }}>
                <td style={{ padding: '0.5rem 1rem 0.5rem 0', color: 'var(--color-base-font)' }}>{key}</td>
                <td style={{ padding: '0.5rem 0', color: 'var(--color-base-font)' }}>{value}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p style={{ color: 'var(--color-base-font)', fontSize: '0.875rem', margin: 0 }}>{result.content || 'No data available'}</p>
      )}
    </div>
  );
}
