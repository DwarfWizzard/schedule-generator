package repository

import (
	"context"
	"errors"

	eduplans "schedule-generator/internal/domain/edu_plans"
	"schedule-generator/internal/infrastructure/db"
	"schedule-generator/internal/infrastructure/db/postgres/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SaveEduPlan
func (r *Repository) SaveEduPlan(ctx context.Context, d *eduplans.EduPlan) error {
	s := schema.EduPlanToSchema(d)
	err := r.client.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(s).Error
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

// GetEduPlan
func (r *Repository) GetEduPlan(ctx context.Context, id uuid.UUID) (*eduplans.EduPlan, error) {
	var s schema.EduPlan
	err := r.client.WithContext(ctx).Where("id = ?", id.String()).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	return schema.EduPlanFromSchema(&s), nil
}

// GetEduPlanFacultyID
func (r *Repository) GetEduPlanFacultyID(ctx context.Context, planID uuid.UUID) (uuid.UUID, error) {
	var s schema.EduPlan
	err := r.client.WithContext(ctx).Preload("Direction.Department").Preload(clause.Associations).Where("id = ?", planID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.UUID{}, db.ErrorNotFound
		}

		return uuid.UUID{}, err
	}

	return s.Direction.Department.FacultyID, nil
}

// ListEduPlan
func (r *Repository) ListEduPlan(ctx context.Context) ([]eduplans.EduPlan, error) {
	var list []schema.EduPlan
	err := r.client.WithContext(ctx).Joins("Direction").Order(`"Direction".name, direction_id, year ASC`).Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]eduplans.EduPlan, len(list))
	for i, v := range list {
		result[i] = *schema.EduPlanFromSchema(&v)
	}

	return result, nil
}

// ListEduPlan
func (r *Repository) ListEduPlanByFaculty(ctx context.Context, facultyID uuid.UUID) ([]eduplans.EduPlan, error) {
	var list []schema.EduPlan
	err := r.client.WithContext(ctx).Joins("Direction.Department").Where("Department.faculty_id = ?", facultyID).Order("direction_id, year ASC").Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, db.ErrorNotFound
		}

		return nil, err
	}

	result := make([]eduplans.EduPlan, len(list))
	for i, v := range list {
		result[i] = *schema.EduPlanFromSchema(&v)
	}

	return result, nil
}

// MapEduPlansByEduGroups
// func (r *Repository) MapEduPlansByEduGroups(ctx context.Context, groupIDs uuid.UUIDs) (map[uuid.UUID]eduplans.EduPlan, error) {
// 	var groupList []schema.EduGroup

// 	err := r.client.WithContext(ctx).Preload("EduPlan").Find(&groupList, groupIDs).Error
// 	if err != nil {
// 		return nil, err
// 	}

// 	result := make(map[uuid.UUID]eduplans.EduPlan)
// 	for _, groupSchema := range groupList {
// 		if groupSchema.EduPlan == nil {
// 			continue
// 		}

// 		result[groupSchema.EduPlanID] = *schema.EduPlanFromSchema(groupSchema.EduPlan)
// 	}

// 	return result, nil
// }

// DeleteEduPlan
func (r *Repository) DeleteEduPlan(ctx context.Context, id uuid.UUID) error {
	err := r.client.WithContext(ctx).Where("id = ?", id).Delete(&schema.EduPlan{}).Error
	if err != nil {
		return err
	}

	return nil
}
