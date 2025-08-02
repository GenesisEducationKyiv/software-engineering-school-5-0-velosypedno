package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var registerConsumerMetricsOnce sync.Once

type ConsumerMetrics struct {
	received   *prometheus.CounterVec
	acked      *prometheus.CounterVec
	notAcked   *prometheus.CounterVec
	handleTime *prometheus.HistogramVec
}

func NewConsumerMetrics(reg prometheus.Registerer) *ConsumerMetrics {
	m := &ConsumerMetrics{
		received: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "consumer_messages_total",
				Help: "Total number of received messages",
			},
			[]string{"consumer"},
		),
		acked: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "consumer_acked_total",
				Help: "Number of successfully acked messages",
			},
			[]string{"consumer"},
		),
		notAcked: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "consumer_not_acked_total",
				Help: "Number of messages that were not acked",
			},
			[]string{"consumer"},
		),
		handleTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "consumer_duration_seconds",
				Help:    "Time taken to handle a message",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"consumer"},
		),
	}

	registerConsumerMetricsOnce.Do(func() {
		reg.MustRegister(m.received, m.acked, m.notAcked, m.handleTime)
	})

	return m
}

func (m *ConsumerMetrics) IncReceived(consumer string) {
	m.received.WithLabelValues(consumer).Inc()
}

func (m *ConsumerMetrics) IncAcked(consumer string) {
	m.acked.WithLabelValues(consumer).Inc()
}

func (m *ConsumerMetrics) IncNotAcked(consumer string) {
	m.notAcked.WithLabelValues(consumer).Inc()
}

func (m *ConsumerMetrics) ObserveHandleDuration(consumer string, seconds float64) {
	m.handleTime.WithLabelValues(consumer).Observe(seconds)
}
