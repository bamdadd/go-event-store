package store

import "fmt"

type WriteCondition interface {
	isWriteCondition()
}

type noCondition struct{}
type positionIsCondition struct{ position *int }
type emptyStreamCondition struct{}
type andCondition struct{ conditions []WriteCondition }
type orCondition struct{ conditions []WriteCondition }

func (noCondition) isWriteCondition()         {}
func (positionIsCondition) isWriteCondition()  {}
func (emptyStreamCondition) isWriteCondition() {}
func (andCondition) isWriteCondition()         {}
func (orCondition) isWriteCondition()          {}

func NoCondition() WriteCondition       { return noCondition{} }
func PositionIs(pos *int) WriteCondition { return positionIsCondition{position: pos} }
func StreamIsEmpty() WriteCondition      { return emptyStreamCondition{} }

func And(a, b WriteCondition, rest ...WriteCondition) WriteCondition {
	return andCondition{conditions: append([]WriteCondition{a, b}, rest...)}
}

func Or(a, b WriteCondition, rest ...WriteCondition) WriteCondition {
	return orCondition{conditions: append([]WriteCondition{a, b}, rest...)}
}

func EvaluateCondition(cond WriteCondition, currentPosition *int) (bool, error) {
	switch c := cond.(type) {
	case noCondition:
		return true, nil
	case positionIsCondition:
		if c.position == nil && currentPosition == nil {
			return true, nil
		}
		if c.position == nil || currentPosition == nil {
			return false, nil
		}
		return *c.position == *currentPosition, nil
	case emptyStreamCondition:
		return currentPosition == nil, nil
	case andCondition:
		for _, sub := range c.conditions {
			ok, err := EvaluateCondition(sub, currentPosition)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil
	case orCondition:
		for _, sub := range c.conditions {
			ok, err := EvaluateCondition(sub, currentPosition)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unknown WriteCondition type: %T", cond)
	}
}
