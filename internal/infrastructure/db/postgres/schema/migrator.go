package schema

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Migrator struct {
	client *gorm.DB
}

func NewMigrator(client *gorm.DB) *Migrator {
	return &Migrator{
		client: client,
	}
}

func (m *Migrator) Migrate(ctx context.Context) error {
	tx := m.client.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin tx error: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil || tx.Error != nil {
			tx.Rollback()
		}
	}()

	err := tx.AutoMigrate(
		&Faculty{},
		&Department{},
		&EduDirection{},
		&EduGroup{},
		&Teacher{},
		&Schedule{},
		&ScheduleItem{},
	)

	if err != nil {
		return fmt.Errorf("make auto migration error: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit tx error: %w", err)
	}

	return nil
}
