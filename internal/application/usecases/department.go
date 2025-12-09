package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"schedule-generator/internal/domain/departments"
	"schedule-generator/internal/domain/faculties"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/pkg/execerror"

	"github.com/google/uuid"
)

type DepartmentUsecaseRepo interface {
	departments.Repository
	faculties.Repository

	MapFacultiesByDepartments(ctx context.Context, departmentIDs uuid.UUIDs) (map[uuid.UUID]faculties.Faculty, error)
}

type DepartmentUsecase struct {
	repo   DepartmentUsecaseRepo
	logger *slog.Logger
}

func NewDepartmentUsecase(repo DepartmentUsecaseRepo, logger *slog.Logger) *DepartmentUsecase {
	return &DepartmentUsecase{
		repo:   repo,
		logger: logger,
	}
}

type CreateDepartmentInput struct {
	FacultyID  uuid.UUID
	ExternalID string
	Name       string
}

type CreateDepartmentOutput struct {
	departments.Department
	FacultyName string
}

// CreateDepartment
func (uc *DepartmentUsecase) CreateDepartment(ctx context.Context, input CreateDepartmentInput) (*CreateDepartmentOutput, error) {
	logger := uc.logger

	faculty, err := uc.repo.GetFaculty(ctx, input.FacultyID)
	if err != nil {
		logger.Error("Get faculty error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("faculty not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	department, err := departments.NewDepartment(faculty.ID, input.ExternalID, input.Name)
	if err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveDepartment(ctx, department)
	if err != nil {
		logger.Error("Save edu department error", "error", err)

		// TODO: use domain errors
		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("department with provided external id already exists"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &CreateDepartmentOutput{
		Department:  *department,
		FacultyName: faculty.Name,
	}, nil
}

type GetDepartmentOutput struct {
	departments.Department
	FacultyName string
}

// GetDepartment
func (uc *DepartmentUsecase) GetDepartment(ctx context.Context, departmentID uuid.UUID) (*GetDepartmentOutput, error) {
	logger := uc.logger

	department, err := uc.repo.GetDepartment(ctx, departmentID)
	if err != nil {
		logger.Error("List department error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("department not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	faculty, err := uc.repo.GetFaculty(ctx, department.FacultyID)
	if err != nil {
		logger.Error("Get faculty error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &GetDepartmentOutput{
		Department:  *department,
		FacultyName: faculty.Name,
	}, nil
}

type ListDepartmentOutput = []GetDepartmentOutput

// ListDepartment
func (uc *DepartmentUsecase) ListDepartment(ctx context.Context) (ListDepartmentOutput, error) {
	logger := uc.logger

	departments, err := uc.repo.ListDepartment(ctx)
	if err != nil {
		logger.Error("List department error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	departmentIDs := make(uuid.UUIDs, len(departments))

	for i, dep := range departments {
		departmentIDs[i] = dep.ID
	}

	faculties, err := uc.repo.MapFacultiesByDepartments(ctx, departmentIDs)
	if err != nil {
		logger.Error("Map faculties by departments error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	result := make(ListDepartmentOutput, len(departments))
	for i, dep := range departments {
		faculty, ok := faculties[dep.FacultyID]
		if !ok {
			logger.Error(fmt.Sprintf("Faculty for department %s not found", dep.ID))
			return nil, execerror.NewExecError(execerror.TypeInternal, nil)
		}

		result[i] = GetDepartmentOutput{
			Department:  dep,
			FacultyName: faculty.Name,
		}
	}

	return result, nil
}

type UpdateDepartmentInput struct {
	DepartmentID uuid.UUID
	ExternalID   *string
	Name         *string
}

type UpdateDepartmentOutput struct {
	departments.Department
	FacultyName string
}

// UpdateDepartment
func (uc *DepartmentUsecase) UpdateDepartment(ctx context.Context, input UpdateDepartmentInput) (*UpdateDepartmentOutput, error) {
	logger := uc.logger

	department, err := uc.repo.GetDepartment(ctx, input.DepartmentID)
	if err != nil {
		logger.Error("List department error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("department not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	faculty, err := uc.repo.GetFaculty(ctx, department.FacultyID)
	if err != nil {
		logger.Error("Get faculty error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if input.ExternalID != nil {
		department.ExternalID = *input.ExternalID
	}

	if input.Name != nil {
		department.Name = *input.Name
	}

	if err := department.Validate(); err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveDepartment(ctx, department)
	if err != nil {
		logger.Error("Save edu department error", "error", err)

		// TODO: use domain errors
		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("department with provided external id already exists"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &UpdateDepartmentOutput{
		Department:  *department,
		FacultyName: faculty.Name,
	}, nil
}

// DeleteDepartment
func (uc *DepartmentUsecase) DeleteDepartment(ctx context.Context, departmentID uuid.UUID) error {
	logger := uc.logger

	err := uc.repo.DeleteDepartment(ctx, departmentID)
	if err != nil {
		logger.Error("Delete edu department error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}
