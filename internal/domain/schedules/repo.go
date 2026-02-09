package schedules

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetSchedule(ctx context.Context, id uuid.UUID) (*Schedule, error)
	ListSchedule(ctx context.Context) ([]Schedule, error)
	ListScheduleByEduGroup(ctx context.Context, groupID uuid.UUID) ([]Schedule, error)
	ListScheduleByFacultyID(ctx context.Context, facultyID uuid.UUID) ([]Schedule, error)
	SaveSchedule(ctx context.Context, schedule *Schedule) error
	DeleteSchedule(ctx context.Context, id uuid.UUID) error
}
