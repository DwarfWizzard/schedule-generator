package handler

import (
	"context"
	"net/http"

	"schedule-generator/internal/application/usecases"
	eduplans "schedule-generator/internal/domain/edu_plans"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type EduPlanUsecase interface {
	CreateEduPlan(ctx context.Context, input usecases.CreateEduPlanInput) (*usecases.CreateEduPlanOutput, error)
	GetEduPlan(ctx context.Context, eduPlanID uuid.UUID) (*usecases.GetEduPlanOutput, error)
	ListEduPlan(ctx context.Context) ([]usecases.GetEduPlanOutput, error)
	DeleteEduPlan(ctx context.Context, eduPlanID uuid.UUID) error
}

type EduPlan struct {
	ID            uuid.UUID         `json:"id"`
	DirectionID   uuid.UUID         `json:"direction_id"`
	DirectionName string            `json:"direction_name"`
	Profile       string            `json:"profile"`
	Year          int64             `json:"year"`
	Modules       []eduplans.Module `json:"modules"`
}

type CreateEduPlanRequest struct {
	DirectionID uuid.UUID `json:"direction_id"`
	Profile     string    `json:"profile"`
	Year        int64     `json:"year"`
}

// CreateEduPlan - POST /v1/edu-plans
func (h *Handler) CreateEduPlan(c echo.Context) error {
	ctx := c.Request().Context()

	var rq CreateEduPlanRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	out, err := h.eduPlan.CreateEduPlan(ctx, usecases.CreateEduPlanInput{
		DirectionID: rq.DirectionID,
		Profile:     rq.Profile,
		Year:        rq.Year,
	})
	if err != nil {
		return err
	}

	return WrapResponse(http.StatusCreated, eduPlanToView(&out.EduPlan, out.DirectionName)).Send(c)
}

// GetEduPlan - GET /v1/edu-plans/:id
func (h *Handler) GetEduPlan(c echo.Context) error {
	ctx := c.Request().Context()

	eduPlanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	out, err := h.eduPlan.GetEduPlan(ctx, eduPlanID)
	if err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, eduPlanToView(&out.EduPlan, out.DirectionName)).Send(c)
}

// ListEduPlan - GET /v1/edu-plans
func (h *Handler) ListEduPlan(c echo.Context) error {
	ctx := c.Request().Context()

	out, err := h.eduPlan.ListEduPlan(ctx)
	if err != nil {
		return err
	}

	result := make([]EduPlan, len(out))
	for i, d := range out {
		result[i] = eduPlanToView(&d.EduPlan, d.DirectionName)
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}

// DeleteEduPlan - DELETE /v1/edu-plans/:id
func (h *Handler) DeleteEduPlan(c echo.Context) error {
	ctx := c.Request().Context()

	eduPlanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	if err := h.eduPlan.DeleteEduPlan(ctx, eduPlanID); err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

func eduPlanToView(model *eduplans.EduPlan, directionName string) EduPlan {
	return EduPlan{
		ID:            model.ID,
		DirectionID:   model.DirectionID,
		DirectionName: directionName,
		Profile:       model.Profile,
		Year:          model.Year,
		Modules:       model.Modules,
	}
}
