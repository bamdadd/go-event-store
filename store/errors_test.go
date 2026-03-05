package store_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/bamdadd/go-event-store/store"
	"github.com/stretchr/testify/assert"
)

func TestUnmetWriteConditionErrorIsErrUnmetWriteCondition(t *testing.T) {
	err := &store.UnmetWriteConditionError{Message: "position mismatch"}
	assert.True(t, errors.Is(err, store.ErrUnmetWriteCondition))
}

func TestUnmetWriteConditionErrorMessage(t *testing.T) {
	err := &store.UnmetWriteConditionError{Message: "stream not empty"}
	assert.Equal(t, "stream not empty", err.Error())
}

func TestUnmetWriteConditionErrorWrapped(t *testing.T) {
	inner := &store.UnmetWriteConditionError{Message: "inner"}
	wrapped := fmt.Errorf("outer: %w", inner)
	assert.True(t, errors.Is(wrapped, store.ErrUnmetWriteCondition))
}

func TestOtherErrorIsNotErrUnmetWriteCondition(t *testing.T) {
	err := errors.New("something else")
	assert.False(t, errors.Is(err, store.ErrUnmetWriteCondition))
}
