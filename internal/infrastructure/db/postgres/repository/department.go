package repository

import (
	"context"
	"errors"

	"schedule-generator/internal/domain/departments"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/internal/infrastructure/db/postgres/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SaveDepartment
func (r *Repository) SaveDepartment(ctx context.Context, d *departments.Department) error {
	s := schema.DepartmentToSchema(d)
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

// GetDepartment
func (r *Repository) GetDepartment(ctx context.Context, id uuid.UUID) (*departments.Department, error) {
	var s schema.Department
	err := r.client.WithContext(ctx).Where("id = ?", id.String()).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.DepartmentFromSchema(&s), nil
}

// ListDepartment
func (r *Repository) ListDepartment(ctx context.Context) ([]departments.Department, error) {
	var list []schema.Department
	err := r.client.WithContext(ctx).Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]departments.Department, len(list))
	for i, v := range list {
		result[i] = *schema.DepartmentFromSchema(&v)
	}

	return result, nil
}

// DeleteDepartment
func (r *Repository) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	err := r.client.WithContext(ctx).Where("id = ?", id).Delete(&schema.Department{}).Error
	if err != nil {
		return err
	}

	return nil
}
