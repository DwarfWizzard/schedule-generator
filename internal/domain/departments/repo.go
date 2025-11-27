package departments

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	SaveDepartment(ctx context.Context, d *Department) error
	GetDepartment(ctx context.Context, id uuid.UUID) (*Department, error)
	ListDepartment(ctx context.Context) ([]Department, error)
	DeleteDepartment(ctx context.Context, id uuid.UUID) error
}
