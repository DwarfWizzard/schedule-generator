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

func (m *Migrator) Migrate(ctx context.Context, migrationVersion int) error {
	tx := m.client.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin tx error: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil || tx.Error != nil {
			tx.Rollback()
		}
	}()

	switch migrationVersion {
	case 1:
		if tx.Migrator().HasTable(&EduDirection{}) &&
			tx.Migrator().HasTable(&EduPlan{}) &&
			tx.Migrator().HasColumn(&EduDirection{}, "department_id") {

			if !tx.Migrator().HasColumn(&EduPlan{}, "department_id") {
				if err := tx.Migrator().AddColumn(&EduPlan{}, "department_id"); err != nil {
					return fmt.Errorf("add department_id to edu_plans: %w", err)
				}
			}

			err := tx.Exec(`
				UPDATE edu_plans 
				SET department_id = edu_directions.department_id 
				FROM edu_directions 
				WHERE edu_plans.direction_id = edu_directions.id 
				AND edu_plans.department_id IS NULL
			`).Error
			if err != nil {
				return fmt.Errorf("copy department_id from edu_directions: %w", err)
			}

			var uncopiedCount int64
			tx.Table("edu_plans p").
				Joins("LEFT JOIN edu_directions d ON p.direction_id = d.id").
				Where("p.department_id IS NULL AND d.department_id IS NOT NULL").
				Count(&uncopiedCount)
			if uncopiedCount > 0 {
				return fmt.Errorf("не перенесено записей EduPlan: %d", uncopiedCount)
			}

			if err := tx.Migrator().DropColumn(&EduDirection{}, "department_id"); err != nil {
				return fmt.Errorf("drop department_id from edu_directions: %w", err)
			}

			err = tx.Exec("DROP INDEX IF EXISTS edu_plan_direction_profile_year_unique").Error
			if err != nil {
				return fmt.Errorf("drop old index: %w", err)
			}

			err = tx.Exec(`
				CREATE UNIQUE INDEX IF NOT EXISTS edu_plan_direction_department_profile_year_unique 
				ON edu_plans (direction_id, department_id, profile, year)
			`).Error
			if err != nil {
				return fmt.Errorf("create new unique index: %w", err)
			}

			if err := tx.Exec("ALTER TABLE edu_plans ALTER COLUMN department_id SET NOT NULL").Error; err != nil {
				return fmt.Errorf("add NOT NULL on department_id: %w", err)
			}
		}
		fallthrough
	default:
		err := tx.AutoMigrate(
			&User{},
			&Faculty{},
			&Department{},
			&EduDirection{},
			&EduGroup{},
			&Teacher{},
			&Module{},
			&EduPlan{},
			&Schedule{},
			&ScheduleItem{},
			&Cabinet{},
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
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit tx error: %w", err)
	}

	return nil
}
