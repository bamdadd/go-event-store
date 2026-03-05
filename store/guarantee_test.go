package store_test

import (
	"testing"

	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/types"
	"github.com/stretchr/testify/assert"
)

func TestLogGuaranteeLockName(t *testing.T) {
	name := store.GuaranteeLog.LockName("ns", types.StreamIdentifier{Category: "cat", Stream: "str"})
	assert.Equal(t, "ns", name)
}

func TestCategoryGuaranteeLockName(t *testing.T) {
	name := store.GuaranteeCategory.LockName("ns", types.StreamIdentifier{Category: "cat", Stream: "str"})
	assert.Equal(t, "ns.cat", name)
}

func TestStreamGuaranteeLockName(t *testing.T) {
	name := store.GuaranteeStream.LockName("ns", types.StreamIdentifier{Category: "cat", Stream: "str"})
	assert.Equal(t, "ns.cat.str", name)
}

func TestLockDigestIsDeterministic(t *testing.T) {
	a := store.LockDigest("event_store.orders.order-1")
	b := store.LockDigest("event_store.orders.order-1")
	assert.Equal(t, a, b)
}

func TestLockDigestDiffersForDifferentNames(t *testing.T) {
	a := store.LockDigest("event_store.orders.order-1")
	b := store.LockDigest("event_store.orders.order-2")
	assert.NotEqual(t, a, b)
}
