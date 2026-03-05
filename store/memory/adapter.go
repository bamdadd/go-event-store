package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/types"
)

type Option func(*InMemoryEventStorageAdapter)

func WithGuarantee(g store.SerialisationGuarantee) Option {
	return func(a *InMemoryEventStorageAdapter) { a.guarantee = g }
}

type InMemoryEventStorageAdapter struct {
	db        *eventsDB
	guarantee store.SerialisationGuarantee
}

func New(opts ...Option) *InMemoryEventStorageAdapter {
	a := &InMemoryEventStorageAdapter{
		db:        newEventsDB(),
		guarantee: store.GuaranteeLog,
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

func (a *InMemoryEventStorageAdapter) SaveToStream(
	_ context.Context,
	target types.StreamIdentifier,
	events []types.NewEvent,
	condition store.WriteCondition,
) ([]types.StoredEvent, error) {
	if len(events) == 0 {
		return nil, nil
	}

	a.db.mu.Lock()
	defer a.db.mu.Unlock()

	currentPos := a.currentPositionLocked(target)
	ok, err := store.EvaluateCondition(condition, currentPos)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, &store.UnmetWriteConditionError{
			Message: fmt.Sprintf("write condition not met for stream %s/%s", target.Category, target.Stream),
		}
	}

	nextPos := 0
	if currentPos != nil {
		nextPos = *currentPos + 1
	}

	stored := make([]types.StoredEvent, 0, len(events))
	for i, e := range events {
		payload, marshalErr := json.Marshal(e.Payload)
		if marshalErr != nil {
			return nil, marshalErr
		}

		seqNum := a.db.nextSeqNum
		a.db.nextSeqNum++

		se := types.StoredEvent{
			ID:             fmt.Sprintf("%s-%s-%d", target.Category, target.Stream, nextPos+i),
			Name:           e.Name,
			Stream:         target.Stream,
			Category:       target.Category,
			Position:       nextPos + i,
			SequenceNumber: seqNum,
			Payload:        payload,
			ObservedAt:     e.ObservedAt,
			OccurredAt:     e.OccurredAt,
		}

		a.db.events[seqNum] = se
		a.db.logIndex = append(a.db.logIndex, seqNum)

		if a.db.streamIndex[target.Category] == nil {
			a.db.streamIndex[target.Category] = make(map[string][]int)
		}
		a.db.streamIndex[target.Category][target.Stream] = append(
			a.db.streamIndex[target.Category][target.Stream], seqNum,
		)
		a.db.categoryIndex[target.Category] = append(a.db.categoryIndex[target.Category], seqNum)

		stored = append(stored, se)
	}

	return stored, nil
}

func (a *InMemoryEventStorageAdapter) SaveToCategory(
	ctx context.Context,
	target types.CategoryIdentifier,
	streams map[string]store.StreamPublishDefinition,
) (map[string][]types.StoredEvent, error) {
	result := make(map[string][]types.StoredEvent)
	for streamName, def := range streams {
		streamTarget := types.StreamIdentifier{Category: target.Category, Stream: streamName}
		stored, err := a.SaveToStream(ctx, streamTarget, def.Events, def.Condition)
		if err != nil {
			return nil, err
		}
		result[streamName] = stored
	}
	return result, nil
}

func (a *InMemoryEventStorageAdapter) Latest(_ context.Context, target types.Targetable) (*types.StoredEvent, error) {
	a.db.mu.RLock()
	defer a.db.mu.RUnlock()

	seqNums := a.seqNumsForTargetLocked(target)
	if len(seqNums) == 0 {
		return nil, nil
	}
	e := a.db.events[seqNums[len(seqNums)-1]]
	return &e, nil
}

func (a *InMemoryEventStorageAdapter) Scan(
	ctx context.Context,
	target types.Targetable,
	opts ...store.ScanOption,
) iter.Seq2[types.StoredEvent, error] {
	cfg := store.BuildScanConfig(opts...)

	return func(yield func(types.StoredEvent, error) bool) {
		a.db.mu.RLock()
		seqNums := a.seqNumsForTargetLocked(target)
		snapshot := make([]types.StoredEvent, 0, len(seqNums))
		for _, sn := range seqNums {
			snapshot = append(snapshot, a.db.events[sn])
		}
		a.db.mu.RUnlock()

		for _, e := range snapshot {
			if !matchesConstraints(e, cfg.Constraints) {
				continue
			}
			if ctx.Err() != nil {
				yield(types.StoredEvent{}, ctx.Err())
				return
			}
			if !yield(e, nil) {
				return
			}
		}
	}
}

func (a *InMemoryEventStorageAdapter) ClearAll() {
	a.db.clear()
}

func (a *InMemoryEventStorageAdapter) currentPositionLocked(target types.StreamIdentifier) *int {
	streams, ok := a.db.streamIndex[target.Category]
	if !ok {
		return nil
	}
	seqNums, ok := streams[target.Stream]
	if !ok || len(seqNums) == 0 {
		return nil
	}
	lastSeq := seqNums[len(seqNums)-1]
	pos := a.db.events[lastSeq].Position
	return &pos
}

func (a *InMemoryEventStorageAdapter) seqNumsForTargetLocked(target types.Targetable) []int {
	switch t := target.(type) {
	case types.StreamIdentifier:
		if streams, ok := a.db.streamIndex[t.Category]; ok {
			return streams[t.Stream]
		}
		return nil
	case types.CategoryIdentifier:
		return a.db.categoryIndex[t.Category]
	case types.LogIdentifier:
		return a.db.logIndex
	default:
		return nil
	}
}

func matchesConstraints(e types.StoredEvent, constraints []store.QueryConstraint) bool {
	for _, c := range constraints {
		switch qc := c.(type) {
		case store.SequenceNumberAfterConstraint:
			if e.SequenceNumber <= qc.SequenceNumber {
				return false
			}
		}
	}
	return true
}
