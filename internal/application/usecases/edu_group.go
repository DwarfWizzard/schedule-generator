package usecases

import (
	"context"
	"errors"
	"log/slog"

	"schedule-generator/internal/application/services"
	edugroups "schedule-generator/internal/domain/edu_groups"
	eduplans "schedule-generator/internal/domain/edu_plans"
	"schedule-generator/internal/domain/users"

	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/pkg/execerror"

	"github.com/google/uuid"
)

type EduGroupUsecaseRepo interface {
	edugroups.Repository
	eduplans.Repository

	db.TransactionalRepository
}

type EduGroupUsecase struct {
	repo    EduGroupUsecaseRepo
	authSvc *services.AuthorizationService
	logger  *slog.Logger
}

func NewEduGroupUsecase(authSvc *services.AuthorizationService, repo EduGroupUsecaseRepo, logger *slog.Logger) *EduGroupUsecase {
	return &EduGroupUsecase{
		authSvc: authSvc,
		repo:    repo,
		logger:  logger,
	}
}

type CreateEdugroupInput struct {
	Number    string
	EduPlanID uuid.UUID
}

type CreateEdugroupOutput struct {
	edugroups.EduGroup
}

// CreateEdugroup
func (uc *EduGroupUsecase) CreateEdugroup(ctx context.Context, input CreateEdugroupInput, user *users.User) (*CreateEdugroupOutput, error) {
	logger := uc.logger

	eduplan, err := uc.repo.GetEduPlan(ctx, input.EduPlanID)
	if err != nil {
		logger.Error("Get edu plan error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu plan not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if ok, err := uc.authSvc.HaveAccessToEduPlan(ctx, eduplan, user); err != nil {
		logger.Error("Check access to edu plan error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	} else if !ok {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have access to edu plan"))
	}

	group, err := edugroups.NewEduGroup(input.Number, eduplan)
	if err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveEduGroup(ctx, group)
	if err != nil {
		logger.Error("Save edu group error", "error", err)

		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu group with provided number already exists"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &CreateEdugroupOutput{
		EduGroup: *group,
	}, nil
}

type GetEduGroupOutput struct {
	edugroups.EduGroup
}

// GetEduGroup
func (uc *EduGroupUsecase) GetEduGroup(ctx context.Context, groupID uuid.UUID, user *users.User) (*GetEduGroupOutput, error) {
	logger := uc.logger

	group, err := uc.repo.GetEduGroup(ctx, groupID)
	if err != nil {
		logger.Error("List edu group error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("group not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if ok, err := uc.authSvc.HaveAccessToEduGroup(ctx, group, user); err != nil {
		logger.Error("Check access to group error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	} else if !ok {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have access to edu group"))
	}

	return &GetEduGroupOutput{
		EduGroup: *group,
	}, nil
}

// ListEduGroup
func (uc *EduGroupUsecase) ListEduGroup(ctx context.Context, user *users.User) ([]GetEduGroupOutput, error) {
	logger := uc.logger

	var groups []edugroups.EduGroup
	var listErr error

	if uc.authSvc.IsAdmin(user) {
		groups, listErr = uc.repo.ListEduGroup(ctx)
	} else {
		if user.FacultyID == nil {
			return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user not accociated with any faculty"))
		}

		groups, listErr = uc.repo.ListEduGroupByFacultyID(ctx, *user.FacultyID)
	}

	if listErr != nil {
		logger.Error("List edu group error", "error", listErr)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	result := make([]GetEduGroupOutput, len(groups))

	for i, group := range groups {
		result[i] = GetEduGroupOutput{
			EduGroup: group,
		}
	}

	return result, nil
}

type UpdateEduGroupInput struct {
	EduGroupID uuid.UUID
	Number     *string
}

type UpdateEduGroupOutput struct {
	edugroups.EduGroup
}

// UpdateEduGroup
func (uc *EduGroupUsecase) UpdateEduGroup(ctx context.Context, input UpdateEduGroupInput, user *users.User) (*UpdateEduGroupOutput, error) {
	logger := uc.logger

	group, err := uc.repo.GetEduGroup(ctx, input.EduGroupID)
	if err != nil {
		logger.Error("List edu group error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu group not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if ok, err := uc.authSvc.HaveAccessToEduGroup(ctx, group, user); err != nil {
		logger.Error("Check access to group error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	} else if !ok {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have access to edu group"))
	}

	if input.Number != nil {
		group.Number = *input.Number
	}

	if err := group.Validate(); err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveEduGroup(ctx, group)
	if err != nil {
		logger.Error("Save edu group error", "error", err)

		// TODO: use domain errors
		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu group with provided number already exists"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &UpdateEduGroupOutput{
		EduGroup: *group,
	}, nil
}

// DeleteEduGroup
func (uc *EduGroupUsecase) DeleteEduGroup(ctx context.Context, groupID uuid.UUID, user *users.User) error {
	logger := uc.logger

	group, err := uc.repo.GetEduGroup(ctx, groupID)
	if err != nil {
		logger.Error("List edu group error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu group not found"))
		}

		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if ok, err := uc.authSvc.HaveAccessToEduGroup(ctx, group, user); err != nil {
		logger.Error("Check access to group error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	} else if !ok {
		return execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have access to edu group"))
	}

	err = uc.repo.DeleteEduGroup(ctx, groupID)
	if err != nil {
		logger.Error("Delete edu group error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}
