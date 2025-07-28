package metrics

import (
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerWeatherMetricsOnce sync.Once
)

type WeatherMetrics struct {
	cacheHits     prometheus.Counter
	cacheMisses   prometheus.Counter
	accessLatency prometheus.Histogram
}

func NewWeatherMetrics(reg prometheus.Registerer) *WeatherMetrics {
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
	registerWeatherMetricsOnce.Do(func() {
		log.Println("Registering weather cache metrics")
		reg.MustRegister(cacheHits, cacheMisses, accessLatency)
	})

	return &WeatherMetrics{
		cacheHits:     cacheHits,
		cacheMisses:   cacheMisses,
		accessLatency: accessLatency,
	}
}

func (m *WeatherMetrics) CacheHit() {
	log.Println("Really inc")
	m.cacheHits.Inc()
}

func (m *WeatherMetrics) CacheMiss() {
	m.cacheMisses.Inc()
}

func (m *WeatherMetrics) CacheAccessLatency(seconds float64) {
	m.accessLatency.Observe(seconds)
}
