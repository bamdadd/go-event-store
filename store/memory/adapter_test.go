package memory_test

import (
	"context"
	"testing"

	"github.com/bamdadd/go-event-store/internal/testutil"
	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/store/memory"
	"github.com/bamdadd/go-event-store/types"
)

func newHarness() testutil.AdapterTestHarness {
	var adapter *memory.InMemoryEventStorageAdapter
	return testutil.AdapterTestHarness{
		ConstructAdapter: func(t *testing.T, g store.SerialisationGuarantee) store.EventStorageAdapter {
			adapter = memory.New(memory.WithGuarantee(g))
			return adapter
		},
		ClearStorage: func(t *testing.T) {
			if adapter != nil {
				adapter.ClearAll()
			}
		},
		RetrieveEvents: func(t *testing.T, a store.EventStorageAdapter, category, stream string) []types.StoredEvent {
			var events []types.StoredEvent
			target := types.StreamIdentifier{Category: category, Stream: stream}
			for e, err := range a.Scan(context.Background(), target) {
				if err != nil {
					t.Fatal(err)
				}
				events = append(events, e)
			}
			return events
		},
	}
}

func TestInMemoryAdapter_StreamSave(t *testing.T) {
	testutil.RunStreamSaveTests(t, newHarness())
}

func TestInMemoryAdapter_Scan(t *testing.T) {
	testutil.RunScanTests(t, newHarness())
}

func TestInMemoryAdapter_Latest(t *testing.T) {
	testutil.RunLatestTests(t, newHarness())
}
