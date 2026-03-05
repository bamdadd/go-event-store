//go:build integration

package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bamdadd/go-event-store/projection"
	projmem "github.com/bamdadd/go-event-store/projection/memory"
	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/store/memory"
	"github.com/bamdadd/go-event-store/types"
)

func TestInMemoryExample_PublishAndProject(t *testing.T) {
	ctx := context.Background()

	eventAdapter := memory.New()
	es := store.NewEventStore(eventAdapter)

	stream := es.Stream("accounts", "alice")

	_, err := stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("AccountOpened", map[string]any{"owner": "Alice"}),
		types.NewNewEvent("MoneyDeposited", map[string]any{"amount": 500.0}),
		types.NewNewEvent("MoneyWithdrawn", map[string]any{"amount": 50.0}),
	}, store.StreamIsEmpty())
	require.NoError(t, err)

	projector := projection.NewProjector(func() AccountBalance { return AccountBalance{} })
	projector.
		On("AccountOpened", func(s AccountBalance, e types.StoredEvent) (AccountBalance, error) {
			var p map[string]any
			json.Unmarshal(e.Payload, &p)
			s.Owner = p["owner"].(string)
			return s, nil
		}).
		On("MoneyDeposited", func(s AccountBalance, e types.StoredEvent) (AccountBalance, error) {
			var p map[string]any
			json.Unmarshal(e.Payload, &p)
			s.Balance += p["amount"].(float64)
			return s, nil
		}).
		On("MoneyWithdrawn", func(s AccountBalance, e types.StoredEvent) (AccountBalance, error) {
			var p map[string]any
			json.Unmarshal(e.Payload, &p)
			s.Balance -= p["amount"].(float64)
			return s, nil
		})

	state, err := projector.Project(ctx, stream.Scan(ctx))
	require.NoError(t, err)

	assert.Equal(t, "Alice", state.Owner)
	assert.InDelta(t, 450.0, state.Balance, 0.01)
}

func TestInMemoryExample_ProjectionStore(t *testing.T) {
	ctx := context.Background()

	projAdapter := projmem.New()
	projStore := projection.NewProjectionStore(projAdapter)

	stateJSON, _ := json.Marshal(AccountBalance{Owner: "Alice", Balance: 450})
	err := projStore.Save(ctx, projection.Projection{
		ID:    "alice",
		Name:  "account-balance",
		State: stateJSON,
	})
	require.NoError(t, err)

	proj, err := projStore.FindOne(ctx, "account-balance", "alice")
	require.NoError(t, err)
	require.NotNil(t, proj)

	var restored AccountBalance
	json.Unmarshal(proj.State, &restored)
	assert.Equal(t, "Alice", restored.Owner)
	assert.InDelta(t, 450.0, restored.Balance, 0.01)
}

func TestInMemoryExample_CategoryScan(t *testing.T) {
	ctx := context.Background()

	eventAdapter := memory.New()
	es := store.NewEventStore(eventAdapter)

	_, err := es.Stream("accounts", "alice").Publish(ctx, []types.NewEvent{
		types.NewNewEvent("MoneyDeposited", map[string]any{"amount": 100.0}),
	}, store.StreamIsEmpty())
	require.NoError(t, err)

	_, err = es.Stream("accounts", "bob").Publish(ctx, []types.NewEvent{
		types.NewNewEvent("MoneyDeposited", map[string]any{"amount": 200.0}),
	}, store.StreamIsEmpty())
	require.NoError(t, err)

	var count int
	for _, scanErr := range es.Category("accounts").Scan(ctx) {
		require.NoError(t, scanErr)
		count++
	}
	assert.Equal(t, 2, count)
}

func TestInMemoryExample_LogScan(t *testing.T) {
	ctx := context.Background()

	eventAdapter := memory.New()
	es := store.NewEventStore(eventAdapter)

	_, err := es.Stream("accounts", "alice").Publish(ctx, []types.NewEvent{
		types.NewNewEvent("MoneyDeposited", map[string]any{"amount": 100.0}),
	}, store.StreamIsEmpty())
	require.NoError(t, err)

	_, err = es.Stream("orders", "order-1").Publish(ctx, []types.NewEvent{
		types.NewNewEvent("OrderCreated", map[string]any{}),
	}, store.StreamIsEmpty())
	require.NoError(t, err)

	var count int
	for _, scanErr := range es.Log().Scan(ctx) {
		require.NoError(t, scanErr)
		count++
	}
	assert.Equal(t, 2, count)
}
