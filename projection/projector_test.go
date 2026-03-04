package projection_test

import (
	"context"
	"encoding/json"
	"iter"
	"testing"

	"github.com/logicblocks/event-store/projection"
	"github.com/logicblocks/event-store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type orderState struct {
	Total int
	Items []string
}

func initialOrderState() orderState { return orderState{} }

func eventsIter(events ...types.StoredEvent) iter.Seq2[types.StoredEvent, error] {
	return func(yield func(types.StoredEvent, error) bool) {
		for _, e := range events {
			if !yield(e, nil) {
				return
			}
		}
	}
}

func TestProjectorApplyCallsRegisteredHandler(t *testing.T) {
	p := projection.NewProjector(initialOrderState).
		On("item-added", func(s orderState, e types.StoredEvent) (orderState, error) {
			s.Items = append(s.Items, e.Name)
			s.Total++
			return s, nil
		})

	state, err := p.Apply(orderState{}, types.StoredEvent{Name: "item-added"})
	require.NoError(t, err)
	assert.Equal(t, 1, state.Total)
}

func TestProjectorApplyRaisesOnMissingHandler(t *testing.T) {
	p := projection.NewProjector(initialOrderState)

	_, err := p.Apply(orderState{}, types.StoredEvent{Name: "unknown-event"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown-event")
}

func TestProjectorApplyIgnoresMissingHandlerWhenConfigured(t *testing.T) {
	p := projection.NewProjector(
		initialOrderState,
		projection.WithMissingHandlerBehaviour[orderState](projection.MissingHandlerBehaviourIgnore),
	)

	state, err := p.Apply(orderState{Total: 5}, types.StoredEvent{Name: "unknown-event"})
	require.NoError(t, err)
	assert.Equal(t, 5, state.Total)
}

func TestProjectorProjectFoldsMultipleEvents(t *testing.T) {
	p := projection.NewProjector(initialOrderState).
		On("item-added", func(s orderState, _ types.StoredEvent) (orderState, error) {
			s.Total++
			return s, nil
		})

	events := eventsIter(
		types.StoredEvent{Name: "item-added", Payload: json.RawMessage(`{}`)},
		types.StoredEvent{Name: "item-added", Payload: json.RawMessage(`{}`)},
		types.StoredEvent{Name: "item-added", Payload: json.RawMessage(`{}`)},
	)

	state, err := p.Project(context.Background(), events)
	require.NoError(t, err)
	assert.Equal(t, 3, state.Total)
}

func TestProjectorProjectStartsFromInitialState(t *testing.T) {
	p := projection.NewProjector(func() orderState {
		return orderState{Total: 10}
	})

	state, err := p.Project(context.Background(), eventsIter())
	require.NoError(t, err)
	assert.Equal(t, 10, state.Total)
}

func TestProjectorOnChaining(t *testing.T) {
	p := projection.NewProjector(initialOrderState).
		On("a", func(s orderState, _ types.StoredEvent) (orderState, error) { s.Total += 1; return s, nil }).
		On("b", func(s orderState, _ types.StoredEvent) (orderState, error) { s.Total += 10; return s, nil })

	events := eventsIter(
		types.StoredEvent{Name: "a"},
		types.StoredEvent{Name: "b"},
		types.StoredEvent{Name: "a"},
	)
	state, err := p.Project(context.Background(), events)
	require.NoError(t, err)
	assert.Equal(t, 12, state.Total)
}
