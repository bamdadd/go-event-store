package processing

import "context"

type EventProcessor interface {
	Process(ctx context.Context, event any) error
}

type EventConsumer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Drain(ctx context.Context) error
	Status() ProcessStatus
}
