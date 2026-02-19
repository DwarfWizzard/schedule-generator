package handler

import (
	"context"
	"net/http"

	"schedule-generator/internal/application/usecases"
	"schedule-generator/internal/domain/departments"
	"schedule-generator/internal/domain/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type DepartmentUsecase interface {
	CreateDepartment(ctx context.Context, input usecases.CreateDepartmentInput, user *users.User) (*usecases.CreateDepartmentOutput, error)
	GetDepartment(ctx context.Context, departmentID uuid.UUID, user *users.User) (*usecases.GetDepartmentOutput, error)
	ListDepartment(ctx context.Context, user *users.User) (usecases.ListDepartmentOutput, error)
	UpdateDepartment(ctx context.Context, input usecases.UpdateDepartmentInput, user *users.User) (*usecases.UpdateDepartmentOutput, error)
	DeleteDepartment(ctx context.Context, departmentID uuid.UUID, user *users.User) error
}

type Department struct {
	ID          uuid.UUID `json:"id"`
	ExternalID  string    `json:"external_id"`
	FacultyID   uuid.UUID `json:"faculty_id"`
	FacultyName string    `json:"faculty_name"`
	Name        string    `json:"name"`
}

type CreateDepartmentRequest struct {
	FacultyID  uuid.UUID `json:"faculty_id"`
	ExternalID string    `json:"external_id"`
	Name       string    `json:"name"`
}

// CreateDepartment - POST /v1/departments
func (h *Handler) CreateDepartment(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	var rq CreateDepartmentRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	out, err := h.department.CreateDepartment(ctx, usecases.CreateDepartmentInput{
		FacultyID:  rq.FacultyID,
		ExternalID: rq.ExternalID,
		Name:       rq.Name,
	}, user)
	if err != nil {
		h.logger.Error("Create department error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, departmentToView(&out.Department, out.FacultyName)).Send(c)
}

// GetDepartment - GET /v1/departments/:id
func (h *Handler) GetDepartment(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	departmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	out, err := h.department.GetDepartment(ctx, departmentID, user)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, departmentToView(&out.Department, out.FacultyName)).Send(c)
}

// ListDepartment - GET /v1/departments
func (h *Handler) ListDepartment(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	out, err := h.department.ListDepartment(ctx, user)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	result := make([]Department, len(out))
	for i, d := range out {
		result[i] = departmentToView(&d.Department, d.FacultyName)
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}

type UpdateDepartmentRequest struct {
	ExternalID *string `json:"external_id"`
	Name       *string `json:"name"`
}

// UpdateDepartment - PUT /v1/departments/:id
func (h *Handler) UpdateDepartment(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	departmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	var rq UpdateDepartmentRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	out, err := h.department.UpdateDepartment(ctx, usecases.UpdateDepartmentInput{
		DepartmentID: departmentID,
		ExternalID:   rq.ExternalID,
		Name:         rq.Name,
	}, user)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, departmentToView(&out.Department, out.FacultyName)).Send(c)
}

// DeleteDepartment - DELETE /v1/departments/:id
func (h *Handler) DeleteDepartment(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	departmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	err = h.department.DeleteDepartment(ctx, departmentID, user)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

func departmentToView(model *departments.Department, facultyName string) Department {
	return Department{
		ID:          model.ID,
		ExternalID:  model.ExternalID,
		FacultyID:   model.FacultyID,
		FacultyName: facultyName,
		Name:        model.Name,
	}
}
