package store

import (
	"context"
	"iter"
	"log/slog"

	"github.com/logicblocks/event-store/types"
)

type EventCategory struct {
	adapter    EventStorageAdapter
	identifier types.CategoryIdentifier
	logger     *slog.Logger
}

func (c *EventCategory) Identifier() types.CategoryIdentifier { return c.identifier }

func (c *EventCategory) Latest(ctx context.Context) (*types.StoredEvent, error) {
	return c.adapter.Latest(ctx, c.identifier)
}

func (c *EventCategory) Scan(ctx context.Context, opts ...ScanOption) iter.Seq2[types.StoredEvent, error] {
	return c.adapter.Scan(ctx, c.identifier, opts...)
}

func (c *EventCategory) Publish(
	ctx context.Context,
	streams map[string]StreamPublishDefinition,
) (map[string][]types.StoredEvent, error) {
	return c.adapter.SaveToCategory(ctx, c.identifier, streams)
}
