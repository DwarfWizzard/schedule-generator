package eduplans

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	SaveEduPlan(ctx context.Context, plan *EduPlan) error
	GetEduPlan(ctx context.Context, id uuid.UUID) (*EduPlan, error)
	ListEduPlan(ctx context.Context) ([]EduPlan, error)
	ListEduPlanByFaculty(ctx context.Context, facultyID uuid.UUID) ([]EduPlan, error)
	DeleteEduPlan(ctx context.Context, id uuid.UUID) error
}
