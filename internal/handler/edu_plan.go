package handler

import (
	"context"
	"net/http"

	"schedule-generator/internal/application/usecases"
	eduplans "schedule-generator/internal/domain/edu_plans"
	"schedule-generator/internal/domain/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type EduPlanUsecase interface {
	CreateEduPlan(ctx context.Context, input usecases.CreateEduPlanInput, user *users.User) (*usecases.CreateEduPlanOutput, error)
	GetEduPlan(ctx context.Context, eduPlanID uuid.UUID, user *users.User) (*usecases.GetEduPlanOutput, error)
	ListEduPlan(ctx context.Context, user *users.User) ([]usecases.GetEduPlanOutput, error)
	UpdateEduPlan(ctx context.Context, input usecases.UpdateEduPlanInput, user *users.User) (*usecases.UpdateEduPlanOutput, error)
	DeleteEduPlan(ctx context.Context, eduPlanID uuid.UUID, user *users.User) error
}

type EduPlan struct {
	ID             uuid.UUID         `json:"id"`
	DirectionID    uuid.UUID         `json:"direction_id"`
	DirectionName  string            `json:"direction_name"`
	DepartmentID   uuid.UUID         `json:"department_id"`
	DepartmentName string            `json:"department_name"`
	Profile        string            `json:"profile"`
	Year           int64             `json:"year"`
	Modules        []eduplans.Module `json:"modules"`
}

type CreateEduPlanRequest struct {
	DirectionID  uuid.UUID `json:"direction_id"`
	DepartmentID uuid.UUID `json:"department_id"`
	Profile      string    `json:"profile"`
	Year         int64     `json:"year"`
}

// CreateEduPlan - POST /v1/edu-plans
func (h *Handler) CreateEduPlan(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	var rq CreateEduPlanRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	out, err := h.eduPlan.CreateEduPlan(ctx, usecases.CreateEduPlanInput{
		DirectionID:  rq.DirectionID,
		DepartmentID: rq.DepartmentID,
		Profile:      rq.Profile,
		Year:         rq.Year,
	}, user)
	if err != nil {
		return err
	}

	return WrapResponse(http.StatusCreated, eduPlanToView(&out.EduPlan, out.DirectionName, out.DepartmentName)).Send(c)
}

// GetEduPlan - GET /v1/edu-plans/:id
func (h *Handler) GetEduPlan(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	eduPlanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	out, err := h.eduPlan.GetEduPlan(ctx, eduPlanID, user)
	if err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, eduPlanToView(&out.EduPlan, out.DirectionName, out.DepartmentName)).Send(c)
}

// ListEduPlan - GET /v1/edu-plans
func (h *Handler) ListEduPlan(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	out, err := h.eduPlan.ListEduPlan(ctx, user)
	if err != nil {
		return err
	}

	result := make([]EduPlan, len(out))
	for i, d := range out {
		result[i] = eduPlanToView(&d.EduPlan, d.DirectionName, d.DepartmentName)
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}

type UpdateEduPlanRequest struct {
	ID           uuid.UUID  `json:"-" param:"id"`
	DirectionID  *uuid.UUID `json:"direction_id" param:"-"`
	DepartmentID *uuid.UUID `json:"department_id" param:"-"`
	Profile      *string    `json:"profile" param:"-"`
	Year         *int64     `json:"year" param:"-"`
}

// UpdateEduPlan - PATCH /v1/edu-plans/:id
func (h *Handler) UpdateEduPlan(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	var rq UpdateEduPlanRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	out, err := h.eduPlan.UpdateEduPlan(ctx, usecases.UpdateEduPlanInput{
		ID:           rq.ID,
		DirectionID:  rq.DirectionID,
		DepartmentID: rq.DepartmentID,
		Profile:      rq.Profile,
		Year:         rq.Year,
	}, user)
	if err != nil {
		h.logger.Error("Update edu plan error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, eduPlanToView(&out.EduPlan, out.DirectionName, out.DepartmentName)).Send(c)
}

// DeleteEduPlan - DELETE /v1/edu-plans/:id
func (h *Handler) DeleteEduPlan(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	eduPlanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	if err := h.eduPlan.DeleteEduPlan(ctx, eduPlanID, user); err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

func eduPlanToView(model *eduplans.EduPlan, directionName, departmentName string) EduPlan {
	return EduPlan{
		ID:             model.ID,
		DirectionID:    model.DirectionID,
		DirectionName:  directionName,
		DepartmentID:   model.DepartmentID,
		DepartmentName: departmentName,
		Profile:        model.Profile,
		Year:           model.Year,
		Modules:        model.Modules,
	}
}
