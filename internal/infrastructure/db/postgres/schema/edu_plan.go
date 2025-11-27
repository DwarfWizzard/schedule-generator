package schema

import (
	eduplans "schedule-generator/internal/domain/edu_plans"

	"github.com/google/uuid"
)

type Module struct {
	Discipline string    `gorm:"column:discipline;primaryKey;uniqueIndex:module_discipline_eduplan_unique"`
	EduPlanID  uuid.UUID `gorm:"column:edu_plan_id;type:string;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;uniqueIndex:module_discipline_eduplan_unique"`
}

type EduPlan struct {
	ID          uuid.UUID `gorm:"column:id;type:string;primaryKey"`
	DirectionID uuid.UUID `gorm:"uniqueIndex:edu_plan_direction_profile_year_unique;column:direction_id;type:string;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Profile     string    `gorm:"uniqueIndex:edu_plan_direction_profile_year_unique;column:profile;not null"`
	Year        int64     `gorm:"uniqueIndex:edu_plan_direction_profile_year_unique;column:year;not null"`

	Modules []Module `gorm:"foreignKey:edu_plan_id"`
}

// EduPlanToSchema
func EduPlanToSchema(eduplan *eduplans.EduPlan) *EduPlan {
	modules := eduplan.ListModule()

	schema := EduPlan{
		ID:          eduplan.ID,
		DirectionID: eduplan.DirectionID,
		Profile:     eduplan.Profile,
		Year:        eduplan.Year,
		Modules:     make([]Module, len(modules)),
	}

	for i, module := range modules {
		schema.Modules[i] = Module{
			Discipline: module.Discipline,
			EduPlanID:  eduplan.ID,
		}
	}

	return &schema
}

// EduPlanFromSchema
func EduPlanFromSchema(schema *EduPlan) *eduplans.EduPlan {
	model := eduplans.EduPlan{
		ID:          schema.ID,
		DirectionID: schema.DirectionID,
		Profile:     schema.Profile,
		Year:        schema.Year,

		Modules: make([]eduplans.Module, 0, len(schema.Modules)),
	}

	for _, module := range schema.Modules {
		_, err := model.AddModule(module.Discipline)
		if err != nil {
			// ignore invalid data from db
			continue
		}
	}

	return &model
}
