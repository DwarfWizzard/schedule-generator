package token

import (
	"context"
	"errors"
	"fmt"
	"schedule-generator/internal/application/services"
	"time"

	"github.com/golang-jwt/jwt/v5"
	j "github.com/golang-jwt/jwt/v5"
)

type jwtTokenService struct {
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewTokenService(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *jwtTokenService {
	return &jwtTokenService{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (s *jwtTokenService) GenerateToken(_ context.Context, claims *services.TokenClaims) (services.TokenPair, error) {
	c := FromClaims(claims)

	return s.generateTokenPair(&c)
}

func (s *jwtTokenService) ParseAccessToken(_ context.Context, access string) (services.TokenClaims, error) {
	token, err := j.ParseWithClaims(access, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.accessSecret), nil
	})
	if err != nil {
		return services.TokenClaims{}, fmt.Errorf("parse token error: %w", err)
	}

	claims, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return services.TokenClaims{}, errors.New("invalid token")
	}

	return ToClaims(claims), nil
}
func (s *jwtTokenService) ParseRefreshToken(_ context.Context, refresh string) (services.TokenClaims, error) {
	token, err := j.ParseWithClaims(refresh, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.refreshSecret), nil
	})
	if err != nil {
		return services.TokenClaims{}, fmt.Errorf("parse token error: %w", err)
	}

	claims, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return services.TokenClaims{}, errors.New("invalid token")
	}

	return ToClaims(claims), nil
}

func (s *jwtTokenService) generateTokenPair(claims *claims) (services.TokenPair, error) {
	now := time.Now()
	access, err := s.generateAccessToken(now, claims)
	if err != nil {
		return services.TokenPair{}, fmt.Errorf("generate access token error: %w", err)
	}

	refresh, err := s.generateRefreshToken(now, claims)
	if err != nil {
		return services.TokenPair{}, fmt.Errorf("generate refresh token error: %w", err)
	}

	return services.TokenPair{
		Access:     access,
		Refresh:    refresh,
		AccessTTL:  s.accessTTL,
		RefreshTTL: s.refreshTTL,
	}, nil
}

func (s *jwtTokenService) generateAccessToken(t time.Time, claims *claims) (string, error) {
	claims.RegisteredClaims = j.RegisteredClaims{
		IssuedAt:  j.NewNumericDate(t),
		ExpiresAt: j.NewNumericDate(t.Add(s.accessTTL)),
	}

	return j.NewWithClaims(j.SigningMethodHS256, claims).SignedString([]byte(s.accessSecret))
}

func (s *jwtTokenService) generateRefreshToken(t time.Time, claims *claims) (string, error) {
	claims.RegisteredClaims = j.RegisteredClaims{
		IssuedAt:  j.NewNumericDate(t),
		ExpiresAt: j.NewNumericDate(t.Add(s.refreshTTL)),
	}

	return j.NewWithClaims(j.SigningMethodHS256, claims).SignedString([]byte(s.refreshSecret))
}
