package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/logicblocks/event-store/store"
	"github.com/logicblocks/event-store/types"
)

type Option func(*PostgresEventStorageAdapter)

func WithGuarantee(g store.SerialisationGuarantee) Option {
	return func(a *PostgresEventStorageAdapter) { a.guarantee = g }
}

func WithNamespace(ns string) Option {
	return func(a *PostgresEventStorageAdapter) { a.namespace = ns }
}

type PostgresEventStorageAdapter struct {
	pool      *pgxpool.Pool
	guarantee store.SerialisationGuarantee
	namespace string
}

func New(pool *pgxpool.Pool, opts ...Option) *PostgresEventStorageAdapter {
	a := &PostgresEventStorageAdapter{
		pool:      pool,
		guarantee: store.GuaranteeLog,
		namespace: "event_store",
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

func (a *PostgresEventStorageAdapter) SaveToStream(
	ctx context.Context,
	target types.StreamIdentifier,
	events []types.NewEvent,
	condition store.WriteCondition,
) ([]types.StoredEvent, error) {
	if len(events) == 0 {
		return nil, nil
	}

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	lockName := a.guarantee.LockName(a.namespace, target)
	lockKey := store.LockDigest(lockName)
	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockKey)
	if err != nil {
		return nil, fmt.Errorf("advisory lock: %w", err)
	}

	var currentPos *int
	err = tx.QueryRow(ctx,
		"SELECT MAX(position) FROM events WHERE category = $1 AND stream = $2",
		target.Category, target.Stream,
	).Scan(&currentPos)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("query position: %w", err)
	}

	ok, evalErr := store.EvaluateCondition(condition, currentPos)
	if evalErr != nil {
		return nil, evalErr
	}
	if !ok {
		return nil, &store.UnmetWriteConditionError{
			Message: fmt.Sprintf("write condition not met for stream %s/%s", target.Category, target.Stream),
		}
	}

	nextPos := 0
	if currentPos != nil {
		nextPos = *currentPos + 1
	}

	stored := make([]types.StoredEvent, 0, len(events))
	for i, e := range events {
		payload, marshalErr := json.Marshal(e.Payload)
		if marshalErr != nil {
			return nil, marshalErr
		}
		id := fmt.Sprintf("%s-%s-%d", target.Category, target.Stream, nextPos+i)

		var se types.StoredEvent
		err = tx.QueryRow(ctx,
			`INSERT INTO events (id, name, stream, category, position, payload, observed_at, occurred_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 RETURNING id, name, stream, category, position, sequence_number, payload, observed_at, occurred_at`,
			id, e.Name, target.Stream, target.Category, nextPos+i, payload, e.ObservedAt, e.OccurredAt,
		).Scan(&se.ID, &se.Name, &se.Stream, &se.Category, &se.Position,
			&se.SequenceNumber, &se.Payload, &se.ObservedAt, &se.OccurredAt)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return nil, &store.UnmetWriteConditionError{Message: pgErr.Detail}
			}
			return nil, fmt.Errorf("insert event: %w", err)
		}
		stored = append(stored, se)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return stored, nil
}

func (a *PostgresEventStorageAdapter) SaveToCategory(
	ctx context.Context,
	target types.CategoryIdentifier,
	streams map[string]store.StreamPublishDefinition,
) (map[string][]types.StoredEvent, error) {
	result := make(map[string][]types.StoredEvent)
	for streamName, def := range streams {
		streamTarget := types.StreamIdentifier{Category: target.Category, Stream: streamName}
		stored, err := a.SaveToStream(ctx, streamTarget, def.Events, def.Condition)
		if err != nil {
			return nil, err
		}
		result[streamName] = stored
	}
	return result, nil
}

func (a *PostgresEventStorageAdapter) Latest(ctx context.Context, target types.Targetable) (*types.StoredEvent, error) {
	query, args := buildLatestQuery(target)
	row := a.pool.QueryRow(ctx, query, args...)

	var e types.StoredEvent
	err := row.Scan(&e.ID, &e.Name, &e.Stream, &e.Category, &e.Position,
		&e.SequenceNumber, &e.Payload, &e.ObservedAt, &e.OccurredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("latest query: %w", err)
	}
	return &e, nil
}

func (a *PostgresEventStorageAdapter) Scan(
	ctx context.Context,
	target types.Targetable,
	opts ...store.ScanOption,
) iter.Seq2[types.StoredEvent, error] {
	cfg := store.BuildScanConfig(opts...)

	return func(yield func(types.StoredEvent, error) bool) {
		query, args := buildScanQuery(target, cfg)
		rows, err := a.pool.Query(ctx, query, args...)
		if err != nil {
			yield(types.StoredEvent{}, fmt.Errorf("scan query: %w", err))
			return
		}
		defer rows.Close()

		for rows.Next() {
			var e types.StoredEvent
			if scanErr := rows.Scan(&e.ID, &e.Name, &e.Stream, &e.Category, &e.Position,
				&e.SequenceNumber, &e.Payload, &e.ObservedAt, &e.OccurredAt); scanErr != nil {
				yield(types.StoredEvent{}, fmt.Errorf("scan row: %w", scanErr))
				return
			}
			if !yield(e, nil) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			yield(types.StoredEvent{}, err)
		}
	}
}
