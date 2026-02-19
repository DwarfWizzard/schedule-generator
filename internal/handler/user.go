package handler

import (
	"context"
	"net/http"

	"schedule-generator/internal/application/services"
	"schedule-generator/internal/application/usecases"
	"schedule-generator/internal/domain/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserUsecase interface {
	CreateUser(ctx context.Context, input usecases.CreateUserInput, user *users.User) (*usecases.CreateUserOutput, error)
	UserAuthentication(ctx context.Context, input usecases.UserAuthenticationInput) (services.TokenPair, error)
	UserAuthorization(ctx context.Context, token string) (*users.User, error)
	RefreshUserToken(ctx context.Context, refresh string) (services.TokenPair, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*usecases.GetUserOutput, error)
	ListUser(ctx context.Context, user *users.User) (usecases.ListUserOutput, error)
}

type User struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Username    string     `json:"username"`
	Role        int8       `json:"role"`
	FacultyID   *uuid.UUID `json:"faculty_id,omitempty"`
	FacultyName *string    `json:"faculty_name"`
}

type CreateUserRequest struct {
	Name      string     `json:"name"`
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	Role      int8       `json:"role"`
	FacultyID *uuid.UUID `json:"faculty_id"`
}

// CreateUser - POST /v1/users
func (h *Handler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	var rq CreateUserRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	out, err := h.user.CreateUser(ctx, usecases.CreateUserInput{
		Name:      rq.Name,
		Username:  rq.Username,
		Password:  rq.Password,
		Role:      rq.Role,
		FacultyID: rq.FacultyID,
	}, user)
	if err != nil {
		h.logger.Error("Create user error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, userToView(&out.User, out.FacultyName)).Send(c)
}

type GetUserRequest struct {
	UserID uuid.UUID `param:"user_id"`
}

// GetUser - GET /v1/users/:id
func (h *Handler) GetUser(c echo.Context) error {
	ctx := c.Request().Context()

	var rq GetUserRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	out, err := h.user.GetUser(ctx, rq.UserID)
	if err != nil {
		h.logger.Error("Get user error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, userToView(&out.User, out.FacultyName)).Send(c)
}

// ListUser - GET /v1/users
func (h *Handler) ListUser(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	list, err := h.user.ListUser(ctx, user)
	if err != nil {
		h.logger.Error("List user error", "error", err)
		return err
	}

	result := make([]*User, len(list))
	for i, u := range list {
		result[i] = userToView(&u.User, u.FacultyName)
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}

func userToView(user *users.User, facultyName *string) *User {
	return &User{
		ID:          user.ID,
		Name:        user.Name,
		FacultyID:   user.FacultyID,
		FacultyName: facultyName,
		Username:    user.Username,
		Role:        int8(user.Role),
	}
}
