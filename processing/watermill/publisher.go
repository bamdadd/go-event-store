package watermill

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/bamdadd/go-event-store/types"
)

type EventStorePublisher struct {
	logger watermill.LoggerAdapter
}

func NewEventStorePublisher(logger watermill.LoggerAdapter) *EventStorePublisher {
	return &EventStorePublisher{logger: logger}
}

func StoredEventToMessage(event types.StoredEvent) (*message.Message, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	msg := message.NewMessage(event.ID, payload)
	msg.Metadata.Set("event_name", event.Name)
	msg.Metadata.Set("stream_category", event.Category)
	msg.Metadata.Set("stream_id", event.Stream)
	return msg, nil
}

func MessageToStoredEvent(msg *message.Message) (types.StoredEvent, error) {
	var event types.StoredEvent
	err := json.Unmarshal(msg.Payload, &event)
	return event, err
}
