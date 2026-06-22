import { useEffect, useState } from "react";
import { fetchStatsEngines, fetchStatsErrors } from "../services/api";
import type { EngineSnapshot, ErrorEntry } from "../types/search";

function ReliabilityBar({ value }: { value: number }) {
  const pct = Math.round(value * 100);
  const bgColor =
    pct >= 90 ? 'var(--color-engine-success)' : pct >= 70 ? 'var(--color-engine-warning)' : 'var(--color-engine-error)';
  return (
    <div style={{ width: '100%', backgroundColor: 'var(--color-result-border)', borderRadius: '0.25rem', height: '1rem' }}>
      <div
        style={{
          height: '1rem',
          borderRadius: '0.25rem',
          backgroundColor: bgColor,
          transition: 'width 0.3s ease',
          width: `${pct}%`,
        }}
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
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh' }}>
        <div style={{
          width: '2rem',
          height: '2rem',
          border: '2px solid var(--color-result-border)',
          borderTopColor: 'var(--color-result-link)',
          borderRadius: '50%',
          animation: 'spin 0.6s linear infinite',
        }} />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center' }}>
        <p style={{ color: 'var(--color-engine-error)' }}>
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
    <div style={{ maxWidth: '56rem', margin: '0 auto', padding: '1rem' }}>
      <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '1.5rem', color: 'var(--color-base-font)' }}>
        Engine Statistics
      </h1>

      {engines.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '3rem 0', color: 'var(--color-result-url)' }}>
          <p style={{ fontSize: '1.125rem' }}>No data</p>
          <p style={{ fontSize: '0.875rem', marginTop: '0.5rem' }}>
            Perform a search to start collecting engine metrics.
          </p>
        </div>
      ) : (
        <div style={{ display: 'grid', gap: '1rem', gridTemplateColumns: 'repeat(auto-fill, minmax(22rem, 1fr))' }}>
          {engines.map((eng) => {
            const err = errorMap.get(eng.engine);
            return (
              <div
                key={eng.engine}
                style={{
                  border: '1px solid var(--color-result-border)',
                  borderRadius: '0.5rem',
                  padding: '1rem',
                  backgroundColor: 'var(--color-result-background)',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                  <h2 style={{ fontWeight: 600, fontSize: '1.125rem', color: 'var(--color-base-font)', textTransform: 'capitalize' }}>
                    {eng.engine}
                  </h2>
                  {eng.suspended && (
                    <span style={{
                      padding: '0.125rem 0.5rem',
                      fontSize: '0.75rem',
                      backgroundColor: 'var(--color-engine-error-background)',
                      color: 'var(--color-engine-error)',
                      borderRadius: '0.25rem',
                    }}>
                      Suspended
                    </span>
                  )}
                </div>

                <ReliabilityBar value={eng.reliability} />

                <div style={{ marginTop: '0.75rem', display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem', fontSize: '0.875rem' }}>
                  <div>
                    <span style={{ color: 'var(--color-result-url)' }}>
                      Reliability
                    </span>
                    <p style={{ fontWeight: 500, color: 'var(--color-base-font)' }}>
                      {Math.round(eng.reliability * 100)}%
                    </p>
                  </div>
                  <div>
                    <span style={{ color: 'var(--color-result-url)' }}>
                      Avg Score
                    </span>
                    <p style={{ fontWeight: 500, color: 'var(--color-base-font)' }}>
                      {eng.score.toFixed(2)}
                    </p>
                  </div>
                  <div>
                    <span style={{ color: 'var(--color-result-url)' }}>
                      Requests
                    </span>
                    <p style={{ fontWeight: 500, color: 'var(--color-base-font)' }}>
                      {eng.request_count}
                    </p>
                  </div>
                  <div>
                    <span style={{ color: 'var(--color-result-url)' }}>
                      P50 Total
                    </span>
                    <p style={{ fontWeight: 500, color: 'var(--color-base-font)' }}>
                      {formatMs(eng.time.total.p50)}
                    </p>
                  </div>
                </div>

                {err && (
                  <div style={{ marginTop: '0.75rem', paddingTop: '0.75rem', borderTop: '1px solid var(--color-result-border)' }}>
                    <span style={{ fontSize: '0.75rem', color: 'var(--color-result-url)' }}>
                      Errors
                    </span>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.25rem', marginTop: '0.25rem' }}>
                      {Object.entries(err.by_class).map(([cls, count]) => (
                        <span
                          key={cls}
                          style={{
                            padding: '0.125rem 0.5rem',
                            fontSize: '0.75rem',
                            backgroundColor: 'var(--color-engine-error-background)',
                            color: 'var(--color-engine-error)',
                            borderRadius: '0.25rem',
                          }}
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
