package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestsTotal     prometheus.Counter
	HTTPRequestDuration   prometheus.Histogram
	DatabaseQueryDuration prometheus.Histogram
}

func New(serviceName string) *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
			ConstLabels: map[string]string{
				"service": serviceName,
			},
		}),
		HTTPRequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: []float64{.001, .01, .1, 1, 10},
			ConstLabels: map[string]string{
				"service": serviceName,
			},
		}),
		DatabaseQueryDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration",
			Buckets: []float64{.001, .01, .1, 1},
			ConstLabels: map[string]string{
				"service": serviceName,
			},
		}),
	}
}
