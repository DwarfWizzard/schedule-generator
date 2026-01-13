package repository

import (
	"context"
	"errors"

	"schedule-generator/internal/domain/cabinets"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/internal/infrastructure/db/postgres/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repository) SaveCabinet(ctx context.Context, c *cabinets.Cabinet) error {
	s := schema.CabinetToSchema(c)

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

func (r *Repository) GetCabinet(ctx context.Context, id uuid.UUID) (*cabinets.Cabinet, error) {
	var s schema.Cabinet
	err := r.client.WithContext(ctx).Where("id = ?", id.String()).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.CabinetFromSchema(&s), nil
}

func (r *Repository) ListCabinet(ctx context.Context) ([]cabinets.Cabinet, error) {
	var list []schema.Cabinet
	err := r.client.WithContext(ctx).Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]cabinets.Cabinet, len(list))
	for i, v := range list {
		result[i] = *schema.CabinetFromSchema(&v)
	}

	return result, nil
}

func (r *Repository) DeleteCabinet(ctx context.Context, id uuid.UUID) error {
	err := r.client.WithContext(ctx).Where("id = ?", id).Delete(&schema.Cabinet{}).Error
	if err != nil {
		return err
	}

	return nil
}
