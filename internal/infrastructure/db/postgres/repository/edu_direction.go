package repository

import (
	"context"
	"errors"

	edudirections "schedule-generator/internal/domain/edu_directions"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/internal/infrastructure/db/postgres/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SaveEduDirection
func (r *Repository) SaveEduDirection(ctx context.Context, d *edudirections.EduDirection) error {
	s := schema.EduDirectionToSchema(d)
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

// GetEduDirection
func (r *Repository) GetEduDirection(ctx context.Context, id uuid.UUID) (*edudirections.EduDirection, error) {
	var s schema.EduDirection
	err := r.client.WithContext(ctx).Where("id = ?", id.String()).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.EduDirectionFromSchema(&s), nil
}

// ListEduDirection
func (r *Repository) ListEduDirection(ctx context.Context) ([]edudirections.EduDirection, error) {
	var list []schema.EduDirection
	err := r.client.WithContext(ctx).Order("name ASC").Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]edudirections.EduDirection, len(list))
	for i, v := range list {
		result[i] = *schema.EduDirectionFromSchema(&v)
	}

	return result, nil
}

// MapEduDirectionByEduPlans
func (r *Repository) MapEduDirectionByEduPlans(ctx context.Context, plansIDs uuid.UUIDs) (map[uuid.UUID]edudirections.EduDirection, error) {
	var planList []schema.EduPlan

	err := r.client.WithContext(ctx).Preload("Direction").Find(&planList, plansIDs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]edudirections.EduDirection)
	for _, planSchema := range planList {
		if planSchema.Direction == nil {
			continue
		}

		result[planSchema.DirectionID] = *schema.EduDirectionFromSchema(planSchema.Direction)
	}

	return result, nil
}

// DeleteEduDirection
func (r *Repository) DeleteEduDirection(ctx context.Context, id uuid.UUID) error {
	err := r.client.WithContext(ctx).Where("id = ?", id).Delete(&schema.EduDirection{}).Error
	if err != nil {
		return err
	}

	return nil
}
