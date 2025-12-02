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
		&Module{},
		&EduPlan{},
		&Schedule{},
		&ScheduleItem{},
	)

	if err != nil {
		return fmt.Errorf("make auto migration error: %w", err)
	}

	err = tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_cycled_schedule_item_weekday_lesson_subgroup_weektype ON schedule_items (schedule_id, weekday, lesson_number, subgroup, weektype) WHERE date IS NULL").Error
	if err != nil {
		return fmt.Errorf("create unique index idx_cycled_schedule_item_weekday_lesson_subgroup_weektype error: %w", err)
	}

	err = tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_calendar_schedule_item_lesson_subgroup_date ON schedule_items (schedule_id, lesson_number, subgroup, date) WHERE weektype IS NULL").Error
	if err != nil {
		return fmt.Errorf("create unique index idx_calendar_schedule_item_lesson_subgroup_date error: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit tx error: %w", err)
	}

	return nil
}
