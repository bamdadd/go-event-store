package processing_test

import (
	"errors"
	"testing"
	"time"

	"github.com/logicblocks/event-store/processing"
	"github.com/stretchr/testify/assert"
)

func TestExponentialBackoffFirstAttempt(t *testing.T) {
	p := processing.ExponentialBackoffPolicy{
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		MaxAttempts: 5,
	}
	ok, delay := p.ShouldRetry(0, errors.New("err"))
	assert.True(t, ok)
	assert.Equal(t, 100*time.Millisecond, delay)
}

func TestExponentialBackoffDoubles(t *testing.T) {
	p := processing.ExponentialBackoffPolicy{
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		MaxAttempts: 10,
	}
	_, delay1 := p.ShouldRetry(1, nil)
	_, delay2 := p.ShouldRetry(2, nil)
	assert.Equal(t, 200*time.Millisecond, delay1)
	assert.Equal(t, 400*time.Millisecond, delay2)
}

func TestExponentialBackoffCapsAtMaxDelay(t *testing.T) {
	p := processing.ExponentialBackoffPolicy{
		BaseDelay:   1 * time.Second,
		MaxDelay:    5 * time.Second,
		MaxAttempts: 20,
	}
	_, delay := p.ShouldRetry(10, nil)
	assert.LessOrEqual(t, delay, 5*time.Second)
}

func TestExponentialBackoffStopsAtMaxAttempts(t *testing.T) {
	p := processing.ExponentialBackoffPolicy{
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		MaxAttempts: 3,
	}
	ok, _ := p.ShouldRetry(3, nil)
	assert.False(t, ok)
}

func TestExponentialBackoffWithJitter(t *testing.T) {
	p := processing.ExponentialBackoffPolicy{
		BaseDelay:    1 * time.Second,
		MaxDelay:     10 * time.Second,
		MaxAttempts:  5,
		JitterFactor: 0.25,
	}
	_, delay := p.ShouldRetry(0, nil)
	assert.InDelta(t, float64(1*time.Second), float64(delay), float64(250*time.Millisecond))
}
