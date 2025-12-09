package handler

import (
	"context"
	"net/http"

	"schedule-generator/internal/application/usecases"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FacultyUsecase interface {
	ListFaculty(ctx context.Context) (usecases.ListFacultyOutput, error)
}

type Faculty struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ListFaculty - GET /v1/faculties
func (h *Handler) ListFaculty(c echo.Context) error {
	ctx := c.Request().Context()

	out, err := h.faculty.ListFaculty(ctx)
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
