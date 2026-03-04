package memory_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/logicblocks/event-store/projection"
	projmem "github.com/logicblocks/event-store/projection/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndFindOne(t *testing.T) {
	adapter := projmem.New()
	ctx := context.Background()

	p := projection.Projection{
		ID:    "order-1",
		Name:  "order-summary",
		State: json.RawMessage(`{"total": 42}`),
	}
	err := adapter.Save(ctx, p)
	require.NoError(t, err)

	found, err := adapter.FindOne(ctx, "order-summary", "order-1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "order-1", found.ID)
	assert.Equal(t, json.RawMessage(`{"total": 42}`), found.State)
}

func TestFindOneReturnsNilWhenNotFound(t *testing.T) {
	adapter := projmem.New()
	ctx := context.Background()

	found, err := adapter.FindOne(ctx, "order-summary", "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestSaveOverwritesExisting(t *testing.T) {
	adapter := projmem.New()
	ctx := context.Background()

	p1 := projection.Projection{ID: "order-1", Name: "order-summary", State: json.RawMessage(`{"total": 1}`)}
	p2 := projection.Projection{ID: "order-1", Name: "order-summary", State: json.RawMessage(`{"total": 99}`)}

	require.NoError(t, adapter.Save(ctx, p1))
	require.NoError(t, adapter.Save(ctx, p2))

	found, err := adapter.FindOne(ctx, "order-summary", "order-1")
	require.NoError(t, err)
	assert.Equal(t, json.RawMessage(`{"total": 99}`), found.State)
}

func TestFindManyReturnsAllForName(t *testing.T) {
	adapter := projmem.New()
	ctx := context.Background()

	require.NoError(t, adapter.Save(ctx, projection.Projection{ID: "1", Name: "orders"}))
	require.NoError(t, adapter.Save(ctx, projection.Projection{ID: "2", Name: "orders"}))
	require.NoError(t, adapter.Save(ctx, projection.Projection{ID: "3", Name: "payments"}))

	results, err := adapter.FindMany(ctx, "orders")
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestFindManyReturnsEmptyForUnknownName(t *testing.T) {
	adapter := projmem.New()
	ctx := context.Background()

	results, err := adapter.FindMany(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, results)
}
