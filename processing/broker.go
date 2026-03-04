package processing

import "context"

type EventSubscriber struct {
	ID      string
	GroupID string
	NodeID  string
	Sources []string
}

type EventBroker interface {
	Register(subscriber EventSubscriber) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Status() ProcessStatus
}
