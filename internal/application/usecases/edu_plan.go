package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"schedule-generator/internal/application/services"
	edudirections "schedule-generator/internal/domain/edu_directions"
	eduplans "schedule-generator/internal/domain/edu_plans"
	"schedule-generator/internal/domain/users"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/pkg/execerror"

	"github.com/google/uuid"
)

type EduPlanUsecaseRepo interface {
	edudirections.Repository
	eduplans.Repository

	MapEduDirectionByEduPlans(ctx context.Context, plansIDs uuid.UUIDs) (map[uuid.UUID]edudirections.EduDirection, error)
}

type EduPlanUsecase struct {
	repo    EduPlanUsecaseRepo
	authSvc *services.AuthorizationService
	logger  *slog.Logger
}

func NewEduPlanUsecase(authSvc *services.AuthorizationService, repo EduPlanUsecaseRepo, logger *slog.Logger) *EduPlanUsecase {
	return &EduPlanUsecase{
		repo:   repo,
		logger: logger,
	}
}

type CreateEduPlanInput struct {
	DirectionID uuid.UUID
	Profile     string
	Year        int64
}

type CreateEduPlanOutput struct {
	eduplans.EduPlan
	DirectionName string
}

// CreateEduPlan
func (uc *EduPlanUsecase) CreateEduPlan(ctx context.Context, input CreateEduPlanInput, user *users.User) (*CreateEduPlanOutput, error) {
	logger := uc.logger

	direction, err := uc.repo.GetEduDirection(ctx, input.DirectionID)
	if err != nil {
		logger.Error("Get edu direction error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu direction not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if ok, err := uc.authSvc.HaveAccessToEduDirection(ctx, direction, user); err != nil {
		logger.Error("Check access to direction error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	} else if !ok {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have access to direction"))
	}

	eduplan, err := eduplans.NewEduPlan(direction.ID, input.Profile, input.Year)
	if err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveEduPlan(ctx, eduplan)
	if err != nil {
		logger.Error("Save eduplan error", "error", err)

		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu plan with provided direction, year and profile already exists"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &CreateEduPlanOutput{
		EduPlan:       *eduplan,
		DirectionName: direction.Name,
	}, nil
}

type GetEduPlanOutput struct {
	eduplans.EduPlan
	DirectionName string
}

// GetEduPlan
func (uc *EduPlanUsecase) GetEduPlan(ctx context.Context, eduplanID uuid.UUID, user *users.User) (*GetEduPlanOutput, error) {
	logger := uc.logger

	eduplan, err := uc.repo.GetEduPlan(ctx, eduplanID)
	if err != nil {
		logger.Error("List eduplan error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("eduplan not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if ok, err := uc.authSvc.HaveAccessToEduPlan(ctx, eduplan, user); err != nil {
		logger.Error("Check access to edu plan error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	} else if !ok {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have access to edu plan"))
	}

	direction, err := uc.repo.GetEduDirection(ctx, eduplan.DirectionID)
	if err != nil {
		logger.Error("Get eduplans direction error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &GetEduPlanOutput{
		EduPlan:       *eduplan,
		DirectionName: direction.Name,
	}, nil
}

// ListEduPlan
func (uc *EduPlanUsecase) ListEduPlan(ctx context.Context, user *users.User) ([]GetEduPlanOutput, error) {
	logger := uc.logger

	var plans []eduplans.EduPlan
	var listErr error

	if uc.authSvc.IsAdmin(user) {
		plans, listErr = uc.repo.ListEduPlan(ctx)
	} else {
		if user.FacultyID == nil {
			return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user not accociated with any faculty"))
		}

		plans, listErr = uc.repo.ListEduPlanByFaculty(ctx, *user.FacultyID)
	}

	if listErr != nil {
		logger.Error("List plans error", "error", listErr)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	planIDs := make(uuid.UUIDs, len(plans))
	for i, plan := range plans {
		planIDs[i] = plan.ID
	}

	directions, err := uc.repo.MapEduDirectionByEduPlans(ctx, planIDs)
	if err != nil {
		logger.Error("Map edu directions error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	result := make([]GetEduPlanOutput, len(plans))

	for i, plan := range plans {
		direction, ok := directions[plan.DirectionID]
		if !ok {
			logger.Error(fmt.Sprintf("Edu direction for edu plan %s not found", plan.ID))
			return nil, execerror.NewExecError(execerror.TypeInternal, nil)
		}

		result[i] = GetEduPlanOutput{
			EduPlan:       plan,
			DirectionName: direction.Name,
		}
	}

	return result, nil
}

// DeleteEduPlan
func (uc *EduPlanUsecase) DeleteEduPlan(ctx context.Context, eduplanID uuid.UUID, user *users.User) error {
	logger := uc.logger

	eduplan, err := uc.repo.GetEduPlan(ctx, eduplanID)
	if err != nil {
		logger.Error("List eduplan error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("eduplan not found"))
		}

		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if ok, err := uc.authSvc.HaveAccessToEduPlan(ctx, eduplan, user); err != nil {
		logger.Error("Check access to edu plan error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	} else if !ok {
		return execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have access to edu plan"))
	}

	err = uc.repo.DeleteEduPlan(ctx, eduplanID)
	if err != nil {
		logger.Error("Delete edu eduplan error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}
