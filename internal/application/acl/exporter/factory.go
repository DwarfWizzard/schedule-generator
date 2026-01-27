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
	logger *slog.Logger
}

func NewExporterFactory(repo ExporterRepository, logger *slog.Logger) Factory {
	return &exporterFactory{
		repo:   repo,
		logger: logger,
	}
}

func (f *exporterFactory) ByFormat(format string) (Exporter, error) {
	switch format {
	case "csv":
		return &csvExporter{repo: f.repo, logger: f.logger.With("exporter", "csv")}, nil
	default:
		return nil, ErrUnknownFormat
	}
}
