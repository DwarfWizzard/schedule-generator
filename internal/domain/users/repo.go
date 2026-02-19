package users

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	SaveUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, userID uuid.UUID) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	ListUser(ctx context.Context) ([]User, error)
}
