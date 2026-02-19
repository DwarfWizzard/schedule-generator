package handler

import (
	"net/http"
	"schedule-generator/internal/application/usecases"
	"strings"

	"github.com/labstack/echo/v4"
)

const TokenPrefix = "Bearer "

type TokenPair struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginReponse struct {
	TokenPair
}

// Login - POST /auth/login
func (h *Handler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var rq LoginRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	tokenPair, err := h.user.UserAuthentication(ctx, usecases.UserAuthenticationInput{
		Username: rq.Username,
		Password: rq.Password,
	})
	if err != nil {
		h.logger.Error("Get user by username and password error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, LoginReponse{TokenPair{
		Access:  tokenPair.Access,
		Refresh: tokenPair.Refresh,
	}}).Send(c)
}

type RefreshRequest struct {
	RefreshToken string `json:"token"`
}

// Refresh - POST /auth/refresh?token=
func (h *Handler) Refresh(c echo.Context) error {
	ctx := c.Request().Context()

	var rq RefreshRequest
	if err := c.Bind(&rq); err != nil {
		return ErrNotParsable
	}

	tokenPair, err := h.user.RefreshUserToken(ctx, rq.RefreshToken)
	if err != nil {
		h.logger.Error("Refresh token error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, LoginReponse{TokenPair{
		Access:  tokenPair.Access,
		Refresh: tokenPair.Refresh,
	}}).Send(c)
}

// AuthorizationMiddleware
func (h *Handler) AuthorizationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, TokenPrefix) {
				return ErrInvalidAuthHeader
			}

			token := strings.TrimPrefix(authHeader, TokenPrefix)

			user, err := h.user.UserAuthorization(ctx, token)
			if err != nil {
				return ErrUnauthorized
			}

			c.Set("authorized-user", user)
			return next(c)
		}
	}
}
