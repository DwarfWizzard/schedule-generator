package services

import (
	"context"
	"schedule-generator/internal/domain/users"
	"time"

	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID uuid.UUID
	Name   string
	Role   users.Role
}

type TokenPair struct {
	Access     string
	Refresh    string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type TokenService interface {
	GenerateToken(ctx context.Context, claims *TokenClaims) (TokenPair, error)
	ParseAccessToken(ctx context.Context, access string) (TokenClaims, error)
	ParseRefreshToken(ctx context.Context, refresh string) (TokenClaims, error)
}
