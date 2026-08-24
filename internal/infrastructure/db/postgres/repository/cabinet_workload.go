package repository

import (
	"context"
	"schedule-generator/internal/infrastructure/db/postgres/schema"
	"time"
)

type CabinetWorkloadItem struct {
	Discipline        string       `gorm:"column:discipline"`
	Weekday           time.Weekday `gorm:"column:weekday"`
	LessonNumber      int8         `gorm:"column:lesson_number"`
	Subgroup          int8         `gorm:"column:subgroup"`
	Weektype          *int8        `gorm:"column:weektype"`
	LessonType        int8         `gorm:"column:lesson_type"`
	StudentsCount     int16        `gorm:"column:students_count"`
	CabinetAuditorium string       `gorm:"column:cabinet_auditorium"`
	CabinetBuilding   string       `gorm:"column:cabinet_building"`

	TeacherName string `gorm:"column:teacher_name"`

	EduGroupNumber string `gorm:"column:edu_group_number"`
}

func (r *Repository) ListCycledCabinetWorkload(ctx context.Context) ([]CabinetWorkloadItem, error) {
	var result []CabinetWorkloadItem

	err := r.client.WithContext(ctx).
		Model(&schema.ScheduleItem{}).
		Select(`
			schedule_items.discipline,
			schedule_items.weekday,
			schedule_items.lesson_number,
			schedule_items.subgroup,
			schedule_items.weektype,
			schedule_items.lesson_type,
			schedule_items.students_count,
			schedule_items.cabinet_auditorium,
			schedule_items.cabinet_building,
			teachers.name AS teacher_name,
			edu_groups.number AS edu_group_number
		`).
		Joins("JOIN schedules ON schedules.id = schedule_items.schedule_id").
		Joins("JOIN edu_groups ON edu_groups.id = schedules.edu_group_id").
		Joins("JOIN teachers ON teachers.id = schedule_items.teacher_id").
		Where("schedules.type = 1").
		Where("schedule_items.weektype IS NOT NULL").
		Order("schedule_items.cabinet_building, schedule_items.cabinet_auditorium, schedule_items.weekday, schedule_items.lesson_number").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

type CabinetWorkloadPractice struct {
	PracticeType   int8      `gorm:"column:practice_type"`
	StartDate      time.Time `gorm:"column:start_date"`
	EndDate        time.Time `gorm:"column:end_date"`
	EduGroupNumber string    `gorm:"column:edu_group_number"`
}

func (r *Repository) ListCycledPractices(ctx context.Context) ([]CabinetWorkloadPractice, error) {
	var result []CabinetWorkloadPractice

	err := r.client.WithContext(ctx).
		Model(&schema.Practice{}).
		Select(`
			practices.type       AS practice_type,
			practices.start_date AS start_date,
			practices.end_date   AS end_date,
			edu_groups.number    AS edu_group_number
		`).
		Joins("JOIN schedules ON schedules.id = practices.schedule_id").
		Joins("JOIN edu_groups ON edu_groups.id = schedules.edu_group_id").
		Where("schedules.type = 1").
		Order("edu_groups.number, practices.start_date").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return result, err
}
