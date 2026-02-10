package repository

import (
	"context"
	"errors"

	"schedule-generator/internal/domain/schedules"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/internal/infrastructure/db/postgres/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SaveSchedule
func (r *Repository) SaveSchedule(ctx context.Context, d *schedules.Schedule) error {
	s := schema.ScheduleToSchema(d)

	err := r.client.WithContext(ctx).Delete(&schema.ScheduleItem{}, "schedule_id = ?", s.ID).Error
	if err != nil {
		return err
	}

	err = r.client.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return db.ErrorUniqueViolation
		}

		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return db.ErrorAssociationViolation
		}

		return err
	}

	return nil
}

// GetSchedule
func (r *Repository) GetSchedule(ctx context.Context, id uuid.UUID) (*schedules.Schedule, error) {
	var s schema.Schedule
	err := r.client.WithContext(ctx).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order(`
			schedule_items.lesson_number,
			schedule_items.subgroup,
			schedule_items.date NULLS LAST,
			schedule_items.weektype NULLS LAST
		`)
	}).Where("id = ?", id.String()).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.ScheduleFromSchema(&s), nil
}

// GetScheduleFacultyID
func (r *Repository) GetScheduleFacultyID(ctx context.Context, scheduleID uuid.UUID) (uuid.UUID, error) {
	var s schema.Schedule
	err := r.client.WithContext(ctx).Preload("EduGroup.EduPlan.Direction.Department").Preload(clause.Associations).Where("id = ?", scheduleID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.UUID{}, db.ErrorNotFound
		}

		return uuid.UUID{}, err
	}

	return s.EduGroup.EduPlan.Direction.Department.FacultyID, nil
}

// GetScheduleByEduGroupIDAndSemester
func (r *Repository) GetScheduleByEduGroupIDAndSemester(ctx context.Context, eduGroupID uuid.UUID, semester int) (*schedules.Schedule, error) {
	var s schema.Schedule
	err := r.client.WithContext(ctx).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order(`
			schedule_items.date NULLS LAST,
			schedule_items.weektype NULLS LAST,
			schedule_items.lesson_number,
			schedule_items.subgroup
		`)
	}).Where("edu_group_id = ? AND semester = ?", eduGroupID.String(), semester).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.ScheduleFromSchema(&s), nil
}

// ListSchedule
func (r *Repository) ListSchedule(ctx context.Context) ([]schedules.Schedule, error) {
	var list []schema.Schedule
	err := r.client.WithContext(ctx).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order(`
			schedule_items.date NULLS LAST,
			schedule_items.weektype NULLS LAST,
			schedule_items.lesson_number,
			schedule_items.subgroup
		`)
	}).Order("edu_group_id ASC, semester DESC").Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]schedules.Schedule, len(list))
	for i, v := range list {
		result[i] = *schema.ScheduleFromSchema(&v)
	}

	return result, nil
}

// ListSchedule
func (r *Repository) ListScheduleByFaculty(ctx context.Context, facultyID uuid.UUID) ([]schedules.Schedule, error) {
	var list []schema.Schedule
	err := r.client.WithContext(ctx).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order(`
			schedule_items.date NULLS LAST,
			schedule_items.weektype NULLS LAST,
			schedule_items.lesson_number,
			schedule_items.subgroup
		`)
	}).Joins("EduGroup.EduPlan.Direction.Department").Where("Department.faculty_id = ?", facultyID).Order("edu_group_id ASC, semester DESC").Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]schedules.Schedule, len(list))
	for i, v := range list {
		result[i] = *schema.ScheduleFromSchema(&v)
	}

	return result, nil
}

// ListScheduleByEduGroup
func (r *Repository) ListScheduleByEduGroup(ctx context.Context, groupID uuid.UUID) ([]schedules.Schedule, error) {
	var list []schema.Schedule
	err := r.client.WithContext(ctx).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order(`
			schedule_items.date NULLS LAST,
			schedule_items.weektype NULLS LAST,
			schedule_items.lesson_number,
			schedule_items.subgroup
		`)
	}).Where("edu_group_id = ?", groupID).Order("semester DESC").Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]schedules.Schedule, len(list))
	for i, v := range list {
		result[i] = *schema.ScheduleFromSchema(&v)
	}

	return result, nil
}

// DeleteSchedule
func (r *Repository) DeleteSchedule(ctx context.Context, id uuid.UUID) error {
	err := r.client.WithContext(ctx).Where("id = ?", id).Delete(&schema.Schedule{}).Error
	if err != nil {
		return err
	}

	return nil
}
