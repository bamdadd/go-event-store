package types_test

import (
	"testing"
	"time"

	"github.com/logicblocks/event-store/internal/clock"
	"github.com/logicblocks/event-store/types"
	"github.com/stretchr/testify/assert"
)

func TestNewNewEventUsesSystemClockByDefault(t *testing.T) {
	before := time.Now().UTC()
	event := types.NewNewEvent("order-placed", map[string]any{"id": "123"})
	after := time.Now().UTC()

	assert.Equal(t, "order-placed", event.Name)
	assert.False(t, event.ObservedAt.Before(before))
	assert.False(t, event.ObservedAt.After(after))
	assert.Equal(t, event.ObservedAt, event.OccurredAt)
}

func TestNewNewEventUsesInjectedClock(t *testing.T) {
	fixed := time.Date(2026, 3, 4, 12, 0, 0, 0, time.UTC)
	event := types.NewNewEvent("order-placed", nil, types.WithClock(clock.StaticClock{Time: fixed}))

	assert.Equal(t, fixed, event.ObservedAt)
	assert.Equal(t, fixed, event.OccurredAt)
}

func TestNewNewEventWithExplicitTimestamps(t *testing.T) {
	observed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	occurred := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	event := types.NewNewEvent("x", nil, types.WithObservedAt(observed), types.WithOccurredAt(occurred))

	assert.Equal(t, observed, event.ObservedAt)
	assert.Equal(t, occurred, event.OccurredAt)
}

func TestNewNewEventOccurredAtDefaultsToObservedAt(t *testing.T) {
	observed := time.Date(2026, 6, 15, 10, 30, 0, 0, time.UTC)
	event := types.NewNewEvent("x", nil, types.WithObservedAt(observed))

	assert.Equal(t, observed, event.ObservedAt)
	assert.Equal(t, observed, event.OccurredAt)
}

func TestNewNewEventPreservesPayload(t *testing.T) {
	payload := map[string]any{"amount": 42.5, "currency": "GBP"}
	event := types.NewNewEvent("payment-received", payload)
	assert.Equal(t, payload, event.Payload)
}

func TestNewNewEventWithNilPayload(t *testing.T) {
	event := types.NewNewEvent("heartbeat", nil)
	assert.Nil(t, event.Payload)
}
