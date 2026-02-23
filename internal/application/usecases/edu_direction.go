package usecases

import (
	"context"
	"errors"
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
	Name string
}

type CreateEduDirectionOutput struct {
	edudirections.EduDirection
}

// CreateEduDirection
func (uc *EduDirectionUsecase) CreateEduDirection(ctx context.Context, input CreateEduDirectionInput, user *users.User) (*CreateEduDirectionOutput, error) {
	logger := uc.logger

	if !uc.authSvc.IsAdmin(user) {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have acces to usecase"))
	}

	direction, err := edudirections.NewEduDirection(input.Name)
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

	return &GetEduDirectionOutput{
		EduDirection: *direction,
	}, nil
}

// ListEduDirection
func (uc *EduDirectionUsecase) ListEduDirection(ctx context.Context, user *users.User) ([]GetEduDirectionOutput, error) {
	logger := uc.logger

	var directions []edudirections.EduDirection
	var listErr error

	directions, listErr = uc.repo.ListEduDirection(ctx)
	if listErr != nil {
		logger.Error("List edu direction error", "error", listErr)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	result := make([]GetEduDirectionOutput, len(directions))

	directionIDs := make(uuid.UUIDs, len(directions))
	for i, direction := range directions {
		directionIDs[i] = direction.ID
	}

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
	GetEduDirectionOutput
}

// UpdateEduDirection
func (uc *EduDirectionUsecase) UpdateEduDirection(ctx context.Context, input UpdateEduDirectionInput, user *users.User) (*UpdateEduDirectionOutput, error) {
	logger := uc.logger

	if !uc.authSvc.IsAdmin(user) {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have acces to usecase"))
	}

	direction, err := uc.repo.GetEduDirection(ctx, input.EduDirectionID)
	if err != nil {
		logger.Error("List edu direction error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu direction not found"))
		}

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
			EduDirection: *direction,
		},
	}, nil
}

// DeleteEduDirection
func (uc *EduDirectionUsecase) DeleteEduDirection(ctx context.Context, directionID uuid.UUID, user *users.User) error {
	logger := uc.logger

	if !uc.authSvc.IsAdmin(user) {
		return execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have acces to usecase"))
	}

	err := uc.repo.DeleteEduDirection(ctx, directionID)
	if err != nil {
		logger.Error("Delete edu direction error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}
