package watermill

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/logicblocks/event-store/processing"
)

type WatermillConsumer struct {
	subscriber message.Subscriber
	topic      string
	processor  processing.EventProcessor
	status     processing.AtomicStatus
}

func NewConsumer(
	subscriber message.Subscriber,
	topic string,
	processor processing.EventProcessor,
) *WatermillConsumer {
	return &WatermillConsumer{
		subscriber: subscriber,
		topic:      topic,
		processor:  processor,
	}
}

func (c *WatermillConsumer) Start(ctx context.Context) error {
	c.status.Set(processing.StatusRunning)
	defer c.status.Set(processing.StatusStopped)

	messages, err := c.subscriber.Subscribe(ctx, c.topic)
	if err != nil {
		return err
	}

	for msg := range messages {
		event, err := MessageToStoredEvent(msg)
		if err != nil {
			msg.Nack()
			continue
		}
		if err := c.processor.Process(ctx, event); err != nil {
			msg.Nack()
			continue
		}
		msg.Ack()
	}
	return nil
}

func (c *WatermillConsumer) Stop(_ context.Context) error {
	c.status.Set(processing.StatusStopping)
	return c.subscriber.Close()
}

func (c *WatermillConsumer) Drain(ctx context.Context) error {
	return c.Stop(ctx)
}

func (c *WatermillConsumer) Status() processing.ProcessStatus {
	return c.status.Get()
}
