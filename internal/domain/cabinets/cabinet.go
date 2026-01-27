package cabinets

import (
	"errors"

	"github.com/google/uuid"
)

type CabinetType int8

var cabinetTypeNames = []string{
	"practice",
	"lecture",
	"mixed",
}

const (
	CabinetTypePractice = iota
	CabinetTypeLecture
	CabinetTypeMixed
)

func NewCabinetType(t int8) (CabinetType, error) {
	if int(t) < 0 || int(t) >= len(cabinetTypeNames) {
		return 0, errors.New("unknown cabinet type")
	}

	return CabinetType(t), nil
}

func (t CabinetType) String() string {
	i := int(t)
	if i < 0 || i >= len(cabinetTypeNames) {
		return "unknown"
	}

	return cabinetTypeNames[t]
}

type CabinetEquipment struct {
	Furniture         string
	TechnicalMeans    string
	ComputerEquipment string
}

type Cabinet struct {
	ID                                 uuid.UUID
	FacultyID                          uuid.UUID
	Type                               CabinetType
	Building                           string
	Auditorium                         string
	SuitableForPeoplesWithSpecialNeeds bool
	Appointment                        *string
	Equipment                          *CabinetEquipment
}

func (c *Cabinet) Validate() error {
	var argErr error

	if len(c.Auditorium) == 0 {
		argErr = errors.Join(argErr, errors.New("invalid auditorium value"))
	}

	if len(c.Building) == 0 {
		argErr = errors.Join(argErr, errors.New("invalid building value"))
	}

	if c.Appointment != nil && len(*c.Appointment) == 0 {
		argErr = errors.Join(argErr, errors.New("invalid appointment value"))
	}

	if argErr != nil {
		return argErr
	}

	return nil
}

func NewCabinet(facultyID uuid.UUID, cabinetType CabinetType, auditorium string, suitableForPeoplesWithSpecialNeeds bool, building string, appointment *string, equipment *CabinetEquipment) (*Cabinet, error) {
	cab := Cabinet{
		ID:                                 uuid.New(),
		FacultyID:                          facultyID,
		Type:                               cabinetType,
		Auditorium:                         auditorium,
		SuitableForPeoplesWithSpecialNeeds: suitableForPeoplesWithSpecialNeeds,
		Building:                           building,
		Appointment:                        appointment,
		Equipment:                          equipment,
	}

	if err := cab.Validate(); err != nil {
		return nil, err
	}

	return &cab, nil
}
