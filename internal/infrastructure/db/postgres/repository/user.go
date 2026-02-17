package repository

import (
	"context"
	"errors"
	"schedule-generator/internal/domain/users"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/internal/infrastructure/db/postgres/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SaveUser
func (r *Repository) SaveUser(ctx context.Context, user *users.User) error {
	s := schema.UserToSchema(user)

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

// GetUser
func (r *Repository) GetUser(ctx context.Context, userID uuid.UUID) (*users.User, error) {
	var s schema.User
	err := r.client.WithContext(ctx).Where("id = ?", userID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.UserFromSchema(&s), nil
}

// GetUserByUsername
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*users.User, error) {
	var s schema.User
	err := r.client.WithContext(ctx).Where("username = ?", username).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.UserFromSchema(&s), nil
}

// ListUser
func (r *Repository) ListUser(ctx context.Context) ([]users.User, error) {
	var list []schema.User

	err := r.client.WithContext(ctx).Order("id").Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]users.User, len(list))
	for i, v := range list {
		result[i] = *schema.UserFromSchema(&v)
	}

	return result, nil
}

// DeleteUser
func (r *Repository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	err := r.client.WithContext(ctx).Where("id = ?", userID).Delete(&schema.User{}).Error
	if err != nil {
		return err
	}

	return nil
}
