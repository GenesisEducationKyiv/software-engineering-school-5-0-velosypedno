package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerProviderMetricsOnce sync.Once
)

type ProviderMetrics struct {
	requests        *prometheus.CounterVec
	errors          *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewProviderMetrics(reg prometheus.Registerer) *ProviderMetrics {
	metrics := &ProviderMetrics{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "provider_requests_total",
				Help: "Total number of requests to weather providers",
			},
			[]string{"provider"},
		),
		errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "provider_errors_total",
				Help: "Total number of errors from weather providers",
			},
			[]string{"provider"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "provider_request_duration_seconds",
				Help:    "Request duration to provider in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"provider"},
		),
	}
	registerProviderMetricsOnce.Do(func() {
		reg.MustRegister(metrics.requests, metrics.errors, metrics.requestDuration)
	})

	return metrics
}

func (m *ProviderMetrics) Request(providerName string) {
	m.requests.WithLabelValues(providerName).Inc()
}

func (m *ProviderMetrics) Error(providerName string) {
	m.errors.WithLabelValues(providerName).Inc()
}

func (m *ProviderMetrics) RequestDuration(providerName string, seconds float64) {
	m.requestDuration.WithLabelValues(providerName).Observe(seconds)
}
