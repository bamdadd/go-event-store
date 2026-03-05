package singleton

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/bamdadd/go-event-store/processing"
)

type SingletonEventBroker struct {
	subscribers []processing.EventSubscriber
	consumers   map[string]processing.EventConsumer
	mu          sync.Mutex
	status      processing.AtomicStatus
}

func New() *SingletonEventBroker {
	return &SingletonEventBroker{
		consumers: make(map[string]processing.EventConsumer),
	}
}

func (b *SingletonEventBroker) Register(sub processing.EventSubscriber) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers = append(b.subscribers, sub)
	return nil
}

func (b *SingletonEventBroker) AddConsumer(id string, consumer processing.EventConsumer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consumers[id] = consumer
}

func (b *SingletonEventBroker) Start(ctx context.Context) error {
	b.status.Set(processing.StatusStarting)

	g, gctx := errgroup.WithContext(ctx)
	for _, sub := range b.subscribers {
		consumer, ok := b.consumers[sub.ID]
		if !ok {
			return fmt.Errorf("no consumer for subscriber %s", sub.ID)
		}
		g.Go(func() error {
			return consumer.Start(gctx)
		})
	}

	b.status.Set(processing.StatusRunning)
	err := g.Wait()
	b.status.Set(processing.StatusStopped)
	return err
}

func (b *SingletonEventBroker) Stop(ctx context.Context) error {
	b.status.Set(processing.StatusStopping)
	for _, consumer := range b.consumers {
		if err := consumer.Drain(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (b *SingletonEventBroker) Status() processing.ProcessStatus {
	return b.status.Get()
}
