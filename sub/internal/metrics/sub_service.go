package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerSubscriptionMetricsOnce sync.Once
)

type SubscriptionMetrics struct {
	subscribeTotal      prometheus.Counter
	subscribeErrTotal   prometheus.Counter
	activateTotal       prometheus.Counter
	activateErrTotal    prometheus.Counter
	unsubscribeTotal    prometheus.Counter
	unsubscribeErrTotal prometheus.Counter
}

func NewSubscriptionMetrics(reg prometheus.Registerer) *SubscriptionMetrics {
	metrics := &SubscriptionMetrics{
		subscribeTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "subscription_subscribe_total",
			Help: "Total number of successful subscriptions",
		}),
		subscribeErrTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "subscription_subscribe_error_internal_total",
			Help: "Total number of internal errors during subscription",
		}),
		activateTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "subscription_activate_total",
			Help: "Total number of successful activations",
		}),
		activateErrTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "subscription_activate_error_internal_total",
			Help: "Total number of internal errors during activation",
		}),
		unsubscribeTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "subscription_unsubscribe_total",
			Help: "Total number of successful unsubscriptions",
		}),
		unsubscribeErrTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "subscription_unsubscribe_error_internal_total",
			Help: "Total number of internal errors during unsubscription",
		}),
	}

	registerSubscriptionMetricsOnce.Do(func() {
		reg.MustRegister(
			metrics.subscribeTotal,
			metrics.subscribeErrTotal,
			metrics.activateTotal,
			metrics.activateErrTotal,
			metrics.unsubscribeTotal,
			metrics.unsubscribeErrTotal,
		)
	})

	return metrics
}

func (m *SubscriptionMetrics) IncSubscribe() {
	m.subscribeTotal.Inc()
}
func (m *SubscriptionMetrics) IncSubscribeError() {
	m.subscribeErrTotal.Inc()
}
func (m *SubscriptionMetrics) IncActivate() {
	m.activateTotal.Inc()
}
func (m *SubscriptionMetrics) IncActivateError() {
	m.activateErrTotal.Inc()
}
func (m *SubscriptionMetrics) IncUnsubscribe() {
	m.unsubscribeTotal.Inc()
}
func (m *SubscriptionMetrics) IncUnsubscribeError() {
	m.unsubscribeErrTotal.Inc()
}
