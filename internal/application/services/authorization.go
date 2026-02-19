package services

import (
	"context"
	"fmt"
	"schedule-generator/internal/domain/departments"
	edudirections "schedule-generator/internal/domain/edu_directions"
	edugroups "schedule-generator/internal/domain/edu_groups"
	eduplans "schedule-generator/internal/domain/edu_plans"
	"schedule-generator/internal/domain/faculties"
	"schedule-generator/internal/domain/schedules"
	"schedule-generator/internal/domain/teachers"
	"schedule-generator/internal/domain/users"

	"github.com/google/uuid"
)

type AuthorizationServiceRepository interface {
	GetEduDirectionFacultyID(ctx context.Context, directionID uuid.UUID) (uuid.UUID, error)
	GetEduGroupFacultyID(ctx context.Context, groupID uuid.UUID) (uuid.UUID, error)
	GetEduPlanFacultyID(ctx context.Context, planID uuid.UUID) (uuid.UUID, error)
	GetScheduleFacultyID(ctx context.Context, scheduleID uuid.UUID) (uuid.UUID, error)
	GetTeacherFacultyID(ctx context.Context, teacherID uuid.UUID) (uuid.UUID, error)
}

type AuthorizationService struct {
	svc  users.AuthorizationService
	repo AuthorizationServiceRepository
}

func NewAuthorizationService(repo AuthorizationServiceRepository) *AuthorizationService {
	return &AuthorizationService{
		svc:  users.AuthorizationService{},
		repo: repo,
	}
}

func (a *AuthorizationService) IsAdmin(user *users.User) bool {
	return a.svc.IsAdmin(user)
}

func (a *AuthorizationService) HaveAccessToFaculty(ctx context.Context, faculty *faculties.Faculty, user *users.User) (bool, error) {
	return a.svc.HaveAccessToFaculty(user, faculty.ID), nil
}

func (a *AuthorizationService) HaveAccessToDepartment(ctx context.Context, department *departments.Department, user *users.User) (bool, error) {
	return a.svc.HaveAccessToFaculty(user, department.FacultyID), nil
}

func (a *AuthorizationService) HaveAccessToEduDirection(ctx context.Context, direction *edudirections.EduDirection, user *users.User) (bool, error) {
	facultyID, err := a.repo.GetEduDirectionFacultyID(ctx, direction.ID)
	if err != nil {
		return false, fmt.Errorf("get direction faculty id error: %w", err)
	}

	return a.svc.HaveAccessToFaculty(user, facultyID), nil
}

func (a *AuthorizationService) HaveAccessToEduGroup(ctx context.Context, group *edugroups.EduGroup, user *users.User) (bool, error) {
	facultyID, err := a.repo.GetEduGroupFacultyID(ctx, group.ID)
	if err != nil {
		return false, fmt.Errorf("get edu group faculty id error: %w", err)
	}

	return a.svc.HaveAccessToFaculty(user, facultyID), nil
}

func (a *AuthorizationService) HaveAccessToEduPlan(ctx context.Context, plan *eduplans.EduPlan, user *users.User) (bool, error) {
	facultyID, err := a.repo.GetEduPlanFacultyID(ctx, plan.ID)
	if err != nil {
		return false, fmt.Errorf("get edu plan faculty id error: %w", err)
	}

	return a.svc.HaveAccessToFaculty(user, facultyID), nil
}

func (a *AuthorizationService) HaveAccessToSchedule(ctx context.Context, schedule *schedules.Schedule, user *users.User) (bool, error) {
	facultyID, err := a.repo.GetScheduleFacultyID(ctx, schedule.ID)
	if err != nil {
		return false, fmt.Errorf("get schedule faculty id error: %w", err)
	}

	return a.svc.HaveAccessToFaculty(user, facultyID), nil
}

func (a *AuthorizationService) HaveAccessToTeacher(ctx context.Context, teacher *teachers.Teacher, user *users.User) (bool, error) {
	facultyID, err := a.repo.GetTeacherFacultyID(ctx, teacher.ID)
	if err != nil {
		return false, fmt.Errorf("get teacher faculty id error: %w", err)
	}

	return a.svc.HaveAccessToFaculty(user, facultyID), nil
}
