package httpx

import (
	"strings"
	"time"

	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/internal/metrics"
)

// Response size limits to prevent OOM from unexpectedly large bodies.
const (
	maxResponseSize = 10 * 1024 * 1024 // 10 MB
	maxRequestSize  = 1 * 1024 * 1024  // 1 MB
)

// recordMetrics records outbound request metrics to Prometheus.
func recordMetrics(network, engine string, statusCode int, duration time.Duration, err error) {
	sc := statusClass(statusCode)
	metrics.OutboundRequestsTotal.WithLabelValues(network, engine, sc).Inc()
	metrics.OutboundRequestDuration.WithLabelValues(network, engine).Observe(duration.Seconds())

	if err != nil {
		ec := errorClass(err)
		if ec != "" {
			metrics.OutboundErrorsTotal.WithLabelValues(network, engine, ec).Inc()
		}
	}
}

// logResponse logs outbound request results.
// Debug level: full URL (including query). Info level: host only.
func logResponse(engine, network, method, url string, statusCode int, err error) {
	host := parseHost(url)
	sc := statusClass(statusCode)

	logger.Debug("outbound request",
		"engine", engine,
		"network", network,
		"method", method,
		"url", url,
		"status", statusCode,
		"status_class", sc,
		"error", err,
	)

	if err != nil {
		logger.Info("outbound request failed",
			"engine", engine,
			"network", network,
			"host", host,
			"status_code", statusCode,
			"status_class", sc,
			"error_class", errorClass(err),
		)
	} else {
		logger.Info("outbound request",
			"engine", engine,
			"network", network,
			"host", host,
			"status_code", statusCode,
			"status_class", sc,
		)
	}
}

// parseHost extracts the hostname from a URL string, stripping port if present.
func parseHost(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	s := rawURL
	if idx := strings.Index(s, "://"); idx != -1 {
		s = s[idx+3:]
	}
	if idx := strings.Index(s, "/"); idx != -1 {
		s = s[:idx]
	}
	if idx := strings.Index(s, "?"); idx != -1 {
		s = s[:idx]
	}
	// Strip port if present
	if idx := strings.Index(s, ":"); idx != -1 {
		s = s[:idx]
	}
	return s
}
