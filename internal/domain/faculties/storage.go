package faculties

import (
	"context"

	"github.com/google/uuid"
)

type Faculty struct {
	ID   uuid.UUID
	Name string
}

type Repository interface {
	SaveFaculty(ctx context.Context, f *Faculty) error
	GetFaculty(ctx context.Context, id uuid.UUID) (*Faculty, error)
	ListFaculty(ctx context.Context) ([]Faculty, error)
	DeleteFaculty(ctx context.Context, id uuid.UUID) error
}
