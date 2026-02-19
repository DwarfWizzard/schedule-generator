package schema

import (
	edudirections "schedule-generator/internal/domain/edu_directions"

	"github.com/google/uuid"
)

type EduDirection struct {
	ID   uuid.UUID `gorm:"column:id;type:string;primaryKey"`
	Name string    `gorm:"column:name;not null"`
}

// EduDirectionToSchema
func EduDirectionToSchema(model *edudirections.EduDirection) *EduDirection {
	return &EduDirection{
		ID:   model.ID,
		Name: model.Name,
	}
}

// EduDirectionFromSchema
func EduDirectionFromSchema(scheme *EduDirection) *edudirections.EduDirection {
	return &edudirections.EduDirection{
		ID:   scheme.ID,
		Name: scheme.Name,
	}
}
