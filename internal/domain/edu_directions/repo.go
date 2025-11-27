package edudirections

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	SaveEduDirection(ctx context.Context, d *EduDirection) error
	GetEduDirection(ctx context.Context, id uuid.UUID) (*EduDirection, error)
	ListEduDirection(ctx context.Context) ([]EduDirection, error)
	DeleteEduDirection(ctx context.Context, id uuid.UUID) error
}
