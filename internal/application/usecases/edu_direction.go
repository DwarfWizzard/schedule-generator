package usecases

import (
	"context"
	"errors"
	"fmt"
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

	MapDepartmentsByEduDirections(ctx context.Context, directionIDs uuid.UUIDs) (map[uuid.UUID]departments.Department, error)
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
	DepartmentName string
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
		EduDirection:   *direction,
		DepartmentName: department.Name,
	}, nil
}

type GetEduDirectionOutput struct {
	edudirections.EduDirection
	DepartmentName string
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

	department, err := uc.repo.GetDepartment(ctx, direction.DepartmentID)
	if err != nil {
		logger.Error("Get edu direction department error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &GetEduDirectionOutput{
		EduDirection:   *direction,
		DepartmentName: department.Name,
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

	directionIDs := make(uuid.UUIDs, len(directions))
	for i, direction := range directions {
		directionIDs[i] = direction.ID
	}

	departments, err := uc.repo.MapDepartmentsByEduDirections(ctx, directionIDs)
	if err != nil {
		logger.Error("Map departments by edu direction error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	for i, direction := range directions {
		department, ok := departments[direction.DepartmentID]
		if !ok {
			logger.Error(fmt.Sprintf("Department for direction %s not found", direction.ID))
			return nil, execerror.NewExecError(execerror.TypeInternal, nil)
		}

		result[i] = GetEduDirectionOutput{
			EduDirection:   direction,
			DepartmentName: department.Name,
		}
	}

	return result, nil
}

type UpdateEduDirectionInput struct {
	EduDirectionID uuid.UUID
	Name           *string
}

type UpdateEduDirectionOutput struct {
	GetEduDirectionOutput
}

// UpdateEduDirection
func (uc *EduDirectionUsecase) UpdateEduDirection(ctx context.Context, input UpdateEduDirectionInput) (*UpdateEduDirectionOutput, error) {
	logger := uc.logger

	direction, err := uc.repo.GetEduDirection(ctx, input.EduDirectionID)
	if err != nil {
		logger.Error("List edu direction error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu direction not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	department, err := uc.repo.GetDepartment(ctx, direction.DepartmentID)
	if err != nil {
		logger.Error("Get edu direction department error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if input.Name != nil {
		direction.Name = *input.Name
	}

	if err := direction.Validate(); err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveEduDirection(ctx, direction)
	if err != nil {
		logger.Error("Save edu edu direction error", "error", err)

		// TODO: use domain errors
		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu direction with provided external id already exists"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &UpdateEduDirectionOutput{
		GetEduDirectionOutput: GetEduDirectionOutput{
			EduDirection:   *direction,
			DepartmentName: department.Name,
		},
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
