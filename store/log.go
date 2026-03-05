package store

import (
	"context"
	"iter"
	"log/slog"

	"github.com/bamdadd/go-event-store/types"
)

type EventLog struct {
	adapter    EventStorageAdapter
	identifier types.LogIdentifier
	logger     *slog.Logger
}

func (l *EventLog) Identifier() types.LogIdentifier { return l.identifier }

func (l *EventLog) Latest(ctx context.Context) (*types.StoredEvent, error) {
	return l.adapter.Latest(ctx, l.identifier)
}

func (l *EventLog) Scan(ctx context.Context, opts ...ScanOption) iter.Seq2[types.StoredEvent, error] {
	return l.adapter.Scan(ctx, l.identifier, opts...)
}
