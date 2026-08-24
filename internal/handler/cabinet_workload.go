package handler

import (
	"context"
	"net/http"
	"schedule-generator/internal/application/usecases"

	"github.com/labstack/echo/v4"
)

type CabinetWorkloadUsecase interface {
	// GetCabinetWorkload(ctx context.Context, user *users.User) (usecases.CabinetWorkloadOutput, error)
	// GetCabinetWorkload(ctx context.Context) (usecases.CabinetWorkloadOutput, error)
	GetCabinetWorkload(ctx context.Context) (*usecases.CabinetWorkloadOutput, error)
}

func (h *Handler) GetCabinetWorkload(c echo.Context) error {
	// user, err := ExtractUserFromClaims(c)
	// if err != nil {
	// 	return ErrUnauthorized
	// }

	// out, err := h.cabinetWorkload.GetCabinetWorkload(c.Request().Context(), user)
	out, err := h.cabinetWorkload.GetCabinetWorkload(c.Request().Context())
	if err != nil {
		h.logger.Error("GetCabinetWorkload error", "error", err)
		return err
	}

	return WrapResponse(http.StatusOK, out).Send(c)

}
