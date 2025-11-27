package exporter

import (
	"context"
	"errors"
	"io"
	"schedule-generator/internal/domain/schedules"
)

var ErrUnknownFormat = errors.New("unimplemented export format")

type Exporter interface {
	Export(ctx context.Context, schedule *schedules.Schedule, dst io.Writer) error
}

type Factory interface {
	ByFormat(format string) (Exporter, error)
}
