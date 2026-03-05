package testutil

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bamdadd/go-event-store/internal/clock"
	"github.com/bamdadd/go-event-store/types"
)

var testClock = clock.StaticClock{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}

func NewTestEvent(name string, payload map[string]any) types.NewEvent {
	return types.NewNewEvent(name, payload, types.WithClock(testClock))
}

func NewTestEvents(count int) []types.NewEvent {
	events := make([]types.NewEvent, count)
	for i := range count {
		events[i] = NewTestEvent(
			fmt.Sprintf("event-%d", i+1),
			map[string]any{"index": i + 1},
		)
	}
	return events
}

func MustMarshalJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
