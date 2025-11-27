package repository

import (
	"context"
	"errors"

	edugroups "schedule-generator/internal/domain/edu_groups"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/internal/infrastructure/db/postgres/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SaveEduGroup
func (r *Repository) SaveEduGroup(ctx context.Context, d *edugroups.EduGroup) error {
	s := schema.EduGroupToSchema(d)
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

// GetEduGroup
func (r *Repository) GetEduGroup(ctx context.Context, id uuid.UUID) (*edugroups.EduGroup, error) {
	var s schema.EduGroup
	err := r.client.WithContext(ctx).Where("id = ?", id.String()).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.EduGroupFromSchema(&s), nil
}

// GetEduGroupByNumber
func (r *Repository) GetEduGroupByNumber(ctx context.Context, number string) (*edugroups.EduGroup, error) {
	var s schema.EduGroup
	err := r.client.WithContext(ctx).Where("number = ?", number).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.EduGroupFromSchema(&s), nil
}

// ListEduGroup
func (r *Repository) ListEduGroup(ctx context.Context) ([]edugroups.EduGroup, error) {
	var list []schema.EduGroup
	err := r.client.WithContext(ctx).Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]edugroups.EduGroup, len(list))
	for i, v := range list {
		result[i] = *schema.EduGroupFromSchema(&v)
	}

	return result, nil
}

// DeleteEduGroup
func (r *Repository) DeleteEduGroup(ctx context.Context, id uuid.UUID) error {
	err := r.client.WithContext(ctx).Where("id = ?", id).Delete(&schema.EduGroup{}).Error
	if err != nil {
		return err
	}

	return nil
}
