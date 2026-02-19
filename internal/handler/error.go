package handler

import (
	"log/slog"
	"net/http"
	"schedule-generator/pkg/execerror"

	"github.com/labstack/echo/v4"
)

// API errors
var (
	ErrNotParsable         = echo.NewHTTPError(http.StatusBadRequest, "Format error. Your request is not parsable.")
	ErrProcessorNotFound   = echo.NewHTTPError(http.StatusNotFound, "The processor was not found for the offer you attempted.")
	ErrOfferNotFound       = echo.NewHTTPError(http.StatusBadRequest, "Offer not found.")
	ErrProcessingConflict  = echo.NewHTTPError(http.StatusBadRequest, "Processing conflict.")
	ErrPermissionDenied    = echo.NewHTTPError(http.StatusForbidden, "The provided credentials do not have access to the requested data.")
	ErrMimeTypeError       = echo.NewHTTPError(http.StatusUnsupportedMediaType, "Expected content type is application/json.")
	ErrInternalServerError = echo.NewHTTPError(http.StatusInternalServerError, "Internal server error.")
	ErrInvalidInput        = echo.NewHTTPError(http.StatusBadRequest, "Invalid input.")
	ErrServiceUnavailable  = echo.NewHTTPError(http.StatusServiceUnavailable, "Service unavailable. Try again later.")
	ErrNotImplemented      = echo.NewHTTPError(http.StatusNotImplemented, "Operation not supproted")
	ErrUnsupportedFormat   = echo.NewHTTPError(http.StatusNotImplemented, "Format not supproted")
	ErrUnauthorized        = echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized request")
	ErrInvalidAuthHeader   = echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header")
)

var execErrorToApiError = map[execerror.ExecErrorType]*echo.HTTPError{
	execerror.TypeInternal:           ErrInternalServerError,
	execerror.TypeInvalidInput:       ErrInvalidInput,
	execerror.TypeProcessingConflict: ErrProcessingConflict,
	execerror.TypeUnimpemented:       ErrNotImplemented,
	execerror.TypeForbbiden:          ErrPermissionDenied,
}

// HttpErrorHandler handles errors that occur while running the API
func NewHttpErrorHandler(logger *slog.Logger) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		var errCode int
		var errMessage string

		extras := make(map[string][]string)

		switch e := err.(type) {
		case *execerror.ExecError:
			desc, ok := execErrorToApiError[e.Type]
			if ok {
				errCode = desc.Code
				errMessage = desc.Message.(string)
			} else {
				errCode = http.StatusInternalServerError
				errMessage = "Internal server error"
			}

			extras = e.Details

			if e.Cause != nil {
				if extras == nil {
					extras = make(map[string][]string)
				}

				extras["cause"] = append(extras["cause"], e.Cause.Error())
			}
		default:
			errCode = http.StatusInternalServerError
			errMessage = e.Error()
		}

		if c.Response().Committed {
			return
		}

		resp := &ResponseWrapper{
			Status:  errCode,
			Message: &errMessage,
			Errors:  extras,
		}

		logger.Debug("aaaa", "err", resp)

		respErr := resp.Send(c)

		if respErr != nil {
			logger.Warn("send response error", "error", respErr)
		}
	}
}
