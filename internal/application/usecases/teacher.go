package usecases

import (
	"context"
	"errors"
	"log/slog"
	"schedule-generator/internal/domain/departments"
	"schedule-generator/internal/domain/teachers"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/pkg/execerror"

	"github.com/google/uuid"
)

type TeacherUsecaseRepo interface {
	departments.Repository
	teachers.Repository

	ListTeacherByDepartment(ctx context.Context, depID string) ([]teachers.Teacher, error)
}

type TeacherUsecase struct {
	repo   TeacherUsecaseRepo
	logger *slog.Logger
}

func NewTeacherUsecase(repo TeacherUsecaseRepo, logger *slog.Logger) *TeacherUsecase {
	return &TeacherUsecase{
		repo:   repo,
		logger: logger,
	}
}

type CreateTeacherInput struct {
	DepartmentID uuid.UUID
	ExternalID   string
	Name         string
	Position     string
	Degree       string
}

type CreateTeacherOutput struct {
	teachers.Teacher
}

// CreateTeacher
func (uc *TeacherUsecase) CreateTeacher(ctx context.Context, input CreateTeacherInput) (*CreateTeacherOutput, error) {
	logger := uc.logger

	department, err := uc.repo.GetDepartment(ctx, input.DepartmentID)
	if err != nil {
		logger.Error("Get department error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("department not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	teacher, err := teachers.NewTeacher(department.ID, input.ExternalID, input.Name, input.Position, input.Degree)
	if err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveTeacher(ctx, teacher)
	if err != nil {
		logger.Error("Save teacher error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &CreateTeacherOutput{
		Teacher: *teacher,
	}, nil
}

type GetTeacherOutput struct {
	teachers.Teacher
}

// GetTeacher
func (uc *TeacherUsecase) GetTeacher(ctx context.Context, teacherID uuid.UUID) (*GetTeacherOutput, error) {
	logger := uc.logger

	teacher, err := uc.repo.GetTeacher(ctx, teacherID)
	if err != nil {
		logger.Error("Get teacher error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("teacher not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &GetTeacherOutput{
		Teacher: *teacher,
	}, nil
}

// ListTeacher
func (uc *TeacherUsecase) ListTeacher(ctx context.Context) ([]GetTeacherOutput, error) {
	logger := uc.logger

	teachers, err := uc.repo.ListTeacher(ctx)
	if err != nil {
		logger.Error("List teacher error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	result := make([]GetTeacherOutput, len(teachers))

	for i, teacher := range teachers {
		result[i] = GetTeacherOutput{
			Teacher: teacher,
		}
	}

	return result, nil
}

type UpdateTeacherInput struct {
	TeacherID  uuid.UUID
	ExternalID *string
	Name       *string
	Position   *string
	Degree     *string
}

type UpdateTeacherOutput struct {
	teachers.Teacher
}

// UpdateTeacher
func (uc *TeacherUsecase) UpdateTeacher(ctx context.Context, input UpdateTeacherInput) (*UpdateTeacherOutput, error) {
	logger := uc.logger

	teacher, err := uc.repo.GetTeacher(ctx, input.TeacherID)
	if err != nil {
		logger.Error("List teacher error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("teacher not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if input.ExternalID != nil {
		teacher.ExternalID = *input.ExternalID
	}

	if input.Name != nil {
		teacher.Name = *input.Name
	}

	if input.Position != nil {
		teacher.Position = *input.Position
	}

	if input.Degree != nil {
		teacher.Position = *input.Degree
	}

	if err := teacher.Validate(); err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveTeacher(ctx, teacher)
	if err != nil {
		logger.Error("Save edu teacher error", "error", err)

		// TODO: use domain errors
		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("teacher with provided external id already exists"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &UpdateTeacherOutput{
		Teacher: *teacher,
	}, nil
}

// DeleteTeacher
func (uc *TeacherUsecase) DeleteTeacher(ctx context.Context, teacherID uuid.UUID) error {
	logger := uc.logger

	err := uc.repo.DeleteTeacher(ctx, teacherID)
	if err != nil {
		logger.Error("Delete teacher error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}
