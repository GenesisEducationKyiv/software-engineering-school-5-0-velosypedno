package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerCacheMetricsOnce sync.Once
)

type CacheMetrics struct {
	cacheHits     prometheus.Counter
	cacheMisses   prometheus.Counter
	accessLatency prometheus.Histogram
}

func NewCacheMetrics(reg prometheus.Registerer) *CacheMetrics {
	cacheHits := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "weather_cache_hits",
		Help: "Number of weather cache hits",
	})

	cacheMisses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "weather_cache_miss",
		Help: "Number of weather cache misses",
	})

	accessLatency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "weather_cache_request_duration",
		Help:    "Duration of weather cache requests in seconds",
		Buckets: prometheus.DefBuckets,
	})
	registerCacheMetricsOnce.Do(func() {
		reg.MustRegister(cacheHits, cacheMisses, accessLatency)
	})

	return &CacheMetrics{
		cacheHits:     cacheHits,
		cacheMisses:   cacheMisses,
		accessLatency: accessLatency,
	}
}

func (m *CacheMetrics) CacheHit() {
	m.cacheHits.Inc()
}

func (m *CacheMetrics) CacheMiss() {
	m.cacheMisses.Inc()
}

func (m *CacheMetrics) CacheAccessLatency(seconds float64) {
	m.accessLatency.Observe(seconds)
}
