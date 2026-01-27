package cabinets

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	SaveCabinet(ctx context.Context, c *Cabinet) error
	GetCabinet(ctx context.Context, id uuid.UUID) (*Cabinet, error)
	ListCabinet(ctx context.Context) ([]Cabinet, error)
	DeleteCabinet(ctx context.Context, id uuid.UUID) error
}
