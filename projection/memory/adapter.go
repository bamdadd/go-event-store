package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/bamdadd/go-event-store/projection"
)

type InMemoryProjectionStorageAdapter struct {
	mu          sync.RWMutex
	projections map[string]projection.Projection
}

func New() *InMemoryProjectionStorageAdapter {
	return &InMemoryProjectionStorageAdapter{
		projections: make(map[string]projection.Projection),
	}
}

func (a *InMemoryProjectionStorageAdapter) Save(_ context.Context, p projection.Projection) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.projections[key(p.Name, p.ID)] = p
	return nil
}

func (a *InMemoryProjectionStorageAdapter) FindOne(_ context.Context, name, id string) (*projection.Projection, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	p, ok := a.projections[key(name, id)]
	if !ok {
		return nil, nil
	}
	return &p, nil
}

func (a *InMemoryProjectionStorageAdapter) FindMany(_ context.Context, name string) ([]projection.Projection, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	var results []projection.Projection
	for _, p := range a.projections {
		if p.Name == name {
			results = append(results, p)
		}
	}
	return results, nil
}

func key(name, id string) string {
	return fmt.Sprintf("%s:%s", name, id)
}
