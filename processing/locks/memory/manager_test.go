package memory_test

import (
	"context"
	"testing"
	"time"

	lockmem "github.com/logicblocks/event-store/processing/locks/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTryLockUncontestedSucceeds(t *testing.T) {
	m := lockmem.New()
	lock, release, err := m.TryLock(context.Background(), "test-lock")
	defer release()
	require.NoError(t, err)
	assert.True(t, lock.Locked)
}

func TestTryLockContestedFails(t *testing.T) {
	m := lockmem.New()
	_, release1, _ := m.TryLock(context.Background(), "test-lock")
	defer release1()

	lock2, release2, err := m.TryLock(context.Background(), "test-lock")
	defer release2()
	require.NoError(t, err)
	assert.False(t, lock2.Locked)
}

func TestTryLockReleaseAllowsReacquisition(t *testing.T) {
	m := lockmem.New()
	_, release1, _ := m.TryLock(context.Background(), "test-lock")
	release1()

	lock2, release2, err := m.TryLock(context.Background(), "test-lock")
	defer release2()
	require.NoError(t, err)
	assert.True(t, lock2.Locked)
}

func TestWaitForLockSucceeds(t *testing.T) {
	m := lockmem.New()
	lock, release, err := m.WaitForLock(context.Background(), "test-lock")
	defer release()
	require.NoError(t, err)
	assert.True(t, lock.Locked)
}

func TestWaitForLockRespectsContextTimeout(t *testing.T) {
	m := lockmem.New()
	_, release1, _ := m.TryLock(context.Background(), "test-lock")
	defer release1()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	lock2, release2, err := m.WaitForLock(ctx, "test-lock")
	defer release2()
	assert.Error(t, err)
	assert.False(t, lock2.Locked)
	assert.True(t, lock2.TimedOut)
}

func TestDifferentLockNamesAreIndependent(t *testing.T) {
	m := lockmem.New()
	_, release1, _ := m.TryLock(context.Background(), "lock-a")
	defer release1()

	lock2, release2, err := m.TryLock(context.Background(), "lock-b")
	defer release2()
	require.NoError(t, err)
	assert.True(t, lock2.Locked)
}
