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
	return cabinetTypeNames[t]
}

type CabinetEquipment struct {
	Furniture         string
	TechnicalMeans    string
	СomputerEquipment string
}

type Cabinet struct {
	ID                                 uuid.UUID
	FacultyID                          uuid.UUID
	Building                           *string
	Auditorium                         string
	Type                               CabinetType
	Appointment                        string
	Equipment                          CabinetEquipment
	SuitableForPeoplesWithSpecialNeeds bool
}
