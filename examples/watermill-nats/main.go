// Example: Async event processing with Watermill and NATS JetStream
//
// This example demonstrates:
//   - Publishing events to an in-memory event store
//   - Forwarding stored events to NATS JetStream via Watermill
//   - Consuming events asynchronously with an EventProcessor
//   - Building projections from async event streams
//
// Prerequisites:
//   docker run -d --name nats -p 4222:4222 nats:latest -js
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"

	wmadapter "github.com/bamdadd/go-event-store/processing/watermill"
	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/store/memory"
	"github.com/bamdadd/go-event-store/types"
)

const topic = "events.inventory"

// InventoryState tracks stock levels, updated asynchronously.
type InventoryState struct {
	mu    sync.Mutex
	stock map[string]int
}

func (s *InventoryState) Apply(item string, qty int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stock[item] += qty
}

func (s *InventoryState) Snapshot() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make(map[string]int, len(s.stock))
	for k, v := range s.stock {
		cp[k] = v
	}
	return cp
}

// inventoryProcessor implements processing.EventProcessor.
type inventoryProcessor struct {
	state *InventoryState
}

func (p *inventoryProcessor) Process(_ context.Context, event any) error {
	e, ok := event.(types.StoredEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	var payload map[string]any
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return err
	}

	item, _ := payload["item"].(string)
	qty := int(payload["quantity"].(float64))

	switch e.Name {
	case "StockReceived":
		p.state.Apply(item, qty)
		fmt.Printf("  [async] received %d x %s\n", qty, item)
	case "StockShipped":
		p.state.Apply(item, -qty)
		fmt.Printf("  [async] shipped %d x %s\n", qty, item)
	}
	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	wmLogger := watermill.NewStdLogger(false, false)

	// --- NATS JetStream publisher ---
	publisher, err := nats.NewPublisher(
		nats.PublisherConfig{
			URL: natsURL(),
			JetStream: nats.JetStreamConfig{
				AutoProvision: true,
			},
		},
		wmLogger,
	)
	if err != nil {
		log.Fatalf("nats publisher: %v", err)
	}
	defer publisher.Close()

	// --- NATS JetStream subscriber ---
	subscriber, err := nats.NewSubscriber(
		nats.SubscriberConfig{
			URL: natsURL(),
			JetStream: nats.JetStreamConfig{
				AutoProvision: true,
			},
		},
		wmLogger,
	)
	if err != nil {
		log.Fatalf("nats subscriber: %v", err)
	}
	defer subscriber.Close()

	// --- Set up event store and async processor ---
	eventAdapter := memory.New()
	es := store.NewEventStore(eventAdapter)

	inventory := &InventoryState{stock: make(map[string]int)}
	processor := &inventoryProcessor{state: inventory}

	consumer := wmadapter.NewConsumer(subscriber, topic, processor)

	// Start consuming in background
	go func() {
		if err := consumer.Start(ctx); err != nil && ctx.Err() == nil {
			log.Printf("consumer stopped: %v", err)
		}
	}()

	// Give consumer time to subscribe
	time.Sleep(500 * time.Millisecond)

	// --- Publish events ---
	stream := es.Stream("inventory", "warehouse-1")

	events := []types.NewEvent{
		types.NewNewEvent("StockReceived", map[string]any{"item": "widget", "quantity": 100}),
		types.NewNewEvent("StockReceived", map[string]any{"item": "gadget", "quantity": 50}),
		types.NewNewEvent("StockShipped", map[string]any{"item": "widget", "quantity": 30}),
		types.NewNewEvent("StockReceived", map[string]any{"item": "widget", "quantity": 20}),
		types.NewNewEvent("StockShipped", map[string]any{"item": "gadget", "quantity": 10}),
	}

	stored, err := stream.Publish(ctx, events, store.StreamIsEmpty())
	if err != nil {
		log.Fatalf("publish: %v", err)
	}
	fmt.Printf("Published %d events to event store\n", len(stored))

	// Forward stored events to NATS via Watermill
	fmt.Println("\nForwarding events to NATS JetStream...")
	for _, e := range stored {
		msg, err := wmadapter.StoredEventToMessage(e)
		if err != nil {
			log.Fatalf("convert: %v", err)
		}
		if err := publisher.Publish(topic, msg); err != nil {
			log.Fatalf("publish to nats: %v", err)
		}
	}

	// Wait for async processing
	fmt.Println("\nWaiting for async processing...")
	time.Sleep(2 * time.Second)

	// --- Show final state ---
	fmt.Println("\nInventory levels (built asynchronously):")
	for item, qty := range inventory.Snapshot() {
		fmt.Printf("  %s: %d\n", item, qty)
	}

	// Also show the event store has all events
	fmt.Println("\nEvent store contents:")
	for event, err := range stream.Scan(ctx) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  seq=%d %s %s\n", event.SequenceNumber, event.Name, event.Payload)
	}

	cancel()
	consumer.Stop(context.Background())
}

func natsURL() string {
	if url := os.Getenv("NATS_URL"); url != "" {
		return url
	}
	return "nats://localhost:4222"
}
