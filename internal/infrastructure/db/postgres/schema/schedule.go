package schema

import (
	"schedule-generator/internal/domain/schedules"
	"time"

	"github.com/google/uuid"
)

type ScheduleItem struct {
	ID            int64        `gorm:"column:id;autoIncrement;primaryKey"`
	ScheduleID    uuid.UUID    `gorm:"column:schedule_id;type:string;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Discipline    string       `gorm:"column:discipline;not null"`
	TeacherID     uuid.UUID    `gorm:"column:teacher_id;type:string;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Teacher       *Teacher     `gorm:"foreignKey:teacher_id"`
	Weekday       time.Weekday `gorm:"column:weekday;not null;default:1"`
	StudentsCount int16        `gorm:"column:students_count;not null;default:0"`
	Date          *time.Time   `gorm:"column:date"`
	LessonNumber  int8         `gorm:"column:lesson_number;not null;default:0"`
	Subgroup      int8         `gorm:"column:subgroup;not null;default:0"`
	Weektype      *int8        `gorm:"column:weektype"`
	Weeknum       *int         `gorm:"column:weeknum"`
	LessonType    int8         `gorm:"column:lesson_type;not null"`
	Classroom     string       `gorm:"column:classroom;not null"`
}

type Schedule struct {
	ID         uuid.UUID `gorm:"column:id;type:string;primaryKey"`
	EduGroupID uuid.UUID `gorm:"column:edu_group_id;type:string;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	EduGroup   *EduGroup `gorm:"foreignKey:edu_group_id"`
	Semester   int       `gorm:"column:semester;not null"`
	Type       int8      `gorm:"column:type;not null"`

	// Cycled schedule specific
	StartDate *time.Time `gorm:"column:start_date"`
	EndDate   *time.Time `gorm:"column:end_date"`

	Items []ScheduleItem `gorm:"foreignKey:schedule_id"`
}

// ScheduleToSchema
func ScheduleToSchema(model *schedules.Schedule) *Schedule {
	items := model.ListItem()

	schema := Schedule{
		ID:         model.ID,
		EduGroupID: model.EduGroupID,
		Semester:   model.Semester,
		Type:       int8(model.Type),
		Items:      make([]ScheduleItem, len(items)),
	}

	if model.Type == schedules.ScheduleTypeCycled {
		schema.StartDate = &model.Cycled.StartDate
		schema.EndDate = &model.Cycled.EndDate
	}

	for i, item := range items {
		si := ScheduleItem{
			ScheduleID:    model.ID,
			Discipline:    item.Discipline,
			TeacherID:     item.TeacherID,
			Weekday:       item.Weekday,
			StudentsCount: item.StudentsCount,
			Date:          item.Date,
			LessonNumber:  item.LessonNumber,
			Subgroup:      item.Subgroup,
			Weeknum:       item.Weeknum,
			LessonType:    int8(item.LessonType),
			Classroom:     string(item.Classroom),
		}

		if item.Weektype != nil {
			wt := int8(*item.Weektype)
			si.Weektype = &wt
		}

		schema.Items[i] = si
	}

	return &schema
}

// ScheduleFromSchema
func ScheduleFromSchema(schema *Schedule) *schedules.Schedule {
	model := schedules.Schedule{
		ID:         schema.ID,
		EduGroupID: schema.EduGroupID,
		Semester:   schema.Semester,
		Type:       schedules.ScheduleType(schema.Type),
	}

	switch model.Type {
	case schedules.ScheduleTypeCycled:
		if schema.StartDate == nil || schema.EndDate == nil {
			return &model
		}

		model.Cycled = &schedules.CycledSchedule{
			StartDate: *schema.StartDate,
			EndDate:   *schema.EndDate,
			Items:     make(map[time.Weekday][]schedules.ScheduleItem),
		}

		for _, item := range schema.Items {
			if item.Weektype == nil {
				continue
			}

			err := model.Cycled.AddItem(
				item.Discipline,
				item.TeacherID,
				item.Weekday,
				item.StudentsCount,
				item.LessonNumber,
				item.Subgroup,
				*item.Weektype,
				item.LessonType,
				item.Classroom,
			)

			if err != nil {
				// ignore invalid data from db
				continue
			}
		}
	case schedules.ScheduleTypeCalendar:
		model.Calendar = &schedules.CalendarSchedule{
			Items: make([]schedules.ScheduleItem, 0, len(schema.Items)),
		}

		for _, item := range schema.Items {
			if item.Weeknum == nil || item.Date == nil {
				continue
			}

			err := model.Calendar.AddItem(
				item.Discipline,
				item.TeacherID,
				*item.Date,
				item.StudentsCount,
				item.LessonNumber,
				item.Subgroup,
				*item.Weeknum,
				item.LessonType,
				item.Classroom,
			)

			if err != nil {
				// ignore invalid data from db
				continue
			}
		}
	}

	return &model
}
