package testutil

import (
	"testing"

	"github.com/logicblocks/event-store/store"
	"github.com/logicblocks/event-store/types"
)

type AdapterTestHarness struct {
	ConstructAdapter func(t *testing.T, guarantee store.SerialisationGuarantee) store.EventStorageAdapter
	ClearStorage     func(t *testing.T)
	RetrieveEvents   func(t *testing.T, adapter store.EventStorageAdapter, category, stream string) []types.StoredEvent
}
