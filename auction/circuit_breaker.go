package auction

import (
	"context"
	"errors"
	"sync"
	"time"
)

type CircuitBreaker struct {
	mu              sync.Mutex
	failureCount    int
	failureThreshold int
	lastFailureTime time.Time
	state           state 
	cooldownPeriod  time.Duration
}

type state int

const (
	stateClosed state = iota
	stateOpen            
	stateHalfOpen        
)

func NewCircuitBreaker(failureThreshold int, cooldownPeriod time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		cooldownPeriod:   cooldownPeriod,
		state:            stateClosed,
	}
}

func (cb *CircuitBreaker) Call(ctx context.Context, f func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case stateClosed:
		result, err := f()
		if err != nil {
				cb.recordFailure()
		} else {
				cb.reset()
		}
		return result, err
	case stateOpen:
		if time.Since(cb.lastFailureTime) > cb.cooldownPeriod {
			cb.state = stateHalfOpen 
		} else {
			return nil, errors.New("circuit breaker is open")
		}
	case stateHalfOpen:
		result, err := f()
		if err != nil {
			cb.recordFailure() 
		} else {
			cb.state = stateClosed 
		}
		return result, err 
	}

  panic("invalid circuit breaker state") // For safety
}

func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	if cb.failureCount >= cb.failureThreshold {
		cb.state = stateOpen 
	}
}

func (cb *CircuitBreaker) reset() {
    cb.failureCount = 0
}
