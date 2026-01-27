package repository

import (
	"context"
	"errors"

	"schedule-generator/internal/domain/faculties"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/internal/infrastructure/db/postgres/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SaveFaculty
func (r *Repository) SaveFaculty(ctx context.Context, d *faculties.Faculty) error {
	s := schema.FacultyToSchema(d)
	err := r.client.WithContext(ctx).Save(s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return db.ErrorUniqueViolation
		}

		return err
	}

	return nil
}

// GetFaculty
func (r *Repository) GetFaculty(ctx context.Context, id uuid.UUID) (*faculties.Faculty, error) {
	var s schema.Faculty
	err := r.client.WithContext(ctx).Where("id = ?", id.String()).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.FacultyFromSchema(&s), nil
}

// ListFaculty
func (r *Repository) ListFaculty(ctx context.Context) ([]faculties.Faculty, error) {
	var list []schema.Faculty
	err := r.client.WithContext(ctx).Order("name ASC").Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]faculties.Faculty, len(list))
	for i, v := range list {
		result[i] = *schema.FacultyFromSchema(&v)
	}

	return result, nil
}

// MapFacultiesByDepartments
func (r *Repository) MapFacultiesByDepartments(ctx context.Context, departmentIDs uuid.UUIDs) (map[uuid.UUID]faculties.Faculty, error) {
	var depList []schema.Department

	err := r.client.WithContext(ctx).Preload("Faculty").Find(&depList, departmentIDs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]faculties.Faculty)
	for _, depSchema := range depList {
		if depSchema.Faculty == nil {
			continue
		}

		result[depSchema.FacultyID] = *schema.FacultyFromSchema(depSchema.Faculty)
	}

	return result, nil
}

// MapFacultiesByDepartments
func (r *Repository) MapFacultiesByCabinets(ctx context.Context, cabinetIDs uuid.UUIDs) (map[uuid.UUID]faculties.Faculty, error) {
	var cabinetList []schema.Cabinet

	err := r.client.WithContext(ctx).Preload("Faculty").Find(&cabinetList, cabinetIDs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]faculties.Faculty)
	for _, cabinetSchema := range cabinetList {
		if cabinetSchema.Faculty == nil {
			continue
		}

		result[cabinetSchema.FacultyID] = *schema.FacultyFromSchema(cabinetSchema.Faculty)
	}

	return result, nil
}

// DeleteFaculty
func (r *Repository) DeleteFaculty(ctx context.Context, id uuid.UUID) error {
	err := r.client.WithContext(ctx).Where("id = ?", id).Delete(&schema.Faculty{}).Error
	if err != nil {
		return err
	}

	return nil
}
