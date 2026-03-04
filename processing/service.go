package processing

import (
	"context"
	"time"
)

type PollingService struct {
	fn       func(ctx context.Context) error
	interval time.Duration
	policy   RetryPolicy
	status   AtomicStatus
}

func NewPollingService(
	fn func(ctx context.Context) error,
	interval time.Duration,
	policy RetryPolicy,
) *PollingService {
	return &PollingService{
		fn:       fn,
		interval: interval,
		policy:   policy,
	}
}

func (s *PollingService) Start(ctx context.Context) error {
	s.status.Set(StatusRunning)
	defer s.status.Set(StatusStopped)

	attempt := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(s.interval):
			if err := s.fn(ctx); err != nil {
				retry, delay := s.policy.ShouldRetry(attempt, err)
				if !retry {
					return err
				}
				attempt++
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
				}
			} else {
				attempt = 0
			}
		}
	}
}

func (s *PollingService) Status() ProcessStatus { return s.status.Get() }
