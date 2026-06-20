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

	OutboundRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_outbound_requests_total",
			Help: "Total number of outbound HTTP requests by network, engine, and status class",
		},
		[]string{"network", "engine", "status_class"},
	)

	OutboundRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "seargo_outbound_request_duration_seconds",
			Help:    "Outbound HTTP request duration in seconds by network and engine",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
		},
		[]string{"network", "engine"},
	)

	OutboundErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_outbound_errors_total",
			Help: "Total number of outbound request errors by network, engine, and error class",
		},
		[]string{"network", "engine", "error_class"},
	)

	EngineReloadsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_engine_reloads_total",
			Help: "Total number of engine reloads.",
		},
		[]string{"engine", "status"},
	)

	EngineParserFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_engine_parser_failures_total",
			Help: "Total number of engine parser failures.",
		},
		[]string{"engine"},
	)

	EngineResults = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "seargo_engine_results",
			Help:    "Number of results per engine per search.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"engine"},
	)

	// Plugin metrics
	PluginLoadTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_plugin_load_total",
			Help: "Total number of plugin load attempts.",
		},
		[]string{"plugin", "status"},
	)
	PluginHookDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "seargo_plugin_hook_duration_seconds",
			Help:    "Duration of plugin hook execution.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"plugin", "hook"},
	)

	// Answerer metrics
	AnswererAskTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seargo_answerer_ask_total",
			Help: "Total number of answerer invocations.",
		},
		[]string{"answerer", "status"},
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
	prometheus.MustRegister(OutboundRequestsTotal)
	prometheus.MustRegister(OutboundRequestDuration)
	prometheus.MustRegister(OutboundErrorsTotal)
	prometheus.MustRegister(EngineReloadsTotal)
	prometheus.MustRegister(EngineParserFailures)
	prometheus.MustRegister(EngineResults)
	prometheus.MustRegister(PluginLoadTotal)
	prometheus.MustRegister(PluginHookDuration)
	prometheus.MustRegister(AnswererAskTotal)
}
