package processing

import (
	"math"
	"math/rand"
	"time"
)

type RetryPolicy interface {
	ShouldRetry(attempt int, err error) (bool, time.Duration)
}

type ExponentialBackoffPolicy struct {
	BaseDelay    time.Duration
	MaxDelay     time.Duration
	MaxAttempts  int
	JitterFactor float64
}

func (p ExponentialBackoffPolicy) ShouldRetry(attempt int, _ error) (bool, time.Duration) {
	if p.MaxAttempts > 0 && attempt >= p.MaxAttempts {
		return false, 0
	}
	delay := time.Duration(float64(p.BaseDelay) * math.Pow(2, float64(attempt)))
	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	if p.JitterFactor > 0 {
		jitter := time.Duration(float64(delay) * p.JitterFactor * (rand.Float64()*2 - 1))
		delay += jitter
	}
	return true, delay
}

func DefaultRetryPolicy() ExponentialBackoffPolicy {
	return ExponentialBackoffPolicy{
		BaseDelay:    100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		MaxAttempts:  10,
		JitterFactor: 0.25,
	}
}
