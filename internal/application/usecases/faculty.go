package usecases

import (
	"context"
	"errors"
	"log/slog"
	"schedule-generator/internal/application/services"
	"schedule-generator/internal/domain/faculties"
	"schedule-generator/internal/domain/users"
	"schedule-generator/pkg/execerror"
)

type FacultyUseCaseRepo interface {
	faculties.Repository
}

type FacultyUsecase struct {
	repo    FacultyUseCaseRepo
	authSvc *services.AuthorizationService
	logger  *slog.Logger
}

func NewFacultyUsecase(authSvc *services.AuthorizationService, repo FacultyUseCaseRepo, logger *slog.Logger) *FacultyUsecase {
	return &FacultyUsecase{
		authSvc: authSvc,
		repo:    repo,
		logger:  logger,
	}
}

type ListFacultyOutput = []faculties.Faculty

// ListFaculty
func (uc *FacultyUsecase) ListFaculty(ctx context.Context, user *users.User) (ListFacultyOutput, error) {
	logger := uc.logger

	var faculties []faculties.Faculty
	var listErr error

	if uc.authSvc.IsAdmin(user) {
		faculties, listErr = uc.repo.ListFaculty(ctx)
	} else {
		if user.FacultyID == nil {
			return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user not accociated with any faculty"))
		}

		faculty, err := uc.repo.GetFaculty(ctx, *user.FacultyID)

		if err != nil {
			listErr = err
		} else {
			faculties = append(faculties, *faculty)
		}
	}

	if listErr != nil {
		logger.Error("List faculty error", "error", listErr)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return faculties, nil
}
