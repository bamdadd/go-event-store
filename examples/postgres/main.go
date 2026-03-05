// Example: PostgreSQL event store with synchronous projection
//
// This example demonstrates:
//   - Connecting to PostgreSQL with pgx
//   - Publishing events with optimistic concurrency
//   - Projecting events synchronously from the database
//   - Write condition failures (optimistic locking)
//
// Prerequisites:
//   make db-start
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bamdadd/go-event-store/projection"
	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/store/postgres"
	"github.com/bamdadd/go-event-store/types"
)

type OrderSummary struct {
	Items int
	Total float64
}

func main() {
	ctx := context.Background()

	// --- Connect to PostgreSQL ---
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://event_store:event_store@localhost:5432/event_store?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}
	fmt.Println("Connected to PostgreSQL")

	// --- Set up event store ---
	adapter := postgres.New(pool)
	es := store.NewEventStore(adapter)

	stream := es.Stream("orders", "order-100")

	// --- Publish initial events with StreamIsEmpty condition ---
	stored, err := stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("OrderCreated", map[string]any{"customer": "alice"}),
		types.NewNewEvent("ItemAdded", map[string]any{"item": "widget", "price": 29.99}),
		types.NewNewEvent("ItemAdded", map[string]any{"item": "gadget", "price": 49.99}),
	}, store.StreamIsEmpty())
	if err != nil {
		log.Fatalf("publish: %v", err)
	}
	fmt.Printf("Published %d events\n", len(stored))

	// --- Append more events using PositionIs for optimistic concurrency ---
	lastPos := stored[len(stored)-1].Position
	_, err = stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("ItemRemoved", map[string]any{"item": "widget", "price": 29.99}),
		types.NewNewEvent("OrderPlaced", map[string]any{}),
	}, store.PositionIs(&lastPos))
	if err != nil {
		log.Fatalf("append: %v", err)
	}
	fmt.Println("Appended 2 more events with optimistic lock")

	// --- Demonstrate write condition failure ---
	stalePos := 0
	_, err = stream.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("ItemAdded", map[string]any{"item": "doodad", "price": 9.99}),
	}, store.PositionIs(&stalePos))
	var condErr *store.UnmetWriteConditionError
	if errors.As(err, &condErr) {
		fmt.Printf("Expected conflict: %s\n", condErr.Error())
	}

	// --- Project into order summary ---
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
	if err != nil {
		log.Fatalf("project: %v", err)
	}
	fmt.Printf("\nOrder summary: %d item(s), total $%.2f\n", summary.Items, summary.Total)

	// --- Print all events ---
	fmt.Println("\nEvent history:")
	for event, err := range stream.Scan(ctx) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  pos=%d seq=%d %s %s\n",
			event.Position, event.SequenceNumber, event.Name, event.Payload)
	}
}
