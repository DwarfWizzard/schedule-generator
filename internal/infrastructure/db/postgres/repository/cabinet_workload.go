package repository

import (
	"context"
	cabinetworkload "schedule-generator/internal/domain/cabinet_workload"
	"schedule-generator/internal/domain/schedules"
	"schedule-generator/internal/infrastructure/db/postgres/schema"
)

func (r *Repository) ListCycledCabinetWorkload(ctx context.Context) ([]cabinetworkload.CabinetWorkloadItem, error) {
	var list []schema.CabinetWorkloadItem

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
		Where("schedules.type = ?", schedules.ScheduleTypeCycled).
		Where("schedule_items.weektype IS NOT NULL").
		Order("schedule_items.cabinet_building, schedule_items.cabinet_auditorium, schedule_items.weekday, schedule_items.lesson_number").
		Scan(&list).Error

	if err != nil {
		return nil, err
	}

	result := make([]cabinetworkload.CabinetWorkloadItem, len(list))
	for i, schemaWLCabItem := range list {
		result[i] = *schema.CabinetWorkloadItemFromSchema(&schemaWLCabItem)
	}

	return result, nil
}

func (r *Repository) ListCycledPractices(ctx context.Context) ([]cabinetworkload.CabinetWorkloadPractice, error) {
	var list []schema.CabinetWorkloadPractice

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
		Where("schedules.type = ?", schedules.ScheduleTypeCycled).
		Order("edu_groups.number, practices.start_date").
		Scan(&list).Error

	if err != nil {
		return nil, err
	}

	result := make([]cabinetworkload.CabinetWorkloadPractice, len(list))
	for i, schemaWLPract := range list {
		result[i] = *schema.CabinetWorkloadPracticeFromSchema(&schemaWLPract)
	}

	return result, err
}
