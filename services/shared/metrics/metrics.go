package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics singleton for Prometheus instrumentation
type Metrics struct {
	// HTTP Metrics
	HTTPRequestsTotal   prometheus.CounterVec
	HTTPRequestDuration prometheus.HistogramVec
	HTTPInProgress      prometheus.GaugeVec

	// Database Metrics
	DatabaseQueryDuration prometheus.HistogramVec
	DatabaseConnections   prometheus.Gauge
	DatabaseConnectionErr prometheus.Counter
	DatabaseErrors        prometheus.CounterVec

	// Service-specific metrics (declared in each service)
	// Data Processor metrics
	ProcessedRecordsTotal  prometheus.Counter
	ProcessingDuration     prometheus.Histogram
	ProcessingErrors       prometheus.CounterVec
	ProcessingQueueDepth   prometheus.Gauge
	LastProcessedTimestamp prometheus.Gauge

	// Oura Collector metrics
	CollectionRunsTotal   prometheus.Counter
	CollectionDuration    prometheus.Histogram
	DataPointsCollected   prometheus.Counter
	CollectionErrors      prometheus.CounterVec
	LastSuccessfulRunTime prometheus.Gauge
}

var (
	once     sync.Once
	instance *Metrics
)

// New creates or returns the singleton Metrics instance
func New(serviceName string) *Metrics {
	once.Do(func() {
		instance = &Metrics{
			// HTTP Metrics - Vector versions for labels
			HTTPRequestsTotal: *promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "http_requests_total",
					Help: "Total HTTP requests",
					ConstLabels: map[string]string{
						"service": serviceName,
					},
				},
				[]string{"method", "endpoint", "status"},
			),
			HTTPRequestDuration: *promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "http_request_duration_seconds",
					Help:    "HTTP request duration in seconds",
					Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
					ConstLabels: map[string]string{
						"service": serviceName,
					},
				},
				[]string{"method", "endpoint"},
			),
			HTTPInProgress: *promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "http_requests_in_progress",
					Help: "Number of HTTP requests in progress",
					ConstLabels: map[string]string{
						"service": serviceName,
					},
				},
				[]string{"method", "endpoint"},
			),

			// Database Metrics
			DatabaseQueryDuration: *promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "db_query_duration_seconds",
					Help:    "Database query duration in seconds",
					Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
					ConstLabels: map[string]string{
						"service": serviceName,
					},
				},
				[]string{"query_type", "table"},
			),
			DatabaseConnections: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "db_connections_open",
				Help: "Number of open database connections",
				ConstLabels: map[string]string{
					"service": serviceName,
				},
			}),
			DatabaseConnectionErr: promauto.NewCounter(prometheus.CounterOpts{
				Name: "db_connection_errors_total",
				Help: "Total database connection errors",
				ConstLabels: map[string]string{
					"service": serviceName,
				},
			}),
			DatabaseErrors: *promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "db_errors_total",
					Help: "Total database errors by type",
					ConstLabels: map[string]string{
						"service": serviceName,
					},
				},
				[]string{"error_type"},
			),

			// Data Processor Metrics
			ProcessedRecordsTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "processed_records_total",
				Help: "Total records processed",
				ConstLabels: map[string]string{
					"service": serviceName,
				},
			}),
			ProcessingDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "processing_duration_seconds",
				Help:    "Data processing duration in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
				ConstLabels: map[string]string{
					"service": serviceName,
				},
			}),
			ProcessingErrors: *promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "processing_errors_total",
					Help: "Total processing errors by type",
					ConstLabels: map[string]string{
						"service": serviceName,
					},
				},
				[]string{"error_type"},
			),
			ProcessingQueueDepth: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "processing_queue_depth",
				Help: "Current depth of processing queue",
				ConstLabels: map[string]string{
					"service": serviceName,
				},
			}),
			LastProcessedTimestamp: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "last_processed_timestamp_seconds",
				Help: "Timestamp of last processed record",
				ConstLabels: map[string]string{
					"service": serviceName,
				},
			}),

			// Oura Collector Metrics
			CollectionRunsTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "collector_runs_total",
				Help: "Total collector runs",
				ConstLabels: map[string]string{
					"service": serviceName,
				},
			}),
			CollectionDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "collector_run_duration_seconds",
				Help:    "Collector run duration in seconds",
				Buckets: []float64{5, 10, 30, 60, 120, 300, 600},
				ConstLabels: map[string]string{
					"service": serviceName,
				},
			}),
			DataPointsCollected: promauto.NewCounter(prometheus.CounterOpts{
				Name: "data_points_collected_total",
				Help: "Total data points collected from Oura API",
				ConstLabels: map[string]string{
					"service": serviceName,
				},
			}),
			CollectionErrors: *promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "collector_errors_total",
					Help: "Total collector errors by data type and error type",
					ConstLabels: map[string]string{
						"service": serviceName,
					},
				},
				[]string{"data_type", "error_type"},
			),
			LastSuccessfulRunTime: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "collector_last_successful_run_timestamp_seconds",
				Help: "Timestamp of last successful collection run",
				ConstLabels: map[string]string{
					"service": serviceName,
				},
			}),
		}
	})
	return instance
}

// GetInstance returns the singleton Metrics instance
func GetInstance() *Metrics {
	if instance == nil {
		return New("unknown")
	}
	return instance
}

// RecordHTTPRequest records an HTTP request with duration
func (m *Metrics) RecordHTTPRequest(method, endpoint string, status int, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, fmt.Sprintf("%d", status)).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordHTTPInProgress increments the in-progress counter
func (m *Metrics) RecordHTTPInProgress(method, endpoint string, increment float64) {
	m.HTTPInProgress.WithLabelValues(method, endpoint).Add(increment)
}

// RecordDatabaseQuery records a database query
func (m *Metrics) RecordDatabaseQuery(queryType, table string, duration time.Duration, err error) {
	m.DatabaseQueryDuration.WithLabelValues(queryType, table).Observe(duration.Seconds())
	if err != nil {
		m.DatabaseErrors.WithLabelValues("query_error").Inc()
	}
}
