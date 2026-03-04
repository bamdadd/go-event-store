package memory

import (
	"context"
	"sync"

	"github.com/logicblocks/event-store/processing/locks"
)

type InMemoryLockManager struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func New() *InMemoryLockManager {
	return &InMemoryLockManager{locks: make(map[string]*sync.Mutex)}
}

func (m *InMemoryLockManager) getOrCreateLock(name string) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()
	lk, ok := m.locks[name]
	if !ok {
		lk = &sync.Mutex{}
		m.locks[name] = lk
	}
	return lk
}

func (m *InMemoryLockManager) TryLock(_ context.Context, name string) (locks.Lock, locks.LockReleaser, error) {
	lk := m.getOrCreateLock(name)
	locked := lk.TryLock()
	release := func() {
		if locked {
			lk.Unlock()
		}
	}
	return locks.Lock{Name: name, Locked: locked}, release, nil
}

func (m *InMemoryLockManager) WaitForLock(ctx context.Context, name string) (locks.Lock, locks.LockReleaser, error) {
	lk := m.getOrCreateLock(name)

	done := make(chan struct{})
	go func() {
		lk.Lock()
		close(done)
	}()

	select {
	case <-done:
		return locks.Lock{Name: name, Locked: true}, func() { lk.Unlock() }, nil
	case <-ctx.Done():
		return locks.Lock{Name: name, Locked: false, TimedOut: true}, func() {}, ctx.Err()
	}
}
