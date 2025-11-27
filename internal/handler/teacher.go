package handler

import (
	"context"
	"net/http"

	"schedule-generator/internal/application/usecases"
	"schedule-generator/internal/domain/teachers"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TeacherUsecase interface {
	CreateTeacher(ctx context.Context, input usecases.CreateTeacherInput) (*usecases.CreateTeacherOutput, error)
	GetTeacher(ctx context.Context, teacherID uuid.UUID) (*usecases.GetTeacherOutput, error)
	ListTeacher(ctx context.Context) ([]usecases.GetTeacherOutput, error)
	UpdateTeacher(ctx context.Context, input usecases.UpdateTeacherInput) (*usecases.UpdateTeacherOutput, error)
	DeleteTeacher(ctx context.Context, teacherID uuid.UUID) error
}

type Teacher struct {
	ID           uuid.UUID `json:"id"`
	ExternalID   string    `json:"external_id"`
	Name         string    `json:"name"`
	Position     string    `json:"position"`
	Degree       string    `json:"degree"`
	DepartmentID uuid.UUID `json:"department_id"`
}

type CreateTeacherRequest struct {
	DepartmentID uuid.UUID `json:"department_id"`
	ExternalID   string    `json:"external_id"`
	Name         string    `json:"name"`
	Position     string    `json:"position"`
	Degree       string    `json:"degree"`
}

// CreateTeacher - POST /v1/teachers
func (h *Handler) CreateTeacher(c echo.Context) error {
	ctx := c.Request().Context()

	var rq CreateTeacherRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	out, err := h.teacher.CreateTeacher(ctx, usecases.CreateTeacherInput{
		DepartmentID: rq.DepartmentID,
		ExternalID:   rq.ExternalID,
		Name:         rq.Name,
		Position:     rq.Position,
		Degree:       rq.Degree,
	})
	if err != nil {
		return err
	}

	return WrapResponse(http.StatusCreated, teacherToView(&out.Teacher)).Send(c)
}

// GetTeacher - GET /v1/teachers/:id
func (h *Handler) GetTeacher(c echo.Context) error {
	ctx := c.Request().Context()

	teacherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	out, err := h.teacher.GetTeacher(ctx, teacherID)
	if err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, teacherToView(&out.Teacher)).Send(c)
}

// ListTeacher - GET /v1/teachers
func (h *Handler) ListTeacher(c echo.Context) error {
	ctx := c.Request().Context()

	out, err := h.teacher.ListTeacher(ctx)
	if err != nil {
		return err
	}

	result := make([]Teacher, len(out))
	for i, t := range out {
		result[i] = teacherToView(&t.Teacher)
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}

type UpdateTeacherRequest struct {
	ExternalID *string `json:"external_id"`
	Name       *string `json:"name"`
	Position   *string `json:"position"`
	Degree     *string `json:"degree"`
}

// UpdateTeacher - PUT /v1/teachers/:id
func (h *Handler) UpdateTeacher(c echo.Context) error {
	ctx := c.Request().Context()

	teacherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	var rq UpdateTeacherRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	out, err := h.teacher.UpdateTeacher(ctx, usecases.UpdateTeacherInput{
		TeacherID:  teacherID,
		ExternalID: rq.ExternalID,
		Name:       rq.Name,
		Position:   rq.Position,
		Degree:     rq.Degree,
	})
	if err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, teacherToView(&out.Teacher)).Send(c)
}

// DeleteTeacher - DELETE /v1/teachers/:id
func (h *Handler) DeleteTeacher(c echo.Context) error {
	ctx := c.Request().Context()

	teacherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	if err := h.teacher.DeleteTeacher(ctx, teacherID); err != nil {
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

func teacherToView(model *teachers.Teacher) Teacher {
	return Teacher{
		ID:           model.ID,
		ExternalID:   model.ExternalID,
		Name:         model.Name,
		Position:     model.Position,
		Degree:       model.Degree,
		DepartmentID: model.DepartmentID,
	}
}
