package clock

import "time"

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type StaticClock struct {
	Time time.Time
}

func (c StaticClock) Now() time.Time { return c.Time }
