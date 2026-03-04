package projection

import "context"

type ProjectionStorageAdapter interface {
	Save(ctx context.Context, projection Projection) error
	FindOne(ctx context.Context, name, id string) (*Projection, error)
	FindMany(ctx context.Context, name string) ([]Projection, error)
}
