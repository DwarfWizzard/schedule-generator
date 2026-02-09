package users

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Role int8

const (
	RoleDeputyDean Role = iota
	RoleAdmin
)

var roleNames = []string{
	"admin",
	"deputy dean",
}

func (r Role) String() string {
	return roleNames[r]
}

func NewRole(r int8) (Role, error) {
	if int(r) < 0 || int(r) >= len(roleNames) {
		return 0, errors.New("unknown role")
	}

	return Role(r), nil
}

type User struct {
	ID        uuid.UUID
	Username  string
	Role      Role
	FacultyID *uuid.UUID
	PwdHash   string
}

func NewUser(username string, role Role, facultyID *uuid.UUID, pwdHash string) (*User, error) {
	if len(username) == 0 {
		return nil, errors.New("empty username")
	}

	if len(pwdHash) == 0 {
		return nil, errors.New("empty password")
	}

	u := User{
		ID:        uuid.New(),
		Username:  username,
		PwdHash:   pwdHash,
		Role:      role,
		FacultyID: facultyID,
	}

	if u.Role != RoleAdmin && u.FacultyID == nil {
		return nil, fmt.Errorf("faculty can not be empty for role %s", role.String())
	} else if u.Role == RoleAdmin {
		u.FacultyID = nil
	}

	return &u, nil
}
