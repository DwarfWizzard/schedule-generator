package schema

import (
	"schedule-generator/internal/domain/departments"

	"github.com/google/uuid"
)

type Department struct {
	ID         uuid.UUID `gorm:"column:id;type:string;primaryKey"`
	ExternalID string    `gorm:"column:external_id;unique;not null"`
	FacultyID  uuid.UUID `gorm:"column:faculty_id;type:string;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Faculty    *Faculty  `gorm:"foreignKey:faculty_id"`
	Name       string    `gorm:"column:name;not null"`
}

// DepartmentToSchema
func DepartmentToSchema(d *departments.Department) *Department {
	return &Department{
		ID:         d.ID,
		ExternalID: d.ExternalID,
		FacultyID:  d.FacultyID,
		Name:       d.Name,
	}
}

// DepartmentFromSchema
func DepartmentFromSchema(scheme *Department) *departments.Department {
	return &departments.Department{
		ID:         scheme.ID,
		ExternalID: scheme.ExternalID,
		FacultyID:  scheme.FacultyID,
		Name:       scheme.Name,
	}
}
