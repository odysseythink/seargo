package storage

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// StorageOpsTotal counts storage operations by backend, operation, and success.
	StorageOpsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "seargo_storage_ops_total",
		Help: "Total number of storage operations.",
	}, []string{"backend", "operation", "success"})

	// StorageOpDuration measures storage operation duration in seconds.
	StorageOpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "seargo_storage_op_duration_seconds",
		Help:    "Duration of storage operations in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"backend", "operation"})
)
