package memory

import (
	"sync"

	"github.com/bamdadd/go-event-store/types"
)

type eventsDB struct {
	mu            sync.RWMutex
	events        map[int]types.StoredEvent
	nextSeqNum    int
	streamIndex   map[string]map[string][]int
	categoryIndex map[string][]int
	logIndex      []int
}

func newEventsDB() *eventsDB {
	return &eventsDB{
		events:        make(map[int]types.StoredEvent),
		nextSeqNum:    1,
		streamIndex:   make(map[string]map[string][]int),
		categoryIndex: make(map[string][]int),
	}
}

func (db *eventsDB) clear() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.events = make(map[int]types.StoredEvent)
	db.nextSeqNum = 1
	db.streamIndex = make(map[string]map[string][]int)
	db.categoryIndex = make(map[string][]int)
	db.logIndex = nil
}
