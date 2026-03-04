package singleton_test

import (
	"context"
	"testing"

	"github.com/logicblocks/event-store/processing"
	"github.com/logicblocks/event-store/processing/broker/singleton"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConsumer struct {
	started bool
	drained bool
	status  processing.AtomicStatus
}

func (m *mockConsumer) Start(ctx context.Context) error {
	m.started = true
	m.status.Set(processing.StatusRunning)
	<-ctx.Done()
	m.status.Set(processing.StatusStopped)
	return ctx.Err()
}

func (m *mockConsumer) Stop(_ context.Context) error {
	m.status.Set(processing.StatusStopping)
	return nil
}

func (m *mockConsumer) Drain(_ context.Context) error {
	m.drained = true
	return nil
}

func (m *mockConsumer) Status() processing.ProcessStatus {
	return m.status.Get()
}

func TestRegisterSubscriber(t *testing.T) {
	broker := singleton.New()
	err := broker.Register(processing.EventSubscriber{ID: "sub-1"})
	require.NoError(t, err)
}

func TestStartWithNoConsumerReturnsError(t *testing.T) {
	broker := singleton.New()
	_ = broker.Register(processing.EventSubscriber{ID: "sub-1"})
	err := broker.Start(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no consumer for subscriber sub-1")
}

func TestStartAndStopLifecycle(t *testing.T) {
	broker := singleton.New()
	consumer := &mockConsumer{}

	_ = broker.Register(processing.EventSubscriber{ID: "sub-1"})
	broker.AddConsumer("sub-1", consumer)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- broker.Start(ctx) }()

	cancel()
	err := <-done
	assert.ErrorIs(t, err, context.Canceled)
	assert.True(t, consumer.started)
}

func TestStopDrainsConsumers(t *testing.T) {
	broker := singleton.New()
	consumer := &mockConsumer{}
	broker.AddConsumer("sub-1", consumer)

	err := broker.Stop(context.Background())
	require.NoError(t, err)
	assert.True(t, consumer.drained)
}

func TestStatusReflectsLifecycle(t *testing.T) {
	broker := singleton.New()
	assert.Equal(t, processing.StatusStopped, broker.Status())
}
