package repository

import (
	"context"
	"errors"

	"schedule-generator/internal/domain/teachers"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/internal/infrastructure/db/postgres/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SaveTeacher
func (r *Repository) SaveTeacher(ctx context.Context, d *teachers.Teacher) error {
	s := schema.TeacherToSchema(d)
	err := r.client.WithContext(ctx).Save(s).Error
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

// GetTeacher
func (r *Repository) GetTeacher(ctx context.Context, id uuid.UUID) (*teachers.Teacher, error) {
	var s schema.Teacher
	err := r.client.WithContext(ctx).Where("id = ?", id.String()).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.TeacherFromSchema(&s), nil
}

// ListTeacher
func (r *Repository) ListTeacher(ctx context.Context) ([]teachers.Teacher, error) {
	var list []schema.Teacher
	err := r.client.WithContext(ctx).Order("name ASC").Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]teachers.Teacher, len(list))
	for i, v := range list {
		result[i] = *schema.TeacherFromSchema(&v)
	}

	return result, nil
}

// ListTeacherByDepartment
func (r *Repository) ListTeacherByDepartment(ctx context.Context, depID string) ([]teachers.Teacher, error) {
	var list []schema.Teacher
	err := r.client.WithContext(ctx).Where("department_id = ?", depID).Order("name ASC").Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]teachers.Teacher, len(list))
	for i, v := range list {
		result[i] = *schema.TeacherFromSchema(&v)
	}

	return result, nil
}

// MapTeachersByIDs
func (r *Repository) MapTeacherByIDs(ctx context.Context, teacherIDs uuid.UUIDs) (map[uuid.UUID]teachers.Teacher, error) {
	var list []schema.Teacher
	err := r.client.WithContext(ctx).Where("id IN ?", teacherIDs).Find(&list).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]teachers.Teacher, len(list))
	for _, v := range list {
		result[v.ID] = *schema.TeacherFromSchema(&v)
	}

	return result, nil
}

// DeleteTeacher
func (r *Repository) DeleteTeacher(ctx context.Context, id uuid.UUID) error {
	err := r.client.WithContext(ctx).Where("id = ?", id).Delete(&schema.Teacher{}).Error
	if err != nil {
		return err
	}

	return nil
}
