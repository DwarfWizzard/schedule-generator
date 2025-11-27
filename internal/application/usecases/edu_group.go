package usecases

import (
	"context"
	"errors"
	"log/slog"

	edugroups "schedule-generator/internal/domain/edu_groups"
	eduplans "schedule-generator/internal/domain/edu_plans"

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
	repo   EduGroupUsecaseRepo
	logger *slog.Logger
}

func NewEduGroupUsecase(repo EduGroupUsecaseRepo, logger *slog.Logger) *EduGroupUsecase {
	return &EduGroupUsecase{
		repo:   repo,
		logger: logger,
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
func (uc *EduGroupUsecase) CreateEdugroup(ctx context.Context, input CreateEdugroupInput) (*CreateEdugroupOutput, error) {
	logger := uc.logger

	eduplan, err := uc.repo.GetEduPlan(ctx, input.EduPlanID)
	if err != nil {
		logger.Error("Get edu plan error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu plan not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
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
func (uc *EduGroupUsecase) GetEduGroup(ctx context.Context, groupID uuid.UUID) (*GetEduGroupOutput, error) {
	logger := uc.logger

	group, err := uc.repo.GetEduGroup(ctx, groupID)
	if err != nil {
		logger.Error("List edu group error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("group not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &GetEduGroupOutput{
		EduGroup: *group,
	}, nil
}

// ListEduGroup
func (uc *EduGroupUsecase) ListEduGroup(ctx context.Context) ([]GetEduGroupOutput, error) {
	logger := uc.logger

	groups, err := uc.repo.ListEduGroup(ctx)
	if err != nil {
		logger.Error("List edu group error", "error", err)
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
func (uc *EduGroupUsecase) UpdateEduGroup(ctx context.Context, input UpdateEduGroupInput) (*UpdateEduGroupOutput, error) {
	logger := uc.logger

	group, err := uc.repo.GetEduGroup(ctx, input.EduGroupID)
	if err != nil {
		logger.Error("List edu group error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("edu group not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
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
func (uc *EduGroupUsecase) DeleteEduGroup(ctx context.Context, groupID uuid.UUID) error {
	logger := uc.logger

	err := uc.repo.DeleteEduGroup(ctx, groupID)
	if err != nil {
		logger.Error("Delete edu group error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}
