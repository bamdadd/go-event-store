package distributed

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/bamdadd/go-event-store/processing"
	"github.com/bamdadd/go-event-store/processing/locks"
)

type DistributedEventBroker struct {
	lockManager     locks.LockManager
	coordinatorLock string
	status          processing.AtomicStatus
}

type Option func(*DistributedEventBroker)

func WithCoordinatorLock(name string) Option {
	return func(b *DistributedEventBroker) { b.coordinatorLock = name }
}

func New(lm locks.LockManager, opts ...Option) *DistributedEventBroker {
	b := &DistributedEventBroker{
		lockManager:     lm,
		coordinatorLock: "distributed_broker_coordinator",
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

func (b *DistributedEventBroker) Register(_ processing.EventSubscriber) error {
	return nil
}

func (b *DistributedEventBroker) Start(ctx context.Context) error {
	b.status.Set(processing.StatusStarting)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return b.runCoordinator(gctx)
	})

	g.Go(func() error {
		return b.runObserver(gctx)
	})

	g.Go(func() error {
		return b.runSubscriberManager(gctx)
	})

	b.status.Set(processing.StatusRunning)
	err := g.Wait()
	b.status.Set(processing.StatusStopped)
	return err
}

func (b *DistributedEventBroker) Stop(_ context.Context) error {
	b.status.Set(processing.StatusStopping)
	return nil
}

func (b *DistributedEventBroker) Status() processing.ProcessStatus {
	return b.status.Get()
}

func (b *DistributedEventBroker) runCoordinator(ctx context.Context) error {
	lock, release, err := b.lockManager.WaitForLock(ctx, b.coordinatorLock)
	if err != nil {
		return err
	}
	defer release()

	if !lock.Locked {
		return nil
	}

	<-ctx.Done()
	return ctx.Err()
}

func (b *DistributedEventBroker) runObserver(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (b *DistributedEventBroker) runSubscriberManager(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
