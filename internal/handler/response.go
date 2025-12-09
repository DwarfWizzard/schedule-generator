package handler

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ResponseWrapper wraps API response
type ResponseWrapper struct {
	Status   int                 `json:"status"`
	Response any                 `json:"response,omitempty"`
	Message  *string             `json:"message,omitempty"`
	Errors   map[string][]string `json:"errors,omitempty"`
}

func WrapResponse(status int, payload any) *ResponseWrapper {
	return &ResponseWrapper{
		Status:   status,
		Response: payload,
	}
}

// Send returns to API client wrapped response in JSON format
func (rw *ResponseWrapper) Send(c echo.Context) error {
	if rw == nil {
		return nil
	}

	return c.JSON(rw.Status, rw)
}

// Send returns to API client wrapped
func (rw *ResponseWrapper) SendAsFile(c echo.Context, filename, format string) error {
	if rw == nil {
		return nil
	}

	buffer, ok := rw.Response.(*bytes.Buffer)
	if !ok {
		return c.NoContent(http.StatusNoContent)
	}

	switch format {
	case "csv":
		c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+filename)
		c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	default:
		return ErrUnsupportedFormat
	}

	c.Response().Status = http.StatusOK
	_, err := buffer.WriteTo(c.Response())
	if err != nil {
		return err
	}

	return nil
}
