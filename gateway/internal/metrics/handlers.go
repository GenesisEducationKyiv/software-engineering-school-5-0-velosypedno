package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerHandlerMetricsOnce sync.Once
)

type HandlerMetrics struct {
	requests        *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewHandlerMetrics(reg prometheus.Registerer) *HandlerMetrics {
	metrics := &HandlerMetrics{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_requests_total",
				Help: "Total number of requests to gateway",
			},
			[]string{"method", "code", "endpoint"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_request_duration_seconds",
				Help:    "Request duration to gateway in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "code", "endpoint"},
		),
	}
	registerHandlerMetricsOnce.Do(func() {
		reg.MustRegister(metrics.requests, metrics.requestDuration)
	})

	return metrics
}

func (m *HandlerMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}
		statusCode := c.Writer.Status()

		labels := prometheus.Labels{
			"method":   method,
			"code":     toCodeLabel(statusCode),
			"endpoint": endpoint,
		}

		m.requests.With(labels).Inc()
		m.requestDuration.With(labels).Observe(duration)
	}
}

func toCodeLabel(code int) string {
	hundred := 100
	return fmt.Sprintf("%dxx", code/hundred)
}
