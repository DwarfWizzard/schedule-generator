package usecases

import (
	"context"
	"errors"
	"log/slog"

	edudirections "schedule-generator/internal/domain/edu_directions"
	eduplans "schedule-generator/internal/domain/edu_plans"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/pkg/execerror"

	"github.com/google/uuid"
)

type EduPlanUsecaseRepo interface {
	edudirections.Repository
	eduplans.Repository
}

type EduPlanUsecase struct {
	repo   EduPlanUsecaseRepo
	logger *slog.Logger
}

func NewEduPlanUsecase(repo EduPlanUsecaseRepo, logger *slog.Logger) *EduPlanUsecase {
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
}

// CreateEduPlan
func (uc *EduPlanUsecase) CreateEduPlan(ctx context.Context, input CreateEduPlanInput) (*CreateEduPlanOutput, error) {
	logger := uc.logger

	direction, err := uc.repo.GetEduDirection(ctx, input.DirectionID)
	if err != nil {
		logger.Error("Get edu direction error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu direction not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
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
		EduPlan: *eduplan,
	}, nil
}

type GetEduPlanOutput struct {
	eduplans.EduPlan
}

// GetEduPlan
func (uc *EduPlanUsecase) GetEduPlan(ctx context.Context, eduplanID uuid.UUID) (*GetEduPlanOutput, error) {
	logger := uc.logger

	eduplan, err := uc.repo.GetEduPlan(ctx, eduplanID)
	if err != nil {
		logger.Error("List eduplan error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("eduplan not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &GetEduPlanOutput{
		EduPlan: *eduplan,
	}, nil
}

// ListEduPlan
func (uc *EduPlanUsecase) ListEduPlan(ctx context.Context) ([]GetEduPlanOutput, error) {
	logger := uc.logger

	eduplans, err := uc.repo.ListEduPlan(ctx)
	if err != nil {
		logger.Error("List eduplan error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	result := make([]GetEduPlanOutput, len(eduplans))

	for i, eduplan := range eduplans {
		result[i] = GetEduPlanOutput{
			EduPlan: eduplan,
		}
	}

	return result, nil
}

// DeleteEduPlan
func (uc *EduPlanUsecase) DeleteEduPlan(ctx context.Context, eduplanID uuid.UUID) error {
	logger := uc.logger

	err := uc.repo.DeleteEduPlan(ctx, eduplanID)
	if err != nil {
		logger.Error("Delete edu eduplan error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}
