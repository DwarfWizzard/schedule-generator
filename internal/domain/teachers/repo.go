package teachers

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	SaveTeacher(ctx context.Context, t *Teacher) error
	GetTeacher(ctx context.Context, id uuid.UUID) (*Teacher, error)
	ListTeacher(ctx context.Context) ([]Teacher, error)
	DeleteTeacher(ctx context.Context, id uuid.UUID) error
}
