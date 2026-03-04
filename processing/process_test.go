package processing_test

import (
	"testing"

	"github.com/logicblocks/event-store/processing"
	"github.com/stretchr/testify/assert"
)

func TestAtomicStatusDefaultIsStopped(t *testing.T) {
	var s processing.AtomicStatus
	assert.Equal(t, processing.StatusStopped, s.Get())
}

func TestAtomicStatusSetAndGet(t *testing.T) {
	var s processing.AtomicStatus
	s.Set(processing.StatusRunning)
	assert.Equal(t, processing.StatusRunning, s.Get())
}

func TestAtomicStatusTransitions(t *testing.T) {
	var s processing.AtomicStatus
	s.Set(processing.StatusStarting)
	assert.Equal(t, processing.StatusStarting, s.Get())
	s.Set(processing.StatusRunning)
	assert.Equal(t, processing.StatusRunning, s.Get())
	s.Set(processing.StatusStopping)
	assert.Equal(t, processing.StatusStopping, s.Get())
	s.Set(processing.StatusStopped)
	assert.Equal(t, processing.StatusStopped, s.Get())
}
