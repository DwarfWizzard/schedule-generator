package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"schedule-generator/internal/domain/cabinets"
	"schedule-generator/internal/domain/faculties"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/pkg/execerror"

	"github.com/google/uuid"
)

type CabinetUsecaseRepo interface {
	cabinets.Repository
	faculties.Repository

	MapFacultiesByCabinets(ctx context.Context, cabinetIDs uuid.UUIDs) (map[uuid.UUID]faculties.Faculty, error)
}

type CabinetUsecase struct {
	repo   CabinetUsecaseRepo
	logger *slog.Logger
}

func NewCabinetUsecase(repo CabinetUsecaseRepo, logger *slog.Logger) *CabinetUsecase {
	return &CabinetUsecase{
		repo:   repo,
		logger: logger,
	}
}

type Equipment struct {
	Furniture         string
	TechnicalMeans    string
	СomputerEquipment string
}

type CreateCabinetInput struct {
	FacultyID                          uuid.UUID
	CabinetType                        int8
	Building                           string
	Auditorium                         string
	SuitableForPeoplesWithSpecialNeeds bool
	Appointment                        *string
	Equipment                          *Equipment
}

type CreateCabinetOutput struct {
	cabinets.Cabinet
	FacultyName string
}

// CreateCabinet
func (uc *CabinetUsecase) CreateCabinet(ctx context.Context, input CreateCabinetInput) (*CreateCabinetOutput, error) {
	logger := uc.logger

	faculty, err := uc.repo.GetFaculty(ctx, input.FacultyID)
	if err != nil {
		logger.Error("Get faculty error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("faculty not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	cabinetType, err := cabinets.NewCabinetType(input.CabinetType)
	if err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	var equipment *cabinets.CabinetEquipment

	if input.Equipment != nil {
		equipment = &cabinets.CabinetEquipment{
			Furniture:         input.Equipment.Furniture,
			TechnicalMeans:    input.Equipment.TechnicalMeans,
			ComputerEquipment: input.Equipment.Furniture,
		}
	}

	cabinet, err := cabinets.NewCabinet(faculty.ID, cabinetType, input.Auditorium, input.SuitableForPeoplesWithSpecialNeeds, input.Building, input.Appointment, equipment)
	if err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveCabinet(ctx, cabinet)
	if err != nil {
		logger.Error("Save edu cabinet error", "error", err)

		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("cabinet with provided external id already exists"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &CreateCabinetOutput{
		Cabinet:     *cabinet,
		FacultyName: faculty.Name,
	}, nil
}

type GetCabinetOutput struct {
	cabinets.Cabinet
	FacultyName string
}

// GetCabinet
func (uc *CabinetUsecase) GetCabinet(ctx context.Context, cabinetID uuid.UUID) (*GetCabinetOutput, error) {
	logger := uc.logger

	cabinet, err := uc.repo.GetCabinet(ctx, cabinetID)
	if err != nil {
		logger.Error("List cabinet error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("cabinet not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	faculty, err := uc.repo.GetFaculty(ctx, cabinet.FacultyID)
	if err != nil {
		logger.Error("Get faculty error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &GetCabinetOutput{
		Cabinet:     *cabinet,
		FacultyName: faculty.Name,
	}, nil
}

type ListCabinetOutput = []GetCabinetOutput

// ListCabinet
func (uc *CabinetUsecase) ListCabinet(ctx context.Context) (ListCabinetOutput, error) {
	logger := uc.logger

	cabinets, err := uc.repo.ListCabinet(ctx)
	if err != nil {
		logger.Error("List cabinet error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	cabinetIDs := make(uuid.UUIDs, len(cabinets))

	for i, dep := range cabinets {
		cabinetIDs[i] = dep.ID
	}

	faculties, err := uc.repo.MapFacultiesByCabinets(ctx, cabinetIDs)
	if err != nil {
		logger.Error("Map faculties by cabinets error", "error", err)
		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	result := make(ListCabinetOutput, len(cabinets))
	for i, dep := range cabinets {
		faculty, ok := faculties[dep.FacultyID]
		if !ok {
			logger.Error(fmt.Sprintf("Faculty for cabinet %s not found", dep.ID))
			return nil, execerror.NewExecError(execerror.TypeInternal, nil)
		}

		result[i] = GetCabinetOutput{
			Cabinet:     dep,
			FacultyName: faculty.Name,
		}
	}

	return result, nil
}

type UpdateCabinetInput struct {
	CabinetID                          uuid.UUID
	FacultyID                          *uuid.UUID
	CabinetType                        *int8
	Building                           *string
	Auditorium                         *string
	SuitableForPeoplesWithSpecialNeeds *bool
	Appointment                        *string
	Equipment                          *Equipment
}

type UpdateCabinetOutput struct {
	cabinets.Cabinet
	FacultyName string
}

// UpdateCabinet
func (uc *CabinetUsecase) UpdateCabinet(ctx context.Context, input UpdateCabinetInput) (*UpdateCabinetOutput, error) {
	logger := uc.logger

	cabinet, err := uc.repo.GetCabinet(ctx, input.CabinetID)
	if err != nil {
		logger.Error("Get cabinet error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("cabinet not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if input.FacultyID != nil {
		cabinet.FacultyID = *input.FacultyID
	}

	faculty, err := uc.repo.GetFaculty(ctx, cabinet.FacultyID)
	if err != nil {
		logger.Error("Get faculty error", "error", err)
		if errors.Is(err, db.ErrorNotFound) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, errors.New("faculty not found"))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	if input.CabinetType != nil {
		cabinetType, err := cabinets.NewCabinetType(*input.CabinetType)
		if err != nil {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
		}

		cabinet.Type = cabinetType
	}

	if input.Building != nil {
		cabinet.Building = *input.Building
	}

	if input.Auditorium != nil {
		cabinet.Auditorium = *input.Auditorium
	}

	if input.SuitableForPeoplesWithSpecialNeeds != nil {
		cabinet.SuitableForPeoplesWithSpecialNeeds = *input.SuitableForPeoplesWithSpecialNeeds
	}
	if input.Appointment != nil {
		if len(*input.Appointment) == 0 {
			cabinet.Appointment = nil
		}

		cabinet.Appointment = input.Appointment
	}

	if input.Equipment != nil {
		cabinet.Equipment = &cabinets.CabinetEquipment{
			Furniture:         input.Equipment.Furniture,
			TechnicalMeans:    input.Equipment.TechnicalMeans,
			ComputerEquipment: input.Equipment.Furniture,
		}
	}

	if err := cabinet.Validate(); err != nil {
		return nil, execerror.NewExecError(execerror.TypeInvalidInput, err)
	}

	err = uc.repo.SaveCabinet(ctx, cabinet)
	if err != nil {
		logger.Error("Save edu cabinet error", "error", err)

		if errors.Is(err, db.ErrorUniqueViolation) {
			return nil, execerror.NewExecError(execerror.TypeInvalidInput, fmt.Errorf("cabinet '%s' in building '%s' already exists", cabinet.Auditorium, cabinet.Building))
		}

		return nil, execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return &UpdateCabinetOutput{
		Cabinet:     *cabinet,
		FacultyName: faculty.Name,
	}, nil
}

// DeleteCabinet
func (uc *CabinetUsecase) DeleteCabinet(ctx context.Context, cabinetID uuid.UUID) error {
	logger := uc.logger

	err := uc.repo.DeleteCabinet(ctx, cabinetID)
	if err != nil {
		logger.Error("Delete edu cabinet error", "error", err)
		return execerror.NewExecError(execerror.TypeInternal, nil)
	}

	return nil
}
