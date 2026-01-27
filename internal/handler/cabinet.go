package handler

import (
	"context"
	"net/http"

	"schedule-generator/internal/application/usecases"
	"schedule-generator/internal/domain/cabinets"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CabinetUsecase interface {
	CreateCabinet(ctx context.Context, input usecases.CreateCabinetInput) (*usecases.CreateCabinetOutput, error)
	GetCabinet(ctx context.Context, cabinetID uuid.UUID) (*usecases.GetCabinetOutput, error)
	ListCabinet(ctx context.Context) (usecases.ListCabinetOutput, error)
	UpdateCabinet(ctx context.Context, input usecases.UpdateCabinetInput) (*usecases.UpdateCabinetOutput, error)
	DeleteCabinet(ctx context.Context, cabinetID uuid.UUID) error
}

type CabinetEquipment struct {
	Furniture         string `json:"furniture"`
	TechnicalMeans    string `json:"technical_means"`
	ComputerEquipment string `json:"computer_equipment"`
}

type Cabinet struct {
	ID                                 uuid.UUID         `json:"id"`
	FacultyID                          uuid.UUID         `json:"faculty_id"`
	FacultyName                        string            `json:"faculty_name"`
	Type                               int8              `json:"type"`
	Building                           string            `json:"building"`
	Auditorium                         string            `json:"auditorium"`
	SuitableForPeoplesWithSpecialNeeds bool              `json:"suitable_for_peoples_with_special_needs"`
	Appointment                        *string           `json:"appointment"`
	Equipment                          *CabinetEquipment `json:"equipment"`
}

type CreateCabinetRequest struct {
	FacultyID                          uuid.UUID         `json:"faculty_id"`
	CabinetType                        int8              `json:"cabinet_type"`
	Building                           string            `json:"building"`
	Auditorium                         string            `json:"auditorium"`
	SuitableForPeoplesWithSpecialNeeds bool              `json:"suitable_for_peoples_with_special_needs"`
	Appointment                        *string           `json:"appointment"`
	Equipment                          *CabinetEquipment `json:"equipment"`
}

// CreateCabinet - POST /v1/cabinets
func (h *Handler) CreateCabinet(c echo.Context) error {
	ctx := c.Request().Context()

	var rq CreateCabinetRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	var equipment *usecases.Equipment
	if rq.Equipment != nil {
		equipment = &usecases.Equipment{
			Furniture:         rq.Equipment.Furniture,
			TechnicalMeans:    rq.Equipment.TechnicalMeans,
			СomputerEquipment: rq.Equipment.ComputerEquipment,
		}
	}

	out, err := h.cabinet.CreateCabinet(ctx, usecases.CreateCabinetInput{
		FacultyID:                          rq.FacultyID,
		CabinetType:                        rq.CabinetType,
		Auditorium:                         rq.Auditorium,
		SuitableForPeoplesWithSpecialNeeds: rq.SuitableForPeoplesWithSpecialNeeds,
		Building:                           rq.Building,
		Appointment:                        rq.Appointment,
		Equipment:                          equipment,
	})
	if err != nil {
		h.logger.Error("Create cabinet error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, cabinetToView(&out.Cabinet, out.FacultyName)).Send(c)
}

// GetCabinet - GET /v1/cabinets/:id
func (h *Handler) GetCabinet(c echo.Context) error {
	ctx := c.Request().Context()

	cabinetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	out, err := h.cabinet.GetCabinet(ctx, cabinetID)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, cabinetToView(&out.Cabinet, out.FacultyName)).Send(c)
}

// ListCabinet - GET /v1/cabinets
func (h *Handler) ListCabinet(c echo.Context) error {
	ctx := c.Request().Context()

	out, err := h.cabinet.ListCabinet(ctx)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	result := make([]Cabinet, len(out))
	for i, d := range out {
		result[i] = cabinetToView(&d.Cabinet, d.FacultyName)
	}

	return WrapResponse(http.StatusOK, result).Send(c)
}

type UpdateCabinetRequest struct {
	FacultyID                          *uuid.UUID `json:"faculty_id"`
	CabinetType                        *int8
	Type                               *int8             `json:"type"`
	Building                           *string           `json:"building"`
	Auditorium                         *string           `json:"auditorium"`
	SuitableForPeoplesWithSpecialNeeds *bool             `json:"suitable_for_peoples_with_special_needs"`
	Appointment                        *string           `json:"appointment"`
	Equipment                          *CabinetEquipment `json:"equipment"`
}

// UpdateCabinet - PUT /v1/cabinets/:id
func (h *Handler) UpdateCabinet(c echo.Context) error {
	ctx := c.Request().Context()

	cabinetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	var rq UpdateCabinetRequest
	if err := c.Bind(&rq); err != nil {
		h.logger.Error("Parse request error", "error", err)
		return ErrNotParsable
	}

	var equipment *usecases.Equipment
	if rq.Equipment != nil {
		equipment = &usecases.Equipment{
			Furniture:         rq.Equipment.Furniture,
			TechnicalMeans:    rq.Equipment.TechnicalMeans,
			СomputerEquipment: rq.Equipment.ComputerEquipment,
		}
	}

	out, err := h.cabinet.UpdateCabinet(ctx, usecases.UpdateCabinetInput{
		CabinetID:                          cabinetID,
		FacultyID:                          rq.FacultyID,
		CabinetType:                        rq.CabinetType,
		Building:                           rq.Building,
		Auditorium:                         rq.Auditorium,
		SuitableForPeoplesWithSpecialNeeds: rq.SuitableForPeoplesWithSpecialNeeds,
		Appointment:                        rq.Appointment,
		Equipment:                          equipment,
	})
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, cabinetToView(&out.Cabinet, out.FacultyName)).Send(c)
}

// DeleteCabinet - DELETE /v1/cabinets/:id
func (h *Handler) DeleteCabinet(c echo.Context) error {
	ctx := c.Request().Context()

	cabinetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return ErrInvalidInput
	}

	err = h.cabinet.DeleteCabinet(ctx, cabinetID)
	if err != nil {
		h.logger.Error("Get list schedule error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, nil).Send(c)
}

func cabinetToView(model *cabinets.Cabinet, facultyName string) Cabinet {
	var equipment *CabinetEquipment
	if model.Equipment != nil {
		equipment = &CabinetEquipment{
			Furniture:         model.Equipment.Furniture,
			TechnicalMeans:    model.Equipment.TechnicalMeans,
			ComputerEquipment: model.Equipment.ComputerEquipment,
		}
	}

	return Cabinet{
		ID:                                 model.ID,
		FacultyID:                          model.FacultyID,
		FacultyName:                        facultyName,
		Type:                               int8(model.Type),
		Auditorium:                         model.Auditorium,
		SuitableForPeoplesWithSpecialNeeds: model.SuitableForPeoplesWithSpecialNeeds,
		Building:                           model.Building,
		Appointment:                        model.Appointment,
		Equipment:                          equipment,
	}
}
