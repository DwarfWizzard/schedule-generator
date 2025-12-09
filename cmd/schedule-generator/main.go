package main

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"sync"

	"schedule-generator/internal/application/acl/exporter"
	"schedule-generator/internal/application/usecases"
	"schedule-generator/internal/handler"
	"schedule-generator/internal/infrastructure/db/postgres/repository"
	"schedule-generator/internal/infrastructure/db/postgres/schema"
	"schedule-generator/pkg/pggorm"
)

func main() {
	logger := slog.Default()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgConnUrl := os.Getenv("POSTGRES_CONNECTION_URL")
	if _, err := url.Parse(pgConnUrl); err != nil {
		logger.Error("Invalid POSTGRES_CONNECTION_URL")
		os.Exit(1)
	}

	db, err := pggorm.NewDB(pgConnUrl)
	if err != nil {
		logger.Error("Connect to postgre error", "error", err)
		os.Exit(1)
	}

	migrator := schema.NewMigrator(db.DB())
	if err := migrator.Migrate(ctx); err != nil {
		logger.Error("Migrate error", "error", err)
		os.Exit(1)
	}

	repo := repository.NewPostgresRepository(db.DB())
	exp := exporter.NewExporterFactory(repo, logger)

	h := handler.NewHandler(
		usecases.NewDepartmentUsecase(repo, logger),
		usecases.NewEduDirectionUsecase(repo, logger),
		usecases.NewEduGroupUsecase(repo, logger),
		usecases.NewEduPlanUsecase(repo, logger),
		usecases.NewFacultyUsecase(repo, logger),
		usecases.NewScheduleUsecase(repo, exp, logger),
		usecases.NewTeacherUsecase(repo, logger),
		logger,
	)

	router := h.InitRouter()

	var wg sync.WaitGroup

	wg.Go(func() {
		defer cancel()
		if err := router.Start(":8080"); err != nil {
			logger.Error("Start router error", "error", err)
		}
	})

	// Stop services without context handling support
	go func() {
		<-ctx.Done()
		if err := router.Close(); err != nil {
			logger.Warn("API web server closing error.", "error", err)
		}
	}()

	wg.Wait()

	logger.Info("Closing postgres connection")
	if err := db.Close(); err != nil {
		logger.Error("Close postgres connection error", "error", err)
	}

	logger.Info("Service finished")
}
