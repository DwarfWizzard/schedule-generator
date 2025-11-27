package schema

import (
	edugroups "schedule-generator/internal/domain/edu_groups"

	"github.com/google/uuid"
)

type EduGroup struct {
	ID            uuid.UUID `gorm:"column:id;type:string;primaryKey"`
	Number        string    `gorm:"column:number;not null;unique"`
	EduPlanID     uuid.UUID `gorm:"column:edu_plan_id;type:string;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Profile       string    `gorm:"column:profile;not null"`
	AdmissionYear int64     `gorm:"column:admission_year;not null"`
}

// EduGroupToSchema
func EduGroupToSchema(model *edugroups.EduGroup) *EduGroup {
	return &EduGroup{
		ID:            model.ID,
		Number:        model.Number,
		EduPlanID:     model.EduPlanID,
		Profile:       model.Profile,
		AdmissionYear: model.AdmissionYear,
	}
}

// EduGroupFromSchema
func EduGroupFromSchema(scheme *EduGroup) *edugroups.EduGroup {
	return &edugroups.EduGroup{
		ID:            scheme.ID,
		Number:        scheme.Number,
		EduPlanID:     scheme.EduPlanID,
		Profile:       scheme.Profile,
		AdmissionYear: scheme.AdmissionYear,
	}
}
