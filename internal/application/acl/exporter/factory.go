package exporter

import (
	"log/slog"
	"schedule-generator/internal/domain/cabinets"
	"schedule-generator/internal/domain/departments"
	edugroups "schedule-generator/internal/domain/edu_groups"
	"schedule-generator/internal/domain/teachers"
)

type ExporterRepository interface {
	teachers.Repository
	edugroups.Repository
	departments.Repository
	cabinets.Repository
}

type exporterFactory struct {
	repo   ExporterRepository
	opt    *Options
	logger *slog.Logger
}

func NewExporterFactory(repo ExporterRepository, logger *slog.Logger, opts ...Option) Factory {
	o := &Options{
		CsvDelimeter: DefaultCsvDelimeter,
	}

	for _, setter := range opts {
		setter(o)
	}

	return &exporterFactory{
		opt:    o,
		repo:   repo,
		logger: logger,
	}
}

func (f *exporterFactory) ByFormat(format string) (Exporter, error) {
	switch format {
	case "csv":
		return &csvExporter{repo: f.repo, logger: f.logger.With("exporter", "csv"), delimeter: f.opt.CsvDelimeter}, nil
	default:
		return nil, ErrUnknownFormat
	}
}
