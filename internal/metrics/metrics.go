package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "seargo_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	EngineQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_engine_queries_total",
			Help: "Total number of engine queries",
		},
		[]string{"engine", "status"},
	)

	EngineQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "seargo_engine_query_duration_seconds",
			Help:    "Engine query duration in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 5, 10},
		},
		[]string{"engine"},
	)

	SearchResultsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_search_results_total",
			Help: "Total number of search results returned",
		},
		[]string{"category"},
	)

	CacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"level"},
	)

	CacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"level"},
	)

	EngineFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_engine_failures_total",
			Help: "Total number of engine failures by reason",
		},
		[]string{"engine", "reason"},
	)

	EngineSuspended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "seargo_engine_suspended",
			Help: "Whether an engine is currently suspended (1=suspended, 0=active)",
		},
		[]string{"engine"},
	)

	ResultStreamTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_result_stream_total",
			Help: "Total number of results by stream type",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(EngineQueriesTotal)
	prometheus.MustRegister(EngineQueryDuration)
	prometheus.MustRegister(SearchResultsTotal)
	prometheus.MustRegister(CacheHits)
	prometheus.MustRegister(CacheMisses)
	prometheus.MustRegister(EngineFailuresTotal)
	prometheus.MustRegister(EngineSuspended)
	prometheus.MustRegister(ResultStreamTotal)
}
