package usecases

import (
	"schedule-generator/internal/domain/users"

	"github.com/google/uuid"
)

type AuthorizationService interface {
	HaveAccessToFaculty(user *users.User, facultyID uuid.UUID) bool
	IsAdmin(user *users.User) bool
}
