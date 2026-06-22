import { useEffect, useState } from "react";
import { fetchStatsEngines, fetchStatsErrors } from "../services/api";
import type { EngineSnapshot, ErrorEntry } from "../types/search";

function ReliabilityBar({ value }: { value: number }) {
  const pct = Math.round(value * 100);
  const color =
    pct >= 90 ? "bg-green-500" : pct >= 70 ? "bg-yellow-500" : "bg-red-500";
  return (
    <div className="w-full bg-gray-200 dark:bg-gray-700 rounded h-4">
      <div
        className={`h-4 rounded ${color} transition-all duration-300`}
        style={{ width: `${pct}%` }}
        title={`${pct}% reliable`}
      />
    </div>
  );
}

function formatMs(v: number) {
  return `${(v * 1000).toFixed(0)}ms`;
}

export default function StatsPage() {
  const [engines, setEngines] = useState<EngineSnapshot[]>([]);
  const [errors, setErrors] = useState<ErrorEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([fetchStatsEngines(), fetchStatsErrors()])
      .then(([engResp, errResp]) => {
        setEngines(engResp.engines || []);
        setErrors(errResp.errors || []);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <p className="text-red-600 dark:text-red-400">
          Failed to load stats: {error}
        </p>
      </div>
    );
  }

  const errorMap = new Map<string, ErrorEntry>();
  for (const e of errors) {
    errorMap.set(e.engine, e);
  }

  return (
    <div className="max-w-4xl mx-auto p-4 md:p-6">
      <h1 className="text-2xl font-bold mb-6 text-gray-900 dark:text-gray-100">
        Engine Statistics
      </h1>

      {engines.length === 0 ? (
        <div className="text-center py-12 text-gray-500 dark:text-gray-400">
          <p className="text-lg">No data</p>
          <p className="text-sm mt-2">
            Perform a search to start collecting engine metrics.
          </p>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {engines.map((eng) => {
            const err = errorMap.get(eng.engine);
            return (
              <div
                key={eng.engine}
                className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-white dark:bg-gray-800"
              >
                <div className="flex items-center justify-between mb-2">
                  <h2 className="font-semibold text-lg text-gray-900 dark:text-gray-100 capitalize">
                    {eng.engine}
                  </h2>
                  {eng.suspended && (
                    <span className="px-2 py-0.5 text-xs bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-300 rounded">
                      Suspended
                    </span>
                  )}
                </div>

                <ReliabilityBar value={eng.reliability} />

                <div className="mt-3 grid grid-cols-2 gap-2 text-sm">
                  <div>
                    <span className="text-gray-500 dark:text-gray-400">
                      Reliability
                    </span>
                    <p className="font-medium text-gray-900 dark:text-gray-100">
                      {Math.round(eng.reliability * 100)}%
                    </p>
                  </div>
                  <div>
                    <span className="text-gray-500 dark:text-gray-400">
                      Avg Score
                    </span>
                    <p className="font-medium text-gray-900 dark:text-gray-100">
                      {eng.score.toFixed(2)}
                    </p>
                  </div>
                  <div>
                    <span className="text-gray-500 dark:text-gray-400">
                      Requests
                    </span>
                    <p className="font-medium text-gray-900 dark:text-gray-100">
                      {eng.request_count}
                    </p>
                  </div>
                  <div>
                    <span className="text-gray-500 dark:text-gray-400">
                      P50 Total
                    </span>
                    <p className="font-medium text-gray-900 dark:text-gray-100">
                      {formatMs(eng.time.total.p50)}
                    </p>
                  </div>
                </div>

                {err && (
                  <div className="mt-3 pt-3 border-t border-gray-100 dark:border-gray-700">
                    <span className="text-xs text-gray-500 dark:text-gray-400">
                      Errors
                    </span>
                    <div className="flex flex-wrap gap-1 mt-1">
                      {Object.entries(err.by_class).map(([cls, count]) => (
                        <span
                          key={cls}
                          className="px-2 py-0.5 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded"
                        >
                          {cls}: {count}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
