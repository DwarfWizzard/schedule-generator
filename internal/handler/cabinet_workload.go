package handler

import (
	"context"
	"net/http"
	"schedule-generator/internal/application/usecases"
	"schedule-generator/internal/domain/users"

	"github.com/labstack/echo/v4"
)

type CabinetWorkloadUsecase interface {
	GetCabinetWorkload(ctx context.Context, user *users.User) (*usecases.CabinetWorkloadOutput, error)
}

func (h *Handler) GetCabinetWorkload(c echo.Context) error {
	user, err := ExtractUserFromClaims(c)
	if err != nil {
		return ErrUnauthorized
	}

	out, err := h.cabinetWorkload.GetCabinetWorkload(c.Request().Context(), user)
	if err != nil {
		h.logger.Error("GetCabinetWorkload error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, out).Send(c)

}
