package types

import (
	"encoding/json"
	"time"

	"github.com/logicblocks/event-store/internal/clock"
)

type NewEvent struct {
	Name       string
	Payload    any
	ObservedAt time.Time
	OccurredAt time.Time
}

type NewEventOption func(*newEventConfig)

type newEventConfig struct {
	observedAt *time.Time
	occurredAt *time.Time
	clock      clock.Clock
}

func WithObservedAt(t time.Time) NewEventOption {
	return func(cfg *newEventConfig) { cfg.observedAt = &t }
}

func WithOccurredAt(t time.Time) NewEventOption {
	return func(cfg *newEventConfig) { cfg.occurredAt = &t }
}

func WithClock(c clock.Clock) NewEventOption {
	return func(cfg *newEventConfig) { cfg.clock = c }
}

func NewNewEvent(name string, payload any, opts ...NewEventOption) NewEvent {
	cfg := &newEventConfig{clock: clock.SystemClock{}}
	for _, o := range opts {
		o(cfg)
	}
	now := cfg.clock.Now()
	observedAt := now
	if cfg.observedAt != nil {
		observedAt = *cfg.observedAt
	}
	occurredAt := observedAt
	if cfg.occurredAt != nil {
		occurredAt = *cfg.occurredAt
	}
	return NewEvent{
		Name:       name,
		Payload:    payload,
		ObservedAt: observedAt,
		OccurredAt: occurredAt,
	}
}

type StoredEvent struct {
	ID             string
	Name           string
	Stream         string
	Category       string
	Position       int
	SequenceNumber int
	Payload        json.RawMessage
	ObservedAt     time.Time
	OccurredAt     time.Time
}
