package component_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bamdadd/go-event-store/internal/testutil"
	"github.com/bamdadd/go-event-store/projection"
	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/store/memory"
	"github.com/bamdadd/go-event-store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type orderTotal struct {
	Count int
	Total float64
}

func TestEndToEndWriteProjectRead(t *testing.T) {
	adapter := memory.New()
	es := store.NewEventStore(adapter)
	ctx := context.Background()

	stream := es.Stream("orders", "order-42")

	events := []types.NewEvent{
		testutil.NewTestEvent("item-added", map[string]any{"price": 9.99}),
		testutil.NewTestEvent("item-added", map[string]any{"price": 14.50}),
		testutil.NewTestEvent("item-removed", map[string]any{"price": 9.99}),
	}

	_, err := stream.Publish(ctx, events, store.StreamIsEmpty())
	require.NoError(t, err)

	projector := projection.NewProjector(func() orderTotal { return orderTotal{} }).
		On("item-added", func(s orderTotal, e types.StoredEvent) (orderTotal, error) {
			var payload map[string]any
			_ = json.Unmarshal(e.Payload, &payload)
			s.Count++
			s.Total += payload["price"].(float64)
			return s, nil
		}).
		On("item-removed", func(s orderTotal, e types.StoredEvent) (orderTotal, error) {
			var payload map[string]any
			_ = json.Unmarshal(e.Payload, &payload)
			s.Count--
			s.Total -= payload["price"].(float64)
			return s, nil
		})

	state, err := projector.Project(ctx, stream.Scan(ctx))
	require.NoError(t, err)
	assert.Equal(t, 1, state.Count)
	assert.InDelta(t, 14.50, state.Total, 0.01)
}
