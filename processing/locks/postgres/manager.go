package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/logicblocks/event-store/processing/locks"
	"github.com/logicblocks/event-store/store"
)

type PostgresLockManager struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *PostgresLockManager {
	return &PostgresLockManager{pool: pool}
}

func (m *PostgresLockManager) TryLock(ctx context.Context, name string) (locks.Lock, locks.LockReleaser, error) {
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return locks.Lock{Name: name}, func() {}, fmt.Errorf("acquire conn: %w", err)
	}

	key := store.LockDigest(name)
	var locked bool
	err = conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&locked)
	if err != nil {
		conn.Release()
		return locks.Lock{Name: name}, func() {}, fmt.Errorf("try advisory lock: %w", err)
	}

	release := func() {
		if locked {
			_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", key)
		}
		conn.Release()
	}

	return locks.Lock{Name: name, Locked: locked}, release, nil
}

func (m *PostgresLockManager) WaitForLock(ctx context.Context, name string) (locks.Lock, locks.LockReleaser, error) {
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return locks.Lock{Name: name}, func() {}, fmt.Errorf("acquire conn: %w", err)
	}

	key := store.LockDigest(name)
	_, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", key)
	if err != nil {
		conn.Release()
		if ctx.Err() != nil {
			return locks.Lock{Name: name, TimedOut: true}, func() {}, ctx.Err()
		}
		return locks.Lock{Name: name}, func() {}, fmt.Errorf("advisory lock: %w", err)
	}

	release := func() {
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", key)
		conn.Release()
	}

	return locks.Lock{Name: name, Locked: true}, release, nil
}
