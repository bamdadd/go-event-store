# CLAUDE.md

## Project Overview

`github.com/bamdadd/go-event-store` is a Go library providing eventing
infrastructure for event-sourced architectures. It is a migration of the
Python `event.store` library.

## Architecture

Key packages under the module root:

- **`types/`**: Core value types — `NewEvent`, `StoredEvent`, `StreamIdentifier`, `CategoryIdentifier`, `LogIdentifier`
- **`store/`**: `EventStore` facade with stream/category/log sources, write conditions, and adapter-based storage
  - **`store/memory/`**: In-memory storage adapter (for tests)
  - **`store/postgres/`**: PostgreSQL storage adapter (for production)
- **`projection/`**: `Projector[S]` for reducing event streams into state, plus `ProjectionStore` with adapters
- **`processing/`**: Event processing pipeline — broker, consumers, locks, Watermill integration
  - **`processing/watermill/`**: Watermill pub/sub adapters
  - **`processing/broker/singleton/`**: Single-node broker
  - **`processing/broker/distributed/`**: Multi-node broker with leader election
  - **`processing/locks/`**: Lock manager interface + in-memory and PostgreSQL implementations
- **`query/`**: Query DSL for projection store search
- **`internal/clock/`**: Clock abstraction for testability
- **`internal/testutil/`**: Test builders, generators, and shared adapter test harness

### Key Patterns

- **Adapter pattern**: Both EventStore and ProjectionStore use interfaces with in-memory (tests) and PostgreSQL (production) implementations
- **Shared test harness**: `internal/testutil/harness.go` contains reusable test functions that both adapter implementations run
- **Table-driven tests**: Go idiomatic testing with testify
- **Watermill for processing**: Event routing, retry middleware, poison queue, multi-backend pub/sub
- **Go generics**: `Projector[S]` for type-safe projections
- **Implicit interfaces**: Small, focused interfaces following ISP

## Development Commands

### Setup
```bash
mise install          # Install Go 1.23, golangci-lint
mise trust            # Trust mise config (first time only)
```

### Build & Test
```bash
make build            # Build all packages
make test-unit        # Run unit tests with race detector
make test-integration # Run integration tests (requires PostgreSQL)
make test-all         # Run all tests
make lint             # Run golangci-lint
```

### Database (for integration tests)
```bash
make db-start         # Start PostgreSQL via docker-compose
make db-stop          # Stop PostgreSQL
```

## Development Approach

- Follow Test-Driven Development (TDD) strictly: red-green-refactor
- Follow SOLID principles, especially ISP for interfaces
- Use Go 1.23 features: `iter.Seq2` for scanning, generics for type safety
- Prefer `context.Context` for cancellation, `errgroup` for goroutine coordination
- Use `(result, error)` returns — no panics for expected errors
- Use functional options pattern for configuration
- Target Go 1.23 or greater

## Testing

- Unit tests: `make test-unit` or `go test -race ./...`
- Integration tests: `make test-integration` or `go test -race -tags=integration ./...`
- Use `//go:build integration` tag for tests requiring PostgreSQL
- Use testify for assertions (`assert`, `require`)
- Single assertion per test where practical
- Shared adapter test harness ensures consistency across storage backends
