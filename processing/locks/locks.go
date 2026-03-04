package locks

import (
	"context"
	"time"
)

type Lock struct {
	Name     string
	Locked   bool
	TimedOut bool
	WaitTime time.Duration
}

type LockReleaser func()

type LockManager interface {
	TryLock(ctx context.Context, name string) (Lock, LockReleaser, error)
	WaitForLock(ctx context.Context, name string) (Lock, LockReleaser, error)
}
