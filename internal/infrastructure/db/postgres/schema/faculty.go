package schema

import (
	"schedule-generator/internal/domain/faculties"

	"github.com/google/uuid"
)

type Faculty struct {
	ID   uuid.UUID `gorm:"column:id;type:string;primaryKey"`
	Name string    `gorm:"column:name;not null"`
}

// FacultyToSchema
func FacultyToSchema(model *faculties.Faculty) *Faculty {
	return &Faculty{
		ID:   model.ID,
		Name: model.Name,
	}
}

// FacultyFromSchema
func FacultyFromSchema(scheme *Faculty) *faculties.Faculty {
	return &faculties.Faculty{
		ID:   scheme.ID,
		Name: scheme.Name,
	}
}
