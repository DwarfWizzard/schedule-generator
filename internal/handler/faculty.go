package handler

import (
	"context"
	"net/http"

	"schedule-generator/internal/application/usecases"
	"schedule-generator/internal/domain/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FacultyUsecase interface {
	ListFaculty(ctx context.Context, user *users.User) (usecases.ListFacultyOutput, error)
}

type Faculty struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ListFaculty - GET /v1/faculties
func (h *Handler) ListFaculty(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	out, err := h.faculty.ListFaculty(ctx, user)
	if err != nil {
		return err
	}

	result := make([]Faculty, len(out))

	for i, faculty := range out {
		result[i] = Faculty{
			ID:   faculty.ID,
			Name: faculty.Name,
		}
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}
