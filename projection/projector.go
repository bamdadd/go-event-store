package projection

import (
	"context"
	"fmt"
	"iter"

	"github.com/logicblocks/event-store/types"
)

type MissingHandlerBehaviour int

const (
	MissingHandlerBehaviourRaise  MissingHandlerBehaviour = iota
	MissingHandlerBehaviourIgnore
)

type EventHandler[S any] func(state S, event types.StoredEvent) (S, error)

type MissingProjectionHandlerError struct {
	EventName string
}

func (e *MissingProjectionHandlerError) Error() string {
	return fmt.Sprintf("no handler registered for event %q", e.EventName)
}

type Projector[S any] struct {
	handlers            map[string]EventHandler[S]
	initialStateFactory func() S
	onMissing           MissingHandlerBehaviour
}

type ProjectorOption[S any] func(*Projector[S])

func WithMissingHandlerBehaviour[S any](b MissingHandlerBehaviour) ProjectorOption[S] {
	return func(p *Projector[S]) { p.onMissing = b }
}

func NewProjector[S any](initialState func() S, opts ...ProjectorOption[S]) *Projector[S] {
	p := &Projector[S]{
		handlers:            make(map[string]EventHandler[S]),
		initialStateFactory: initialState,
		onMissing:           MissingHandlerBehaviourRaise,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func (p *Projector[S]) On(eventName string, handler EventHandler[S]) *Projector[S] {
	p.handlers[eventName] = handler
	return p
}

func (p *Projector[S]) Apply(state S, event types.StoredEvent) (S, error) {
	handler, ok := p.handlers[event.Name]
	if !ok {
		switch p.onMissing {
		case MissingHandlerBehaviourRaise:
			return state, &MissingProjectionHandlerError{EventName: event.Name}
		default:
			return state, nil
		}
	}
	return handler(state, event)
}

func (p *Projector[S]) Project(
	_ context.Context,
	scan iter.Seq2[types.StoredEvent, error],
) (S, error) {
	state := p.initialStateFactory()
	for event, err := range scan {
		if err != nil {
			return state, err
		}
		state, err = p.Apply(state, event)
		if err != nil {
			return state, err
		}
	}
	return state, nil
}
