package usecases

import (
	"context"
	"errors"
	"log/slog"

	"schedule-generator/internal/domain/departments"
	edudirections "schedule-generator/internal/domain/edu_directions"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/pkg/execerror"

	"github.com/google/uuid"
)

type EduDirectionUsecaseRepo interface {
	departments.Repository
	edudirections.Repository
}

type EduDirectionUsecase struct {
	repo   EduDirectionUsecaseRepo
	logger *slog.Logger
}

func NewEduDirectionUsecase(repo EduDirectionUsecaseRepo, logger *slog.Logger) *EduDirectionUsecase {
	return &EduDirectionUsecase{
		repo:   repo,
		logger: logger,
	}
}

type CreateEduDirectionInput struct {
	DepartmentID uuid.UUID
	Name         string
}

type CreateEduDirectionOutput struct {
	edudirections.EduDirection
}

// CreateEduDirection
func (uc *EduDirectionUsecase) CreateEduDirection(ctx context.Context, input CreateEduDirectionInput) (*CreateEduDirectionOutput, error) {
	logger := uc.logger

	department, err := uc.repo.GetDepartment(ctx, input.DepartmentID)
	if err != nil {
		logger.Error("Get department error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("department not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	direction, err := edudirections.NewEduDirection(department.ID, input.Name)
	if err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveEduDirection(ctx, direction)
	if err != nil {
		logger.Error("Save edu direction error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &CreateEduDirectionOutput{
		EduDirection: *direction,
	}, nil
}

type GetEduDirectionOutput struct {
	edudirections.EduDirection
}

// GetEduDirection
func (uc *EduDirectionUsecase) GetEduDirection(ctx context.Context, directionID uuid.UUID) (*GetEduDirectionOutput, error) {
	logger := uc.logger

	direction, err := uc.repo.GetEduDirection(ctx, directionID)
	if err != nil {
		logger.Error("Get edu direction error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu direction not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &GetEduDirectionOutput{
		EduDirection: *direction,
	}, nil
}

// ListEduDirection
func (uc *EduDirectionUsecase) ListEduDirection(ctx context.Context) ([]GetEduDirectionOutput, error) {
	logger := uc.logger

	directions, err := uc.repo.ListEduDirection(ctx)
	if err != nil {
		logger.Error("List edu direction error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	result := make([]GetEduDirectionOutput, len(directions))

	for i, direction := range directions {
		result[i] = GetEduDirectionOutput{
			EduDirection: direction,
		}
	}

	return result, nil
}

type UpdateEduDirectionInput struct {
	EduDirectionID uuid.UUID
	Name           *string
}

type UpdateEduDirectionOutput struct {
	edudirections.EduDirection
}

// UpdateEduDirection
func (uc *EduDirectionUsecase) UpdateEduDirection(ctx context.Context, input UpdateEduDirectionInput) (*UpdateEduDirectionOutput, error) {
	logger := uc.logger

	edudirection, err := uc.repo.GetEduDirection(ctx, input.EduDirectionID)
	if err != nil {
		logger.Error("List edu direction error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu direction not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if input.Name != nil {
		edudirection.Name = *input.Name
	}

	if err := edudirection.Validate(); err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveEduDirection(ctx, edudirection)
	if err != nil {
		logger.Error("Save edu edu direction error", "error", err)

		// TODO: use domain errors
		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu direction with provided external id already exists"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &UpdateEduDirectionOutput{
		EduDirection: *edudirection,
	}, nil
}

// DeleteEduDirection
func (uc *EduDirectionUsecase) DeleteEduDirection(ctx context.Context, directionID uuid.UUID) error {
	logger := uc.logger

	err := uc.repo.DeleteEduDirection(ctx, directionID)
	if err != nil {
		logger.Error("Delete edu direction error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}
