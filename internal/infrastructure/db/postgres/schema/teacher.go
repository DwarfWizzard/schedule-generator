package schema

import (
	"schedule-generator/internal/domain/teachers"

	"github.com/google/uuid"
)

type Teacher struct {
	ID           uuid.UUID   `gorm:"column:id;type:string;primaryKey"`
	ExternalID   string      `gorm:"column:external_id;unique;not null"`
	Name         string      `gorm:"column:name;not null"`
	Position     string      `gorm:"column:position;not null"`
	Degree       string      `gorm:"column:degree;not null"`
	DepartmentID uuid.UUID   `gorm:"column:department_id;type:string;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Department   *Department `gorm:"foreignKey:department_id"`
}

// TeacherToSchema
func TeacherToSchema(model *teachers.Teacher) *Teacher {
	return &Teacher{
		ID:           model.ID,
		ExternalID:   model.ExternalID,
		Name:         model.Name,
		Position:     model.Position,
		Degree:       model.Degree,
		DepartmentID: model.DepartmentID,
	}
}

// TeacherFromSchema
func TeacherFromSchema(scheme *Teacher) *teachers.Teacher {
	return &teachers.Teacher{
		ID:           scheme.ID,
		ExternalID:   scheme.ExternalID,
		Name:         scheme.Name,
		Position:     scheme.Position,
		Degree:       scheme.Degree,
		DepartmentID: scheme.DepartmentID,
	}
}
