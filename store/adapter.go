package store

import (
	"context"
	"iter"

	"github.com/bamdadd/go-event-store/types"
)

type StreamPublishDefinition struct {
	Events    []types.NewEvent
	Condition WriteCondition
}

type EventStorageAdapter interface {
	SaveToStream(
		ctx context.Context,
		target types.StreamIdentifier,
		events []types.NewEvent,
		condition WriteCondition,
	) ([]types.StoredEvent, error)

	SaveToCategory(
		ctx context.Context,
		target types.CategoryIdentifier,
		streams map[string]StreamPublishDefinition,
	) (map[string][]types.StoredEvent, error)

	Latest(ctx context.Context, target types.Targetable) (*types.StoredEvent, error)

	Scan(ctx context.Context, target types.Targetable, opts ...ScanOption) iter.Seq2[types.StoredEvent, error]
}
