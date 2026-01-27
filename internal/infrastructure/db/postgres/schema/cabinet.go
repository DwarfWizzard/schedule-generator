package schema

import (
	"schedule-generator/internal/domain/cabinets"

	"github.com/google/uuid"
)

type Cabinet struct {
	ID                                 uuid.UUID `gorm:"column:id;type:string;primaryKey"`
	FacultyID                          uuid.UUID `gorm:"column:faculty_id;type:string;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Faculty                            *Faculty  `gorm:"foreignKey:faculty_id"`
	Building                           string    `gorm:"column:building;type:string;uniqueIndex:cabinet_building_auditorium;not null"`
	Auditorium                         string    `gorm:"column:auditorium;type:string;uniqueIndex:cabinet_building_auditorium;not null"`
	Type                               int8      `gorm:"column:type;not null"`
	Appointment                        *string   `gorm:"column:appointment"`
	EquipmentFurniture                 *string   `gorm:"column:equipment_furniture"`
	EquipmentTechnicalMeans            *string   `gorm:"column:equipment_technical_means"`
	EquipmentСomputerEquipment         *string   `gorm:"column:equipment_computer"`
	SuitableForPeoplesWithSpecialNeeds bool      `gorm:"column:suitable_for_peoples_with_special_needs"`
}

// CabinetToSchema
func CabinetToSchema(c *cabinets.Cabinet) *Cabinet {
	cab := Cabinet{
		ID:                                 c.ID,
		FacultyID:                          c.FacultyID,
		Building:                           c.Building,
		Auditorium:                         c.Auditorium,
		Type:                               int8(c.Type),
		Appointment:                        c.Appointment,
		SuitableForPeoplesWithSpecialNeeds: c.SuitableForPeoplesWithSpecialNeeds,
	}

	if c.Equipment != nil {
		cab.EquipmentFurniture = &c.Equipment.Furniture
		cab.EquipmentTechnicalMeans = &c.Equipment.TechnicalMeans
		cab.EquipmentСomputerEquipment = &c.Equipment.ComputerEquipment
	}

	return &cab
}

// CabinetFromSchema
func CabinetFromSchema(scheme *Cabinet) *cabinets.Cabinet {
	cab := cabinets.Cabinet{
		ID:                                 scheme.ID,
		FacultyID:                          scheme.FacultyID,
		Building:                           scheme.Building,
		Auditorium:                         scheme.Auditorium,
		Type:                               cabinets.CabinetType(scheme.Type),
		Appointment:                        scheme.Appointment,
		SuitableForPeoplesWithSpecialNeeds: scheme.SuitableForPeoplesWithSpecialNeeds,
	}

	if scheme.EquipmentFurniture != nil {
		if cab.Equipment == nil {
			cab.Equipment = &cabinets.CabinetEquipment{}
		}

		cab.Equipment.Furniture = *scheme.EquipmentFurniture
	}

	if scheme.EquipmentTechnicalMeans != nil {
		if cab.Equipment == nil {
			cab.Equipment = &cabinets.CabinetEquipment{}
		}

		cab.Equipment.TechnicalMeans = *scheme.EquipmentTechnicalMeans
	}

	if scheme.EquipmentСomputerEquipment != nil {
		if cab.Equipment == nil {
			cab.Equipment = &cabinets.CabinetEquipment{}
		}

		cab.Equipment.ComputerEquipment = *scheme.EquipmentСomputerEquipment
	}

	return &cab
}
