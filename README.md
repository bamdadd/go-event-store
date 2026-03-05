# event-store-go

A Go library providing eventing infrastructure for event-sourced architectures.

> Also available in Python: [event.store](https://github.com/logicblocks/event.store)

## Installation

```bash
go get github.com/bamdadd/go-event-store
```

Requires Go 1.24+.

## Quick Start

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/bamdadd/go-event-store/projection"
    "github.com/bamdadd/go-event-store/store"
    "github.com/bamdadd/go-event-store/store/memory"
    "github.com/bamdadd/go-event-store/types"
)

type AccountBalance struct {
    Balance int
}

func main() {
    ctx := context.Background()

    // Create an event store with the in-memory adapter
    adapter := memory.NewAdapter()
    es := store.NewEventStore(adapter)

    // Publish events to a stream
    stream := es.Stream("accounts", "acc-123")
    stream.Publish(ctx, []types.NewEvent{
        types.NewNewEvent("MoneyDeposited", map[string]any{"amount": 100}),
        types.NewNewEvent("MoneyWithdrawn", map[string]any{"amount": 30}),
    }, store.NoCondition())

    // Project events into state
    projector := projection.NewProjector(func() AccountBalance {
        return AccountBalance{}
    })
    projector.
        On("MoneyDeposited", func(s AccountBalance, e types.StoredEvent) (AccountBalance, error) {
            var payload map[string]json.Number
            json.Unmarshal(e.Payload, &payload)
            amount, _ := payload["amount"].Int64()
            s.Balance += int(amount)
            return s, nil
        }).
        On("MoneyWithdrawn", func(s AccountBalance, e types.StoredEvent) (AccountBalance, error) {
            var payload map[string]json.Number
            json.Unmarshal(e.Payload, &payload)
            amount, _ := payload["amount"].Int64()
            s.Balance -= int(amount)
            return s, nil
        })

    balance, _ := projector.Project(ctx, stream.Scan(ctx))
    fmt.Printf("Balance: %d\n", balance.Balance) // Balance: 70
}
```

## Features

### Event Store

- **Stream / Category / Log** hierarchy for organising events
- **Append-only** immutable event log
- **Optimistic concurrency** via composable write conditions (`PositionIs`, `StreamIsEmpty`, `And`, `Or`)
- **Bi-temporal** event timestamps (observed and occurred)
- **Iterator-based scanning** using `iter.Seq2` for memory-efficient reads

### Storage Adapters

| Adapter | Package | Use Case |
|---------|---------|----------|
| In-memory | `store/memory` | Testing and prototyping |
| PostgreSQL | `store/postgres` | Production workloads |

Both adapters pass the same shared test harness ensuring consistent behaviour.

### Projections

- **Type-safe projectors** using Go generics (`Projector[S]`)
- **Projection store** with pluggable storage adapters (in-memory, PostgreSQL)
- **Query DSL** for searching projections (`query` package)

### Event Processing

- **Polling service** with configurable retry policies
- **Broker** abstraction with singleton and distributed implementations
- **Distributed coordination** with leader election for multi-node deployments
- **Lock management** (in-memory and PostgreSQL backends)
- **Watermill integration** for pub/sub message routing

## Architecture

```
store/              EventStore, streams, categories, log, write conditions
  memory/           In-memory adapter
  postgres/         PostgreSQL adapter (pgx v5)
types/              NewEvent, StoredEvent, identifiers
projection/         Projector[S], ProjectionStore
  memory/           In-memory projection adapter
processing/         Polling service, retry policies, broker/consumer interfaces
  broker/
    singleton/      Single-node broker
    distributed/    Multi-node broker with leader election
  locks/
    memory/         In-memory lock manager
    postgres/       PostgreSQL lock manager
  watermill/        Watermill pub/sub adapters
query/              Query DSL for projection store
```

## Write Conditions

Optimistic concurrency control for safe concurrent writes:

```go
// Only write if the stream is at a known position
stream.Publish(ctx, events, store.PositionIs(&pos))

// Only write to a new stream
stream.Publish(ctx, events, store.StreamIsEmpty())

// Compose conditions
stream.Publish(ctx, events, store.Or(store.StreamIsEmpty(), store.PositionIs(&pos)))
```

## Development

### Prerequisites

```bash
mise install   # Installs Go, golangci-lint
```

### Commands

```bash
make build              # Build all packages
make test-unit          # Run unit tests
make test-integration   # Run integration tests (requires PostgreSQL)
make test-all           # Run all tests
make lint               # Run linter
```

### Database (integration tests)

```bash
make db-start   # Start PostgreSQL via docker-compose
make db-stop    # Stop PostgreSQL
```

## License

MIT
