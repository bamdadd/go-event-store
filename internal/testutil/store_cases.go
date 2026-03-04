package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/logicblocks/event-store/store"
	"github.com/logicblocks/event-store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func RunStreamSaveTests(t *testing.T, h AdapterTestHarness) {
	t.Helper()

	t.Run("stores single event for later retrieval", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}
		events := NewTestEvents(1)

		stored, err := adapter.SaveToStream(ctx, target, events, store.NoCondition())
		require.NoError(t, err)
		assert.Len(t, stored, 1)
		assert.Equal(t, "event-1", stored[0].Name)
		assert.Equal(t, "orders", stored[0].Category)
		assert.Equal(t, "order-1", stored[0].Stream)
		assert.Equal(t, 0, stored[0].Position)
		assert.NotEmpty(t, stored[0].ID)
		assert.Greater(t, stored[0].SequenceNumber, 0)
	})

	t.Run("stores multiple events with sequential positions", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}

		stored, err := adapter.SaveToStream(ctx, target, NewTestEvents(3), store.NoCondition())
		require.NoError(t, err)
		assert.Len(t, stored, 3)
		for i, e := range stored {
			assert.Equal(t, i, e.Position)
			assert.Equal(t, fmt.Sprintf("event-%d", i+1), e.Name)
		}
	})

	t.Run("appends to existing stream with correct positions", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}

		_, err := adapter.SaveToStream(ctx, target, NewTestEvents(2), store.NoCondition())
		require.NoError(t, err)

		stored, err := adapter.SaveToStream(ctx, target, NewTestEvents(1), store.NoCondition())
		require.NoError(t, err)
		assert.Len(t, stored, 1)
		assert.Equal(t, 2, stored[0].Position)
	})

	t.Run("enforces StreamIsEmpty condition on non-empty stream", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}

		_, err := adapter.SaveToStream(ctx, target, NewTestEvents(1), store.NoCondition())
		require.NoError(t, err)

		_, err = adapter.SaveToStream(ctx, target, NewTestEvents(1), store.StreamIsEmpty())
		assert.ErrorIs(t, err, store.ErrUnmetWriteCondition)
	})

	t.Run("allows StreamIsEmpty condition on empty stream", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}

		stored, err := adapter.SaveToStream(ctx, target, NewTestEvents(1), store.StreamIsEmpty())
		require.NoError(t, err)
		assert.Len(t, stored, 1)
	})

	t.Run("enforces PositionIs condition on mismatched position", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}

		_, err := adapter.SaveToStream(ctx, target, NewTestEvents(1), store.NoCondition())
		require.NoError(t, err)

		wrongPos := 99
		_, err = adapter.SaveToStream(ctx, target, NewTestEvents(1), store.PositionIs(&wrongPos))
		assert.ErrorIs(t, err, store.ErrUnmetWriteCondition)
	})

	t.Run("allows PositionIs condition when position matches", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}

		stored, err := adapter.SaveToStream(ctx, target, NewTestEvents(1), store.NoCondition())
		require.NoError(t, err)

		correctPos := stored[0].Position
		stored2, err := adapter.SaveToStream(ctx, target, NewTestEvents(1), store.PositionIs(&correctPos))
		require.NoError(t, err)
		assert.Len(t, stored2, 1)
		assert.Equal(t, 1, stored2[0].Position)
	})

	t.Run("stores zero events as no-op", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}

		stored, err := adapter.SaveToStream(ctx, target, []types.NewEvent{}, store.NoCondition())
		require.NoError(t, err)
		assert.Empty(t, stored)
	})
}

func RunScanTests(t *testing.T, h AdapterTestHarness) {
	t.Helper()

	t.Run("scans all events from stream", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}
		_, err := adapter.SaveToStream(ctx, target, NewTestEvents(3), store.NoCondition())
		require.NoError(t, err)

		var results []types.StoredEvent
		for event, scanErr := range adapter.Scan(ctx, target) {
			require.NoError(t, scanErr)
			results = append(results, event)
		}
		assert.Len(t, results, 3)
	})

	t.Run("scans empty stream returns no events", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "nonexistent"}

		var results []types.StoredEvent
		for event, scanErr := range adapter.Scan(ctx, target) {
			require.NoError(t, scanErr)
			results = append(results, event)
		}
		assert.Empty(t, results)
	})

	t.Run("scans events from category", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		t1 := types.StreamIdentifier{Category: "orders", Stream: "order-1"}
		t2 := types.StreamIdentifier{Category: "orders", Stream: "order-2"}
		_, err := adapter.SaveToStream(ctx, t1, NewTestEvents(2), store.NoCondition())
		require.NoError(t, err)
		_, err = adapter.SaveToStream(ctx, t2, NewTestEvents(1), store.NoCondition())
		require.NoError(t, err)

		var results []types.StoredEvent
		for event, scanErr := range adapter.Scan(ctx, types.CategoryIdentifier{Category: "orders"}) {
			require.NoError(t, scanErr)
			results = append(results, event)
		}
		assert.Len(t, results, 3)
	})

	t.Run("scans all events from log", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		t1 := types.StreamIdentifier{Category: "orders", Stream: "order-1"}
		t2 := types.StreamIdentifier{Category: "payments", Stream: "pay-1"}
		_, err := adapter.SaveToStream(ctx, t1, NewTestEvents(2), store.NoCondition())
		require.NoError(t, err)
		_, err = adapter.SaveToStream(ctx, t2, NewTestEvents(1), store.NoCondition())
		require.NoError(t, err)

		var results []types.StoredEvent
		for event, scanErr := range adapter.Scan(ctx, types.LogIdentifier{}) {
			require.NoError(t, scanErr)
			results = append(results, event)
		}
		assert.Len(t, results, 3)
	})

	t.Run("scans with sequence number after constraint", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}
		stored, err := adapter.SaveToStream(ctx, target, NewTestEvents(5), store.NoCondition())
		require.NoError(t, err)

		afterSeq := stored[2].SequenceNumber
		var results []types.StoredEvent
		for event, scanErr := range adapter.Scan(ctx, target, store.WithSequenceNumberAfter(afterSeq)) {
			require.NoError(t, scanErr)
			results = append(results, event)
		}
		assert.Len(t, results, 2)
	})

	t.Run("scan respects context cancellation", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx, cancel := context.WithCancel(context.Background())
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}
		_, err := adapter.SaveToStream(ctx, target, NewTestEvents(10), store.NoCondition())
		require.NoError(t, err)

		count := 0
		for _, scanErr := range adapter.Scan(ctx, target) {
			count++
			if count == 3 {
				cancel()
			}
			if scanErr != nil {
				break
			}
		}
		assert.LessOrEqual(t, count, 5)
	})
}

func RunLatestTests(t *testing.T, h AdapterTestHarness) {
	t.Helper()

	t.Run("returns nil for empty stream", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "nonexistent"}
		latest, err := adapter.Latest(ctx, target)
		require.NoError(t, err)
		assert.Nil(t, latest)
	})

	t.Run("returns most recent event for stream", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		target := types.StreamIdentifier{Category: "orders", Stream: "order-1"}
		_, err := adapter.SaveToStream(ctx, target, NewTestEvents(3), store.NoCondition())
		require.NoError(t, err)

		latest, err := adapter.Latest(ctx, target)
		require.NoError(t, err)
		require.NotNil(t, latest)
		assert.Equal(t, "event-3", latest.Name)
		assert.Equal(t, 2, latest.Position)
	})

	t.Run("returns most recent event for category", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		t1 := types.StreamIdentifier{Category: "orders", Stream: "order-1"}
		t2 := types.StreamIdentifier{Category: "orders", Stream: "order-2"}
		_, err := adapter.SaveToStream(ctx, t1, NewTestEvents(2), store.NoCondition())
		require.NoError(t, err)
		_, err = adapter.SaveToStream(ctx, t2, NewTestEvents(1), store.NoCondition())
		require.NoError(t, err)

		latest, err := adapter.Latest(ctx, types.CategoryIdentifier{Category: "orders"})
		require.NoError(t, err)
		require.NotNil(t, latest)
	})

	t.Run("returns most recent event for log", func(t *testing.T) {
		t.Helper()
		adapter := h.ConstructAdapter(t, store.GuaranteeLog)
		defer h.ClearStorage(t)

		ctx := context.Background()
		t1 := types.StreamIdentifier{Category: "orders", Stream: "order-1"}
		_, err := adapter.SaveToStream(ctx, t1, NewTestEvents(3), store.NoCondition())
		require.NoError(t, err)

		latest, err := adapter.Latest(ctx, types.LogIdentifier{})
		require.NoError(t, err)
		require.NotNil(t, latest)
		assert.Equal(t, "event-3", latest.Name)
	})
}
