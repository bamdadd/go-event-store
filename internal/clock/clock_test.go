package clock_test

import (
	"testing"
	"time"

	"github.com/logicblocks/event-store/internal/clock"
	"github.com/stretchr/testify/assert"
)

func TestStaticClockReturnsFixedTime(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c := clock.StaticClock{Time: fixed}
	assert.Equal(t, fixed, c.Now())
}

func TestSystemClockReturnsCurrentTime(t *testing.T) {
	c := clock.SystemClock{}
	before := time.Now().UTC()
	now := c.Now()
	after := time.Now().UTC()
	assert.False(t, now.Before(before))
	assert.False(t, now.After(after))
}
