//go:build integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcnats "github.com/testcontainers/testcontainers-go/modules/nats"

	wmadapter "github.com/bamdadd/go-event-store/processing/watermill"
	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/store/memory"
	"github.com/bamdadd/go-event-store/types"
)

func setupNATS(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	container, err := tcnats.Run(ctx, "nats:2")
	require.NoError(t, err)
	t.Cleanup(func() { container.Terminate(context.Background()) })

	url, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	return url
}

func TestWatermillNATS_PublishAndConsume(t *testing.T) {
	natsURL := setupNATS(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wmLogger := watermill.NewStdLogger(false, false)

	publisher, err := nats.NewPublisher(
		nats.PublisherConfig{
			URL: natsURL,
			JetStream: nats.JetStreamConfig{
				AutoProvision: true,
			},
		},
		wmLogger,
	)
	require.NoError(t, err)
	defer publisher.Close()

	subscriber, err := nats.NewSubscriber(
		nats.SubscriberConfig{
			URL: natsURL,
			JetStream: nats.JetStreamConfig{
				AutoProvision: true,
			},
		},
		wmLogger,
	)
	require.NoError(t, err)
	defer subscriber.Close()

	// Set up event store and publish events
	eventAdapter := memory.New()
	es := store.NewEventStore(eventAdapter)
	stream := es.Stream("inventory", "warehouse-1")

	stored, err := stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("StockReceived", map[string]any{"item": "widget", "quantity": 100}),
		types.NewNewEvent("StockReceived", map[string]any{"item": "gadget", "quantity": 50}),
		types.NewNewEvent("StockShipped", map[string]any{"item": "widget", "quantity": 30}),
	}, store.StreamIsEmpty())
	require.NoError(t, err)

	// Set up async consumer
	inventory := &InventoryState{stock: make(map[string]int)}
	processor := &inventoryProcessor{state: inventory}
	consumer := wmadapter.NewConsumer(subscriber, topic, processor)

	go func() {
		consumer.Start(ctx)
	}()

	// Give consumer time to subscribe
	time.Sleep(500 * time.Millisecond)

	// Forward events to NATS
	for _, e := range stored {
		msg, err := wmadapter.StoredEventToMessage(e)
		require.NoError(t, err)
		err = publisher.Publish(topic, msg)
		require.NoError(t, err)
	}

	// Wait for async processing with timeout
	require.Eventually(t, func() bool {
		snap := inventory.Snapshot()
		return snap["widget"] == 70 && snap["gadget"] == 50
	}, 10*time.Second, 100*time.Millisecond,
		"expected widget=70, gadget=50, got: %v", inventory.Snapshot())

	snap := inventory.Snapshot()
	assert.Equal(t, 70, snap["widget"])
	assert.Equal(t, 50, snap["gadget"])

	cancel()
	consumer.Stop(context.Background())
}

func TestWatermillNATS_MessageRoundtrip(t *testing.T) {
	event := types.StoredEvent{
		ID:       "test-1",
		Name:     "StockReceived",
		Stream:   "warehouse-1",
		Category: "inventory",
		Position: 0,
		Payload:  json.RawMessage(`{"item":"widget","quantity":100}`),
	}

	msg, err := wmadapter.StoredEventToMessage(event)
	require.NoError(t, err)

	assert.Equal(t, "test-1", msg.UUID)
	assert.Equal(t, "StockReceived", msg.Metadata.Get("event_name"))
	assert.Equal(t, "inventory", msg.Metadata.Get("stream_category"))

	restored, err := wmadapter.MessageToStoredEvent(msg)
	require.NoError(t, err)
	assert.Equal(t, event.ID, restored.ID)
	assert.Equal(t, event.Name, restored.Name)
	assert.Equal(t, event.Category, restored.Category)
	assert.Equal(t, fmt.Sprintf("%s", event.Payload), fmt.Sprintf("%s", restored.Payload))
}
