# Architecture

## High-Level Overview

```mermaid
graph TB
    subgraph Application
        APP[Application Code]
    end

    subgraph Event Store
        ES[EventStore]
        STR[Stream]
        CAT[Category]
        LOG[Log]
    end

    subgraph Storage
        ESA[EventStorageAdapter]
        MEM[Memory Adapter]
        PG[PostgreSQL Adapter]
    end

    subgraph Projections
        PRJ[Projector S]
        PS[ProjectionStore]
        PSA[ProjectionStorageAdapter]
        PMEM[Memory Adapter]
    end

    subgraph Processing
        BRK[EventBroker]
        SB[Singleton Broker]
        DB[Distributed Broker]
        CON[EventConsumer]
        SVC[PollingService]
        LM[LockManager]
        WM[Watermill Pub/Sub]
    end

    APP --> ES
    APP --> PRJ
    APP --> BRK

    ES --> STR
    ES --> CAT
    ES --> LOG

    STR --> ESA
    CAT --> ESA
    LOG --> ESA

    ESA -.-> MEM
    ESA -.-> PG

    PRJ -->|fold events| STR
    PRJ --> PS
    PS --> PSA
    PSA -.-> PMEM

    BRK -.-> SB
    BRK -.-> DB
    DB --> LM
    BRK --> CON
    CON --> SVC
    CON --> WM

    style ES fill:#4a90d9,color:#fff
    style PRJ fill:#50b87a,color:#fff
    style BRK fill:#e8854a,color:#fff
    style PG fill:#336791,color:#fff
```

## Event Lifecycle

```mermaid
sequenceDiagram
    participant App as Application
    participant ES as EventStore
    participant Stream as Stream
    participant Adapter as StorageAdapter
    participant DB as PostgreSQL

    App->>ES: .Stream("orders", "order-123")
    ES-->>Stream: EventStream

    App->>Stream: Publish(events, condition)
    Stream->>Adapter: SaveToStream(id, events, condition)
    Adapter->>DB: INSERT with optimistic lock
    DB-->>Adapter: StoredEvents
    Adapter-->>Stream: []StoredEvent
    Stream-->>App: []StoredEvent

    App->>Stream: Scan(ctx)
    Stream->>Adapter: Scan(id, opts)
    Adapter-->>App: iter.Seq2[StoredEvent, error]
```

## Event Organisation

```mermaid
graph LR
    subgraph Log [Event Log - all events]
        subgraph Cat1 [Category: orders]
            S1[Stream: order-1]
            S2[Stream: order-2]
        end
        subgraph Cat2 [Category: payments]
            S3[Stream: pay-1]
        end
    end

    S1 -->|pos 0| E1[OrderCreated]
    S1 -->|pos 1| E2[ItemAdded]
    S1 -->|pos 2| E3[OrderPlaced]
    S2 -->|pos 0| E4[OrderCreated]
    S3 -->|pos 0| E5[PaymentReceived]

    style Log fill:#f5f5f5,stroke:#999
    style Cat1 fill:#e3f2fd,stroke:#4a90d9
    style Cat2 fill:#e8f5e9,stroke:#50b87a
```

## Projection Flow

```mermaid
graph LR
    S[Stream Scan] -->|iter.Seq2| P[Projector S]
    P -->|fold| STATE[Computed State]
    STATE --> PS[ProjectionStore]
    PS --> PSA[StorageAdapter]

    P -->|On MoneyDeposited| H1[Handler 1]
    P -->|On MoneyWithdrawn| H2[Handler 2]

    style P fill:#50b87a,color:#fff
    style STATE fill:#f9e547,color:#333
```

## Write Conditions

Optimistic concurrency is enforced via composable write conditions evaluated before each write.

```mermaid
graph TD
    WC[WriteCondition]
    WC --> NC[NoCondition]
    WC --> PI[PositionIs pos]
    WC --> SE[StreamIsEmpty]
    WC --> AND[And c1 c2 ...]
    WC --> OR[Or c1 c2 ...]

    AND --> WC
    OR --> WC

    style WC fill:#4a90d9,color:#fff
```

## Processing Pipeline

```mermaid
graph TB
    subgraph Broker
        SB[SingletonBroker]
        DB[DistributedBroker]
        COORD[Coordinator]
        LE[Leader Election]
    end

    subgraph Consumer
        EC[EventConsumer]
        EP[EventProcessor]
        PS2[PollingService]
        RP[RetryPolicy]
    end

    subgraph Infrastructure
        LM[LockManager]
        LMEM[Memory Lock]
        LPG[Postgres Lock]
        PUB[Watermill Publisher]
        SUB[Watermill Subscriber]
    end

    DB --> COORD
    COORD --> LE
    LE --> LM
    LM -.-> LMEM
    LM -.-> LPG

    SB --> EC
    DB --> EC
    EC --> EP
    EC --> PS2
    PS2 --> RP

    EC --> SUB
    EP --> PUB

    style DB fill:#e8854a,color:#fff
    style SB fill:#e8854a,color:#fff
    style LM fill:#9b59b6,color:#fff
```

## Package Structure

```mermaid
graph TD
    ROOT[github.com/bamdadd/go-event-store]

    ROOT --> TYPES[types]
    ROOT --> STORE[store]
    ROOT --> PROJ[projection]
    ROOT --> PROC[processing]
    ROOT --> QUERY[query]

    STORE --> SMEM[store/memory]
    STORE --> SPG[store/postgres]

    PROJ --> PMEM[projection/memory]

    PROC --> LOCKS[processing/locks]
    PROC --> BRKR[processing/broker]
    PROC --> WMILL[processing/watermill]

    LOCKS --> LMEM[locks/memory]
    LOCKS --> LPG[locks/postgres]

    BRKR --> SING[broker/singleton]
    BRKR --> DIST[broker/distributed]

    TYPES -.->|imported by| STORE
    TYPES -.->|imported by| PROJ
    STORE -.->|imported by| PROC

    style ROOT fill:#333,color:#fff
    style TYPES fill:#4a90d9,color:#fff
    style STORE fill:#4a90d9,color:#fff
    style PROJ fill:#50b87a,color:#fff
    style PROC fill:#e8854a,color:#fff
    style QUERY fill:#9b59b6,color:#fff
```
