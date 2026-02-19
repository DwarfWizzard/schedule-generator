package handler

import (
	"context"
	"net/http"

	"schedule-generator/internal/application/usecases"
	edudirections "schedule-generator/internal/domain/edu_directions"
	"schedule-generator/internal/domain/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type EduDirectionUsecase interface {
	CreateEduDirection(ctx context.Context, input usecases.CreateEduDirectionInput, user *users.User) (*usecases.CreateEduDirectionOutput, error)
	GetEduDirection(ctx context.Context, directionID uuid.UUID, user *users.User) (*usecases.GetEduDirectionOutput, error)
	ListEduDirection(ctx context.Context, user *users.User) ([]usecases.GetEduDirectionOutput, error)
	UpdateEduDirection(ctx context.Context, input usecases.UpdateEduDirectionInput, user *users.User) (*usecases.UpdateEduDirectionOutput, error)
	DeleteEduDirection(ctx context.Context, directionID uuid.UUID, user *users.User) error
}

type EduDirection struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	DepartmentID   uuid.UUID `json:"department_id"`
	DepartmentName string    `json:"department_name"`
}

type CreateEduDirectionRequest struct {
	Name         string    `json:"name"`
	DepartmentID uuid.UUID `json:"department_id"`
}

type UpdateEduDirectionRequest struct {
	Name *string `json:"name"`
}

func (h *Handler) CreateEduDirection(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	var rq CreateEduDirectionRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	out, err := h.eduDirection.CreateEduDirection(ctx, usecases.CreateEduDirectionInput{
		Name:         rq.Name,
		DepartmentID: rq.DepartmentID,
	}, user)
	if err != nil {
		h.logger.Error("Create edu direction error", "error", err)
		return err
	}

	return WrapResponse(http.StatusCreated, eduDirectionToView(&out.EduDirection, out.DepartmentName)).Send(c)
}

func (h *Handler) GetEduDirection(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	directionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	out, err := h.eduDirection.GetEduDirection(ctx, directionID, user)
	if err != nil {
		h.logger.Error("Get edu direction error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, eduDirectionToView(&out.EduDirection, out.DepartmentName)).Send(c)
}

func (h *Handler) ListEduDirection(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	out, err := h.eduDirection.ListEduDirection(ctx, user)
	if err != nil {
		h.logger.Error("List edu direction error", "error", err)
		return err
	}

	result := make([]EduDirection, len(out))
	for i, d := range out {
		result[i] = eduDirectionToView(&d.EduDirection, d.DepartmentName)
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}

func (h *Handler) UpdateEduDirection(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	directionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	var rq UpdateEduDirectionRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	out, err := h.eduDirection.UpdateEduDirection(ctx, usecases.UpdateEduDirectionInput{
		EduDirectionID: directionID,
		Name:           rq.Name,
	}, user)
	if err != nil {
		h.logger.Error("Update edu direction error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, eduDirectionToView(&out.EduDirection, out.DepartmentName)).Send(c)
}

func (h *Handler) DeleteEduDirection(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	directionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	if err := h.eduDirection.DeleteEduDirection(ctx, directionID, user); err != nil {
		h.logger.Error("Delete edu direction error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

func eduDirectionToView(model *edudirections.EduDirection) EduDirection {
	return EduDirection{
		ID:   model.ID,
		Name: model.Name,
	}
}
