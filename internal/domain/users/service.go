package users

import (
	"github.com/google/uuid"
)

type AuthorizationService struct{}

func (s *AuthorizationService) HaveAccessToFaculty(user *User, facultyID uuid.UUID) bool {
	if user == nil {
		return false
	}

	return s.IsAdmin(user) || (user.FacultyID != nil && *user.FacultyID == facultyID)
}

func (s *AuthorizationService) IsAdmin(user *User) bool {
	if user == nil {
		return false
	}

	return user.Role == RoleAdmin
}
