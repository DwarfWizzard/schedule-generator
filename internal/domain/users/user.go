package users

import "github.com/google/uuid"

type Role string

type User struct {
	ID        uuid.UUID
	Username  string
	Role      Role
	FacultyID *uuid.UUID
	PwdHash   string
}
