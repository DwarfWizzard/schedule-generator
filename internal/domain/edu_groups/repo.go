package edugroups

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	SaveEduGroup(ctx context.Context, group *EduGroup) error
	GetEduGroup(ctx context.Context, id uuid.UUID) (*EduGroup, error)
	GetEduGroupByNumber(ctx context.Context, number string) (*EduGroup, error)
	ListEduGroup(ctx context.Context) ([]EduGroup, error)
	ListEduGroupByFaculty(ctx context.Context, facultyID uuid.UUID) ([]EduGroup, error)
	DeleteEduGroup(ctx context.Context, id uuid.UUID) error
}
