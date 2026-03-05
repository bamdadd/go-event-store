package store_test

import (
	"context"
	"testing"

	"github.com/bamdadd/go-event-store/internal/testutil"
	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/store/memory"
	"github.com/bamdadd/go-event-store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventStoreStreamPublishAndScan(t *testing.T) {
	adapter := memory.New()
	es := store.NewEventStore(adapter)
	ctx := context.Background()

	stream := es.Stream("orders", "order-1")
	events := testutil.NewTestEvents(3)

	stored, err := stream.Publish(ctx, events, store.NoCondition())
	require.NoError(t, err)
	assert.Len(t, stored, 3)

	var scanned []types.StoredEvent
	for e, scanErr := range stream.Scan(ctx) {
		require.NoError(t, scanErr)
		scanned = append(scanned, e)
	}
	assert.Len(t, scanned, 3)
}

func TestEventStoreStreamLatest(t *testing.T) {
	adapter := memory.New()
	es := store.NewEventStore(adapter)
	ctx := context.Background()

	stream := es.Stream("orders", "order-1")
	_, err := stream.Publish(ctx, testutil.NewTestEvents(3), store.NoCondition())
	require.NoError(t, err)

	latest, err := stream.Latest(ctx)
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, "event-3", latest.Name)
}

func TestEventStoreCategoryScan(t *testing.T) {
	adapter := memory.New()
	es := store.NewEventStore(adapter)
	ctx := context.Background()

	s1 := es.Stream("orders", "order-1")
	s2 := es.Stream("orders", "order-2")
	_, _ = s1.Publish(ctx, testutil.NewTestEvents(2), store.NoCondition())
	_, _ = s2.Publish(ctx, testutil.NewTestEvents(1), store.NoCondition())

	cat := es.Category("orders")
	var scanned []types.StoredEvent
	for e, scanErr := range cat.Scan(ctx) {
		require.NoError(t, scanErr)
		scanned = append(scanned, e)
	}
	assert.Len(t, scanned, 3)
}

func TestEventStoreLogScan(t *testing.T) {
	adapter := memory.New()
	es := store.NewEventStore(adapter)
	ctx := context.Background()

	s1 := es.Stream("orders", "order-1")
	s2 := es.Stream("payments", "pay-1")
	_, _ = s1.Publish(ctx, testutil.NewTestEvents(2), store.NoCondition())
	_, _ = s2.Publish(ctx, testutil.NewTestEvents(1), store.NoCondition())

	log := es.Log()
	var scanned []types.StoredEvent
	for e, scanErr := range log.Scan(ctx) {
		require.NoError(t, scanErr)
		scanned = append(scanned, e)
	}
	assert.Len(t, scanned, 3)
}

func TestEventStoreStreamIdentifier(t *testing.T) {
	adapter := memory.New()
	es := store.NewEventStore(adapter)
	stream := es.Stream("orders", "order-1")
	id := stream.Identifier()
	assert.Equal(t, "orders", id.Category)
	assert.Equal(t, "order-1", id.Stream)
}
