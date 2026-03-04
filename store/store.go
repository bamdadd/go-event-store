package store

import (
	"log/slog"

	"github.com/logicblocks/event-store/types"
)

type EventStore struct {
	adapter EventStorageAdapter
	logger  *slog.Logger
}

type EventStoreOption func(*EventStore)

func WithLogger(l *slog.Logger) EventStoreOption {
	return func(es *EventStore) { es.logger = l }
}

func NewEventStore(adapter EventStorageAdapter, opts ...EventStoreOption) *EventStore {
	es := &EventStore{adapter: adapter, logger: slog.Default()}
	for _, o := range opts {
		o(es)
	}
	return es
}

func (es *EventStore) Stream(category, stream string) *EventStream {
	return &EventStream{
		adapter:    es.adapter,
		identifier: types.StreamIdentifier{Category: category, Stream: stream},
		logger:     es.logger,
	}
}

func (es *EventStore) Category(category string) *EventCategory {
	return &EventCategory{
		adapter:    es.adapter,
		identifier: types.CategoryIdentifier{Category: category},
		logger:     es.logger,
	}
}

func (es *EventStore) Log() *EventLog {
	return &EventLog{
		adapter:    es.adapter,
		identifier: types.LogIdentifier{},
		logger:     es.logger,
	}
}
