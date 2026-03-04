package store

import (
	"context"
	"iter"
	"log/slog"

	"github.com/logicblocks/event-store/types"
)

type EventStream struct {
	adapter    EventStorageAdapter
	identifier types.StreamIdentifier
	logger     *slog.Logger
}

func (s *EventStream) Identifier() types.StreamIdentifier { return s.identifier }

func (s *EventStream) Latest(ctx context.Context) (*types.StoredEvent, error) {
	return s.adapter.Latest(ctx, s.identifier)
}

func (s *EventStream) Scan(ctx context.Context, opts ...ScanOption) iter.Seq2[types.StoredEvent, error] {
	return s.adapter.Scan(ctx, s.identifier, opts...)
}

func (s *EventStream) Publish(
	ctx context.Context,
	events []types.NewEvent,
	condition WriteCondition,
) ([]types.StoredEvent, error) {
	return s.adapter.SaveToStream(ctx, s.identifier, events, condition)
}
