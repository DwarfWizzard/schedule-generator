package schema

import (
	"schedule-generator/internal/domain/users"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `gorm:"column:id;type:string;primaryKey"`
	Name      string     `gorm:"column:name;not_null"`
	Username  string     `gorm:"column:username;unique;not null"`
	Role      int8       `gorm:"column:role;not null"`
	FacultyID *uuid.UUID `gorm:"column:faculty_id;type:string;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Faculty   *Faculty   `gorm:"foreignKey:faculty_id"`
	PwdHash   string     `gorm:"column:pwd_hash;not null"`
}

// UserToSchema
func UserToSchema(u *users.User) *User {
	return &User{
		ID:        u.ID,
		Name:      u.Name,
		Username:  u.Username,
		Role:      int8(u.Role),
		FacultyID: u.FacultyID,
		PwdHash:   u.PwdHash,
	}
}

// UserFromSchema
func UserFromSchema(scheme *User) *users.User {
	return &users.User{
		ID:        scheme.ID,
		Name:      scheme.Name,
		Username:  scheme.Username,
		Role:      users.Role(scheme.Role),
		FacultyID: scheme.FacultyID,
		PwdHash:   scheme.PwdHash,
	}
}
