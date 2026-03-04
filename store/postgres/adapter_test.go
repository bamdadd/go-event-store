//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/logicblocks/event-store/internal/testutil"
	"github.com/logicblocks/event-store/store"
	"github.com/logicblocks/event-store/store/postgres"
	"github.com/logicblocks/event-store/types"
)

func newPostgresHarness(t *testing.T) testutil.AdapterTestHarness {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/event_store_test?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	t.Cleanup(pool.Close)

	return testutil.AdapterTestHarness{
		ConstructAdapter: func(t *testing.T, g store.SerialisationGuarantee) store.EventStorageAdapter {
			return postgres.New(pool, postgres.WithGuarantee(g))
		},
		ClearStorage: func(t *testing.T) {
			_, err := pool.Exec(context.Background(), "TRUNCATE events RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("truncate: %v", err)
			}
		},
		RetrieveEvents: func(t *testing.T, a store.EventStorageAdapter, category, stream string) []types.StoredEvent {
			var events []types.StoredEvent
			target := types.StreamIdentifier{Category: category, Stream: stream}
			for e, scanErr := range a.Scan(context.Background(), target) {
				if scanErr != nil {
					t.Fatal(scanErr)
				}
				events = append(events, e)
			}
			return events
		},
	}
}

func TestPostgresAdapter_StreamSave(t *testing.T) {
	testutil.RunStreamSaveTests(t, newPostgresHarness(t))
}

func TestPostgresAdapter_Scan(t *testing.T) {
	testutil.RunScanTests(t, newPostgresHarness(t))
}

func TestPostgresAdapter_Latest(t *testing.T) {
	testutil.RunLatestTests(t, newPostgresHarness(t))
}
