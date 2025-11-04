package s3

import (
	"github.com/prometheus/client_golang/prometheus"
)

// metricsExporter holds all Prometheus metrics for the S3 plugin
type metricsExporter struct {
	// operationsTotal tracks total operations by operation, bucket, and status
	operationsTotal *prometheus.CounterVec

	// errorsTotal tracks errors by bucket and error type
	errorsTotal *prometheus.CounterVec
}

// newMetricsExporter creates a new metrics exporter for S3 operations
// Returns error if metrics registration fails
func newMetricsExporter() (*metricsExporter, error) {
	m := &metricsExporter{
		// Operation counter with labels: operation, bucket, status
		operationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rr_s3_operations_total",
				Help: "Total number of S3 operations by type, bucket, and status",
			},
			[]string{"operation", "bucket", "status"},
		),

		// Error counter with labels: bucket, error_type
		errorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rr_s3_errors_total",
				Help: "Total number of S3 errors by bucket and error type",
			},
			[]string{"bucket", "error_type"},
		),
	}

	// Register metrics with Prometheus default registry
	// This ensures metrics are available even if MetricsCollector() isn't called
	if err := prometheus.Register(m.operationsTotal); err != nil {
		// Check if already registered (happens on plugin reload)
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return nil, err
		}
	}

	if err := prometheus.Register(m.errorsTotal); err != nil {
		// Check if already registered (happens on plugin reload)
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return nil, err
		}
	}

	return m, nil
}

// RecordOperation increments the operation counter
// operation: write, read, delete, copy, move, list, exists, get_metadata, set_visibility, get_url
// bucket: bucket name
// status: success, error
func (m *metricsExporter) RecordOperation(bucket, operation, status string) {
	if m == nil {
		return
	}
	m.operationsTotal.WithLabelValues(operation, bucket, status).Inc()
}

// RecordError increments the error counter
// bucket: bucket name
// errorType: error code from ErrorCode constants
func (m *metricsExporter) RecordError(bucket string, errorType ErrorCode) {
	if m == nil {
		return
	}
	m.errorsTotal.WithLabelValues(bucket, string(errorType)).Inc()
}

// getCollectors returns all Prometheus collectors for registration
func (m *metricsExporter) getCollectors() []prometheus.Collector {
	if m == nil {
		return nil
	}
	return []prometheus.Collector{
		m.operationsTotal,
		m.errorsTotal,
	}
}
