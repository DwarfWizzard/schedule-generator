package token

import (
	"schedule-generator/internal/application/services"
	"schedule-generator/internal/domain/users"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type claims struct {
	j.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
	Role   int8      `json:"user_role"`
}

func FromClaims(c *services.TokenClaims) claims {
	return claims{
		UserID: c.UserID,
		Role:   int8(c.Role),
	}
}

func ToClaims(c *claims) services.TokenClaims {
	r, _ := users.NewRole(c.Role)
	return services.TokenClaims{
		UserID: c.UserID,
		Role:   r,
	}
}
