package projection

import "context"

type ProjectionStore struct {
	adapter ProjectionStorageAdapter
}

func NewProjectionStore(adapter ProjectionStorageAdapter) *ProjectionStore {
	return &ProjectionStore{adapter: adapter}
}

func (ps *ProjectionStore) Save(ctx context.Context, p Projection) error {
	return ps.adapter.Save(ctx, p)
}

func (ps *ProjectionStore) FindOne(ctx context.Context, name, id string) (*Projection, error) {
	return ps.adapter.FindOne(ctx, name, id)
}

func (ps *ProjectionStore) FindMany(ctx context.Context, name string) ([]Projection, error) {
	return ps.adapter.FindMany(ctx, name)
}
