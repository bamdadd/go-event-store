// Example: In-memory event store with synchronous projection
//
// This example demonstrates:
//   - Publishing events to a stream using the in-memory adapter
//   - Synchronously projecting events into computed state
//   - Storing and retrieving projections
//   - Scanning events from a category
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/bamdadd/go-event-store/projection"
	projmem "github.com/bamdadd/go-event-store/projection/memory"
	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/store/memory"
	"github.com/bamdadd/go-event-store/types"
)

type AccountBalance struct {
	Owner   string
	Balance float64
}

func main() {
	ctx := context.Background()

	// --- Set up stores ---
	eventAdapter := memory.New()
	es := store.NewEventStore(eventAdapter)

	projAdapter := projmem.New()
	projStore := projection.NewProjectionStore(projAdapter)

	// --- Publish events to multiple streams ---
	alice := es.Stream("accounts", "alice")
	bob := es.Stream("accounts", "bob")

	_, err := alice.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("AccountOpened", map[string]any{"owner": "Alice"}),
		types.NewNewEvent("MoneyDeposited", map[string]any{"amount": 500.0}),
		types.NewNewEvent("MoneyWithdrawn", map[string]any{"amount": 50.0}),
	}, store.StreamIsEmpty())
	if err != nil {
		log.Fatal(err)
	}

	_, err = bob.Publish(ctx, []types.NewEvent{
		types.NewNewEvent("AccountOpened", map[string]any{"owner": "Bob"}),
		types.NewNewEvent("MoneyDeposited", map[string]any{"amount": 1000.0}),
		types.NewNewEvent("MoneyDeposited", map[string]any{"amount": 250.0}),
	}, store.StreamIsEmpty())
	if err != nil {
		log.Fatal(err)
	}

	// --- Build a projector ---
	projector := projection.NewProjector(func() AccountBalance {
		return AccountBalance{}
	})
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

	// --- Synchronous projection: project each stream and store ---
	for _, stream := range []*store.EventStream{alice, bob} {
		state, err := projector.Project(ctx, stream.Scan(ctx))
		if err != nil {
			log.Fatal(err)
		}

		stateJSON, _ := json.Marshal(state)
		err = projStore.Save(ctx, projection.Projection{
			ID:    stream.Identifier().Stream,
			Name:  "account-balance",
			State: stateJSON,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%-6s balance: $%.2f\n", state.Owner, state.Balance)
	}

	// --- Retrieve a stored projection ---
	proj, err := projStore.FindOne(ctx, "account-balance", "alice")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nStored projection for alice: %s\n", proj.State)

	// --- Scan all events across the category ---
	fmt.Println("\nAll events in 'accounts' category:")
	for event, err := range es.Category("accounts").Scan(ctx) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  [%s/%s pos=%d] %s %s\n",
			event.Category, event.Stream, event.Position, event.Name, event.Payload)
	}

	// --- Scan the full event log ---
	fmt.Println("\nFull event log:")
	for event, err := range es.Log().Scan(ctx) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  seq=%-2d [%s/%s] %s\n",
			event.SequenceNumber, event.Category, event.Stream, event.Name)
	}
}
