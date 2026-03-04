package distributed_test

import (
	"context"
	"testing"

	"github.com/logicblocks/event-store/processing"
	"github.com/logicblocks/event-store/processing/broker/distributed"
	lockmem "github.com/logicblocks/event-store/processing/locks/memory"
	"github.com/stretchr/testify/assert"
)

func TestDistributedBrokerStartAndStop(t *testing.T) {
	lm := lockmem.New()
	broker := distributed.New(lm)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- broker.Start(ctx) }()

	cancel()
	err := <-done
	assert.ErrorIs(t, err, context.Canceled)
}

func TestDistributedBrokerStatusReflectsLifecycle(t *testing.T) {
	lm := lockmem.New()
	broker := distributed.New(lm)
	assert.Equal(t, processing.StatusStopped, broker.Status())
}

func TestDistributedBrokerRegisterDoesNotError(t *testing.T) {
	lm := lockmem.New()
	broker := distributed.New(lm)
	err := broker.Register(processing.EventSubscriber{ID: "sub-1"})
	assert.NoError(t, err)
}
