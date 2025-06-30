package cb

import (
	"sync"
	"time"
)

// CircuitBreaker is a simple thread-safe circuit breaker implementation.
// It temporarily blocks execution after too many consecutive failures.
type CircuitBreaker struct {
	mu  sync.Mutex
	Now func() time.Time

	timeout      time.Duration // How long the circuit remains open after tripping
	maxFails     int           // Number of failures allowed before opening the circuit
	lastFailTime time.Time     // Time of the last recorded failure
	closedUntil  time.Time     // The circuit remains open until this time
	failCount    int           // Consecutive failure count
}

// NewCircuitBreaker creates a new CircuitBreaker with the given timeout and maximum failures.
func NewCircuitBreaker(timeout time.Duration, maxFails int) *CircuitBreaker {
	now := time.Now()
	return &CircuitBreaker{
		timeout:      timeout,
		maxFails:     maxFails,
		lastFailTime: now,
		closedUntil:  now,
		failCount:    0,
		Now:          time.Now,
	}
}

// IsClosed returns true if the circuit is closed (i.e., calls are allowed).
func (cb *CircuitBreaker) IsClosed() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return cb.Now().After(cb.closedUntil)
}

// Fail should be called after a failed operation.
// It increments the failure count and opens the circuit if the threshold is exceeded.
func (cb *CircuitBreaker) Fail() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := cb.Now()

	// Reset fail count if last failure was too long ago
	if cb.lastFailTime.Add(cb.timeout).Before(now) {
		cb.failCount = 0
	}

	cb.lastFailTime = now
	cb.failCount++

	if cb.failCount >= cb.maxFails {
		cb.closedUntil = now.Add(cb.timeout)
		cb.failCount = 0
	}
}
