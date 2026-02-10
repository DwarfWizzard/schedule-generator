package usecases

import (
	"context"
	"errors"
	"log/slog"

	"schedule-generator/internal/application/services"
	"schedule-generator/internal/domain/faculties"
	"schedule-generator/internal/domain/users"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/pkg/execerror"

	"github.com/google/uuid"
)

type UserUsecaseRepository interface {
	users.Repository
	faculties.Repository
}

type UserUsecase struct {
	authSvc  *services.AuthorizationService
	pwdSvc   services.PasswordService
	tokenSvc services.TokenService
	repo     UserUsecaseRepository
	logger   *slog.Logger
}

type CreateUserInput struct {
	Username  string
	Password  string
	Role      int8
	FacultyID *uuid.UUID
}

func NewUserUsecase(authSvc *services.AuthorizationService, pwdSvc services.PasswordService, tokenSvc services.TokenService, repo UserUsecaseRepository, logger *slog.Logger) *UserUsecase {
	return &UserUsecase{
		repo:     repo,
		authSvc:  authSvc,
		pwdSvc:   pwdSvc,
		tokenSvc: tokenSvc,
		logger:   logger,
	}
}

// CreateUser
func (uc *UserUsecase) CreateUser(ctx context.Context, input CreateUserInput, user *users.User) (*users.User, error) {
	logger := uc.logger

	if !uc.authSvc.IsAdmin(user) {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have acces to usecase"))
	}

	if input.FacultyID != nil {
		_, err := uc.repo.GetFaculty(ctx, *input.FacultyID)
		if err != nil {
			logger.Error("Get faculty error", "error", err)
			if errors.Is(err, db.ErrorNotFound) {
				return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("faculty not found"))
			}

			return nil, execerror.NewExecError(execerror.TypeInternal, nil)
		}
	}

	role, err := users.NewRole(input.Role)
	if err != nil {
		logger.Error("Invalid role", "role", role.String())
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	pwdHash, err := uc.pwdSvc.HashPassword(user.PwdHash)
	if err != nil {
		logger.Error("Hash password error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("password not allowed"))
	}

	u, err := users.NewUser(input.Username, role, input.FacultyID, pwdHash)
	if err != nil {
		logger.Error("Invalid user data", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveUser(ctx, u)
	if err != nil {
		logger.Error("Save user error", "error", err)
		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("user with provied username already exists"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return u, nil
}

type UserAuthenticationInput struct {
	Username string
	Password string
}

// UserAuthentication
func (uc *UserUsecase) UserAuthentication(ctx context.Context, input UserAuthenticationInput) (services.TokenPair, error) {
	logger := uc.logger

	user, err := uc.repo.GetUserByUsername(ctx, input.Username)
	if err != nil {
		logger.Error("Get user by username error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return services.TokenPair{}, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("user not found"))
		}

		return services.TokenPair{}, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if !uc.pwdSvc.CompareWithHash(user.PwdHash, input.Password) {
		return services.TokenPair{}, execerror.NewExecError(execerror.TypeForbbiden, errors.New("password mismatch"))
	}

	pair, err := uc.tokenSvc.GenerateToken(ctx, &services.TokenClaims{
		UserID: user.ID,
		Role:   user.Role,
	})
	if err != nil {
		logger.Error("Generate token error", "error", err)
		return services.TokenPair{}, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return pair, nil
}

// UserAuthorization
func (uc *UserUsecase) UserAuthorization(ctx context.Context, token string) (*users.User, error) {
	logger := uc.logger

	claims, err := uc.tokenSvc.ParseAccessToken(ctx, token)
	if err != nil {
		logger.Error("Parse token error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("invalid token"))
	}

	user, err := uc.repo.GetUser(ctx, claims.UserID)
	if err != nil {
		logger.Error("Get user error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return user, nil
}

// RefreshUserToken
func (uc *UserUsecase) RefreshUserToken(ctx context.Context, refresh string) (services.TokenPair, error) {
	logger := uc.logger

	claims, err := uc.tokenSvc.ParseRefreshToken(ctx, refresh)
	if err != nil {
		logger.Error("Parse refresh token error", "error", err)
		return services.TokenPair{}, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("invalid token"))
	}

	pair, err := uc.tokenSvc.GenerateToken(ctx, &claims)
	if err != nil {
		logger.Error("Generate token pair error", "error", err)
		return services.TokenPair{}, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return pair, nil
}

// GetUser
func (uc *UserUsecase) GetUser(ctx context.Context, userID uuid.UUID) (*users.User, error) {
	logger := uc.logger

	user, err := uc.repo.GetUser(ctx, userID)
	if err != nil {
		logger.Error("Get user error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("user not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return user, nil
}

// ListUser
func (uc *UserUsecase) ListUser(ctx context.Context, user *users.User) ([]users.User, error) {
	logger := uc.logger

	if !uc.authSvc.IsAdmin(user) {
		return nil, execerror.NewExecError(execerror.TypeForbbiden, errors.New("user does not have acces to usecase"))
	}

	users, err := uc.repo.ListUser(ctx)
	if err != nil {
		logger.Error("List user error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return users, nil
}
