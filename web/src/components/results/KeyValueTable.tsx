import type { KeyValueResult } from '../../types/search';

interface Props { result: KeyValueResult }

export function KeyValueTable({ result }: Props) {
  const extra = result.extra;
  const kvMap = extra?.kv_map;

  return (
    <div className="p-5 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)] rounded-xl">
      {extra?.caption && (
        <p className="text-sm font-medium text-[#e5e5e5] mb-3">{extra.caption}</p>
      )}
      {kvMap && Object.keys(kvMap).length > 0 ? (
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[rgba(255,255,255,0.08)]">
              <th className="text-left py-2 pr-4 text-[#6b7280] font-medium">
                {extra?.key_title || 'Key'}
              </th>
              <th className="text-left py-2 text-[#6b7280] font-medium">
                {extra?.value_title || 'Value'}
              </th>
            </tr>
          </thead>
          <tbody>
            {Object.entries(kvMap).map(([key, value]) => (
              <tr key={key} className="border-b border-[rgba(255,255,255,0.04)] last:border-0">
                <td className="py-2 pr-4 text-[#9ca3af]">{key}</td>
                <td className="py-2 text-[#e5e5e5]">{value}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p className="text-[#9ca3af] text-sm">{result.content || 'No data available'}</p>
      )}
    </div>
  );
}
