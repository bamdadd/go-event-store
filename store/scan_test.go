package store_test

import (
	"testing"

	"github.com/logicblocks/event-store/store"
	"github.com/stretchr/testify/assert"
)

func TestBuildScanConfigNoOptions(t *testing.T) {
	cfg := store.BuildScanConfig()
	assert.Empty(t, cfg.Constraints)
}

func TestBuildScanConfigWithSequenceNumberAfter(t *testing.T) {
	cfg := store.BuildScanConfig(store.WithSequenceNumberAfter(42))
	assert.Len(t, cfg.Constraints, 1)
	c, ok := cfg.Constraints[0].(store.SequenceNumberAfterConstraint)
	assert.True(t, ok)
	assert.Equal(t, 42, c.SequenceNumber)
}

func TestBuildScanConfigMultipleOptions(t *testing.T) {
	cfg := store.BuildScanConfig(
		store.WithSequenceNumberAfter(10),
		store.WithSequenceNumberAfter(20),
	)
	assert.Len(t, cfg.Constraints, 2)
}
