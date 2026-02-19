package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"schedule-generator/internal/application/services"
	"schedule-generator/internal/domain/departments"
	edudirections "schedule-generator/internal/domain/edu_directions"
	"schedule-generator/internal/domain/users"
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
	repo    EduDirectionUsecaseRepo
	authSvc *services.AuthorizationService
	logger  *slog.Logger
}

func NewEduDirectionUsecase(authSvc *services.AuthorizationService, repo EduDirectionUsecaseRepo, logger *slog.Logger) *EduDirectionUsecase {
	return &EduDirectionUsecase{
		repo:    repo,
		authSvc: authSvc,
		logger:  logger,
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
func (uc *EduDirectionUsecase) CreateEduDirection(ctx context.Context, input CreateEduDirectionInput, user *users.User) (*CreateEduDirectionOutput, error) {
	logger := uc.logger

	department, err := uc.repo.GetDepartment(ctx, input.DepartmentID)
	if err != nil {
		logger.Error("Get department error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("department not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if ok, err := uc.authSvc.HaveAccessToDepartment(ctx, department, user); err != nil {
		logger.Error("Check access to department error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	} else if !ok {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have access to department"))
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
func (uc *EduDirectionUsecase) GetEduDirection(ctx context.Context, directionID uuid.UUID, user *users.User) (*GetEduDirectionOutput, error) {
	logger := uc.logger

	direction, err := uc.repo.GetEduDirection(ctx, directionID)
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
func (uc *EduDirectionUsecase) ListEduDirection(ctx context.Context, user *users.User) ([]GetEduDirectionOutput, error) {
	logger := uc.logger

	var directions []edudirections.EduDirection
	var listErr error

	if uc.authSvc.IsAdmin(user) {
		directions, listErr = uc.repo.ListEduDirection(ctx)
	} else {
		if user.FacultyID == nil {
			return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user not accociated with any faculty"))
		}

		directions, listErr = uc.repo.ListEduDirectionByFaculty(ctx, *user.FacultyID)
	}

	if listErr != nil {
		logger.Error("List edu direction error", "error", listErr)
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
func (uc *EduDirectionUsecase) UpdateEduDirection(ctx context.Context, input UpdateEduDirectionInput, user *users.User) (*UpdateEduDirectionOutput, error) {
	logger := uc.logger

	direction, err := uc.repo.GetEduDirection(ctx, input.EduDirectionID)
	if err != nil {
		logger.Error("List edu direction error", "error", err)
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
func (uc *EduDirectionUsecase) DeleteEduDirection(ctx context.Context, directionID uuid.UUID, user *users.User) error {
	logger := uc.logger

	direction, err := uc.repo.GetEduDirection(ctx, directionID)
	if err != nil {
		logger.Error("List edu direction error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu direction not found"))
		}

		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if ok, err := uc.authSvc.HaveAccessToEduDirection(ctx, direction, user); err != nil {
		logger.Error("Check access to direction error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	} else if !ok {
		return execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have access to direction"))
	}

	err = uc.repo.DeleteEduDirection(ctx, directionID)
	if err != nil {
		logger.Error("Delete edu direction error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}
