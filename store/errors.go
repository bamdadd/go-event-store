package store

import "errors"

var ErrUnmetWriteCondition = errors.New("unmet write condition")

type UnmetWriteConditionError struct {
	Message string
}

func (e *UnmetWriteConditionError) Error() string { return e.Message }

func (e *UnmetWriteConditionError) Is(target error) bool {
	return target == ErrUnmetWriteCondition
}
