//go:build integration

package main

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/bamdadd/go-event-store/projection"
	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/store/postgres"
	"github.com/bamdadd/go-event-store/types"
)

func setupPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	initScript, err := filepath.Abs("../../db/init.sql")
	require.NoError(t, err)

	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("event_store_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.WithInitScripts(initScript),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { container.Terminate(context.Background()) })

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return pool
}

func TestPostgresExample_PublishAndProject(t *testing.T) {
	pool := setupPostgres(t)
	ctx := context.Background()

	adapter := postgres.New(pool)
	es := store.NewEventStore(adapter)
	stream := es.Stream("orders", "order-100")

	stored, err := stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("OrderCreated", map[string]any{"customer": "alice"}),
		types.NewNewEvent("ItemAdded", map[string]any{"item": "widget", "price": 29.99}),
		types.NewNewEvent("ItemAdded", map[string]any{"item": "gadget", "price": 49.99}),
	}, store.StreamIsEmpty())
	require.NoError(t, err)
	assert.Len(t, stored, 3)

	projector := projection.NewProjector(func() OrderSummary { return OrderSummary{} },
		projection.WithMissingHandlerBehaviour[OrderSummary](projection.MissingHandlerBehaviourIgnore),
	)
	projector.
		On("ItemAdded", func(s OrderSummary, e types.StoredEvent) (OrderSummary, error) {
			var p map[string]any
			json.Unmarshal(e.Payload, &p)
			s.Items++
			s.Total += p["price"].(float64)
			return s, nil
		}).
		On("ItemRemoved", func(s OrderSummary, e types.StoredEvent) (OrderSummary, error) {
			var p map[string]any
			json.Unmarshal(e.Payload, &p)
			s.Items--
			s.Total -= p["price"].(float64)
			return s, nil
		})

	summary, err := projector.Project(ctx, stream.Scan(ctx))
	require.NoError(t, err)
	assert.Equal(t, 2, summary.Items)
	assert.InDelta(t, 79.98, summary.Total, 0.01)
}

func TestPostgresExample_OptimisticConcurrency(t *testing.T) {
	pool := setupPostgres(t)
	ctx := context.Background()

	adapter := postgres.New(pool)
	es := store.NewEventStore(adapter)
	stream := es.Stream("orders", "order-200")

	stored, err := stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("OrderCreated", map[string]any{"customer": "bob"}),
	}, store.StreamIsEmpty())
	require.NoError(t, err)

	// Append with correct position
	lastPos := stored[len(stored)-1].Position
	_, err = stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("ItemAdded", map[string]any{"item": "thing", "price": 10.0}),
	}, store.PositionIs(&lastPos))
	require.NoError(t, err)

	// Append with stale position should fail
	stalePos := 0
	_, err = stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("ItemAdded", map[string]any{"item": "other", "price": 5.0}),
	}, store.PositionIs(&stalePos))

	var condErr *store.UnmetWriteConditionError
	assert.True(t, errors.As(err, &condErr))
}

func TestPostgresExample_StreamIsEmptyRejectsSecondWrite(t *testing.T) {
	pool := setupPostgres(t)
	ctx := context.Background()

	adapter := postgres.New(pool)
	es := store.NewEventStore(adapter)
	stream := es.Stream("orders", "order-300")

	_, err := stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("OrderCreated", map[string]any{}),
	}, store.StreamIsEmpty())
	require.NoError(t, err)

	_, err = stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("OrderCreated", map[string]any{}),
	}, store.StreamIsEmpty())

	var condErr *store.UnmetWriteConditionError
	assert.True(t, errors.As(err, &condErr))
}

func TestPostgresExample_ScanAfterPublish(t *testing.T) {
	pool := setupPostgres(t)
	ctx := context.Background()

	adapter := postgres.New(pool)
	es := store.NewEventStore(adapter)
	stream := es.Stream("orders", "order-400")

	_, err := stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("OrderCreated", map[string]any{}),
		types.NewNewEvent("ItemAdded", map[string]any{"item": "a", "price": 1.0}),
		types.NewNewEvent("ItemAdded", map[string]any{"item": "b", "price": 2.0}),
	}, store.StreamIsEmpty())
	require.NoError(t, err)

	var events []types.StoredEvent
	for e, scanErr := range stream.Scan(ctx) {
		require.NoError(t, scanErr)
		events = append(events, e)
	}

	assert.Len(t, events, 3)
	assert.Equal(t, "OrderCreated", events[0].Name)
	assert.Equal(t, 0, events[0].Position)
	assert.Equal(t, 2, events[2].Position)
}
