package usecases

import (
	"context"
	"log/slog"
	"schedule-generator/internal/domain/faculties"
	"schedule-generator/pkg/execerror"
)

type FacultyUseCaseRepo interface {
	faculties.Repository
}

type FacultyUsecase struct {
	repo   FacultyUseCaseRepo
	logger *slog.Logger
}

func NewFacultyUsecase(repo FacultyUseCaseRepo, logger *slog.Logger) *FacultyUsecase {
	return &FacultyUsecase{
		repo:   repo,
		logger: logger,
	}
}

type ListFacultyOutput = []faculties.Faculty

// ListFaculty
func (uc *FacultyUsecase) ListFaculty(ctx context.Context) (ListFacultyOutput, error) {
	logger := uc.logger

	faculties, err := uc.repo.ListFaculty(ctx)
	if err != nil {
		logger.Error("List faculty error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return faculties, nil
}
