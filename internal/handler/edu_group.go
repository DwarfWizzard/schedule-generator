package handler

import (
	"context"
	"net/http"

	"schedule-generator/internal/application/usecases"
	edugroups "schedule-generator/internal/domain/edu_groups"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type EduGroupUsecase interface {
	CreateEdugroup(ctx context.Context, input usecases.CreateEdugroupInput) (*usecases.CreateEdugroupOutput, error)
	GetEduGroup(ctx context.Context, groupID uuid.UUID) (*usecases.GetEduGroupOutput, error)
	ListEduGroup(ctx context.Context) ([]usecases.GetEduGroupOutput, error)
	UpdateEduGroup(ctx context.Context, input usecases.UpdateEduGroupInput) (*usecases.UpdateEduGroupOutput, error)
	DeleteEduGroup(ctx context.Context, groupID uuid.UUID) error
}

type EduGroup struct {
	ID            uuid.UUID `json:"id"`
	Number        string    `json:"number"`
	EduPlanID     uuid.UUID `json:"edu_plan_id"`
	Profile       string    `json:"profile"`
	AdmissionYear int64     `json:"admission_year"`
}

type CreateEduGroupRequest struct {
	Number    string    `json:"number"`
	EduPlanID uuid.UUID `json:"edu_plan_id"`
}

type UpdateEduGroupRequest struct {
	Number *string `json:"number"`
}

// CreateEduGroup - POST /v1/edu-groups
func (h *Handler) CreateEduGroup(c echo.Context) error {
	ctx := c.Request().Context()

	var rq CreateEduGroupRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	out, err := h.eduGroup.CreateEdugroup(ctx, usecases.CreateEdugroupInput{
		Number:    rq.Number,
		EduPlanID: rq.EduPlanID,
	})
	if err != nil {
		return err
	}

	return WrapResponse(http.StatusCreated, eduGroupToView(&out.EduGroup)).Send(c)
}

// GetEduGroup - GET /v1/edu-groups/:id
func (h *Handler) GetEduGroup(c echo.Context) error {
	ctx := c.Request().Context()

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	out, err := h.eduGroup.GetEduGroup(ctx, groupID)
	if err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, eduGroupToView(&out.EduGroup)).Send(c)
}
// ListEduGroup - GET /v1/edu-groups
func (h *Handler) ListEduGroup(c echo.Context) error {
	ctx := c.Request().Context()

	out, err := h.eduGroup.ListEduGroup(ctx)
	if err != nil {
		return err
	}

	result := make([]EduGroup, len(out))
	for i, d := range out {
		result[i] = eduGroupToView(&d.EduGroup)
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}

// UpdateEduGroup - PUT /v1/edu-groups/:id
func (h *Handler) UpdateEduGroup(c echo.Context) error {
	ctx := c.Request().Context()

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	var rq UpdateEduGroupRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	out, err := h.eduGroup.UpdateEduGroup(ctx, usecases.UpdateEduGroupInput{
		EduGroupID: groupID,
		Number:     rq.Number,
	})
	if err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, eduGroupToView(&out.EduGroup)).Send(c)
}

// DeleteEduGroup - DELETE /v1/edu-groups/:id
func (h *Handler) DeleteEduGroup(c echo.Context) error {
	ctx := c.Request().Context()

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	if err := h.eduGroup.DeleteEduGroup(ctx, groupID); err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

func eduGroupToView(model *edugroups.EduGroup) EduGroup {
	return EduGroup{
		ID:            model.ID,
		Number:        model.Number,
		EduPlanID:     model.EduPlanID,
		Profile:       model.Profile,
		AdmissionYear: model.AdmissionYear,
	}
}
