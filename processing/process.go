package processing

import "sync/atomic"

type ProcessStatus int32

const (
	StatusStopped  ProcessStatus = iota
	StatusStarting
	StatusRunning
	StatusStopping
)

type AtomicStatus struct {
	v atomic.Int32
}

func (s *AtomicStatus) Set(status ProcessStatus) { s.v.Store(int32(status)) }
func (s *AtomicStatus) Get() ProcessStatus       { return ProcessStatus(s.v.Load()) }
