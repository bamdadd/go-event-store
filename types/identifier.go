package types

// Targetable is a sealed interface for event source targets.
// The unexported marker method prevents external implementations.
type Targetable interface {
	targetable()
}

type StreamIdentifier struct {
	Category string
	Stream   string
}

type CategoryIdentifier struct {
	Category string
}

type LogIdentifier struct{}

func (StreamIdentifier) targetable()   {}
func (CategoryIdentifier) targetable() {}
func (LogIdentifier) targetable()      {}
