package cb

import (
	"sync"
	"time"
)

// State represents the current state of the CircuitBreaker.
type State int

const (
	Closed   State = iota // All requests are allowed.
	HalfOpen              // Limited requests are allowed to test recovery.
	Open                  // All requests are blocked for a timeout duration.
)

// CircuitBreaker is a simple thread-safe implementation of the Circuit Breaker pattern.
// It prevents system overload by blocking calls to a failing resource for a specified duration.
type CircuitBreaker struct {
	mu  sync.Mutex
	Now func() time.Time // Injected clock function for testing or time mocking.

	state             State
	timeout           time.Duration // Duration the circuit remains open before attempting recovery.
	maxFails          int           // Number of consecutive failures allowed before opening the circuit.
	lastFailTime      time.Time     // Timestamp of the most recent failure.
	openedUntil       time.Time     // If open, the circuit remains blocked until this time.
	failCount         int           // Consecutive failure counter.
	recoverCount      int           // Number of successful calls during HalfOpen state.
	attemptsToRecover int           // Required successes in HalfOpen to transition back to Closed.
}

// NewCircuitBreaker initializes a new CircuitBreaker with specified timeout, failure threshold,
// and the number of successful attempts required to recover.
func NewCircuitBreaker(timeout time.Duration, maxFails int, attemptsToRecover int) *CircuitBreaker {
	now := time.Now()
	return &CircuitBreaker{
		state:             Closed,
		timeout:           timeout,
		maxFails:          maxFails,
		attemptsToRecover: attemptsToRecover,
		lastFailTime:      now,
		openedUntil:       now,
		failCount:         0,
		recoverCount:      0,
		Now:               time.Now,
	}
}

// defineStatus evaluates and updates the internal state of the circuit breaker.
// It handles transitions based on timeouts, failure counts, and recovery attempts.
func (cb *CircuitBreaker) defineStatus() {
	cb.openedUntil = cb.lastFailTime.Add(cb.timeout)
	now := cb.Now()

	switch cb.state {
	case Closed:
		cb.recoverCount = 0

		// Reset fail count if last failure is older than timeout.
		if cb.lastFailTime.Add(cb.timeout).Before(now) {
			cb.failCount = 0
		}

		// Transition to Open if failure threshold is exceeded.
		if cb.failCount >= cb.maxFails {
			cb.failCount = 0
			cb.state = Open
		}

	case HalfOpen:
		// If enough successful attempts in HalfOpen, restore to Closed.
		if cb.recoverCount == cb.attemptsToRecover {
			cb.state = Closed
			cb.recoverCount = 0
			cb.failCount = 0
		}

	case Open:
		// If timeout has expired, allow test calls by entering HalfOpen.
		if now.After(cb.openedUntil) {
			cb.state = HalfOpen
			cb.recoverCount = 0
			cb.failCount = 0
		}
	}
}

// Allowed returns true if the circuit is not in the Open state.
// This method should be called before executing any protected logic.
func (cb *CircuitBreaker) Allowed() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.defineStatus()
	return cb.state != Open
}

// Fail reports a failed operation to the circuit breaker.
// It triggers state transitions and timeout timers as needed.
func (cb *CircuitBreaker) Fail() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.defineStatus()

	now := cb.Now()

	switch cb.state {
	case Closed:
		cb.lastFailTime = now
		cb.failCount++

	case HalfOpen:
		// Any failure during recovery puts the breaker back to Open immediately.
		cb.lastFailTime = now
		cb.state = Open
	}
}

// Success reports a successful operation to the circuit breaker.
// It is only relevant during the HalfOpen state, where it counts toward recovery.
func (cb *CircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.defineStatus()

	if cb.state == HalfOpen {
		cb.recoverCount++
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.defineStatus()
	return cb.state
}
