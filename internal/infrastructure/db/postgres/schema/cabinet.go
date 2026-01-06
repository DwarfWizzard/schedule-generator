package schema

import "github.com/google/uuid"

type Cabinet struct {
	ID                                 uuid.UUID `gorm:"column:id;type:string;primaryKey"`
	FacultyID                          uuid.UUID `gorm:"column:faculty_id;type:string;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Faculty                            *Faculty  `gorm:"foreignKey:faculty_id"`
	Building                           string    `gorm:"column:building;type:string"`
	Auditorium                         string    `gorm:"column:auditorium;type:string;unique;not null"`
	Type                               int8      `gorm:"column:type;not null"`
	Appointment                        *string   `gorm:"column:appointment"`
	EquipmentFurniture                 *string   `gorm:"column:equipment_furniture"`
	EquipmentTechnicalMeans            *string   `gorm:"column:equipment_technical_means"`
	EquipmentСomputerEquipment         *string   `gorm:"column:equipment_computer"`
	SuitableForPeoplesWithSpecialNeeds bool      `gorm:"column:suitable_for_peoples_with_special_needs"`
}
