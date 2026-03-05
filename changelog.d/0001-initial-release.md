# Initial Release

## Added

- Core event types: `NewEvent`, `StoredEvent`, stream/category/log identifiers
- `EventStore` with stream, category, and log access
- Write conditions for optimistic concurrency: `NoCondition`, `PositionIs`, `StreamIsEmpty`, `And`, `Or`
- Iterator-based scanning with `iter.Seq2`
- In-memory storage adapter for testing
- PostgreSQL storage adapter with pgx v5
- Type-safe projections with `Projector[S]` using Go generics
- Projection store with in-memory adapter and query DSL
- Event processing pipeline with polling service and retry policies
- Singleton and distributed broker implementations
- Distributed coordination with leader election
- Lock management with in-memory and PostgreSQL backends
- Watermill pub/sub integration
- Shared adapter test harness for consistent cross-backend behaviour
