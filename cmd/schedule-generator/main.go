package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"schedule-generator/internal/application/acl/exporter"
	"schedule-generator/internal/application/services"
	"schedule-generator/internal/application/usecases"
	"schedule-generator/internal/handler"
	"schedule-generator/internal/infrastructure/db/postgres/repository"
	"schedule-generator/internal/infrastructure/db/postgres/schema"
	"schedule-generator/internal/infrastructure/services/pwd"
	"schedule-generator/internal/infrastructure/services/token"
	"schedule-generator/pkg/pggorm"

	"github.com/ardanlabs/conf/v3"
)

const appPrefix = ""

type Config struct {
	PostgresConnectionUrl string        `conf:"required,notzero"`
	AccessTokenSecret     string        `conf:"required,mask,notzero"`
	RefreshTokenSecret    string        `conf:"required,mask,notzero"`
	AccessTTL             time.Duration `conf:"default:15m"`
	RefreshTTL            time.Duration `conf:"default:24h"`
	PasswordSalt          string        `conf:"required,mask,notzero"`
}

func main() {
	logger := slog.Default()

	var cfg Config
	help, err := conf.Parse(appPrefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
		} else {
			logger.Error("Invalid configuration")
		}

		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := pggorm.NewDB(cfg.PostgresConnectionUrl)
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
	exp := exporter.NewExporterFactory(repo, logger, exporter.CsvDelimeter(';'))
	authSvc := services.NewAuthorizationService(repo)
	tokenSvc := token.NewTokenService(
		cfg.AccessTokenSecret,
		cfg.RefreshTokenSecret,
		cfg.AccessTTL,
		cfg.RefreshTTL,
	)
	pwdSvc := pwd.NewPasswordService(cfg.PasswordSalt)

	h := handler.NewHandler(
		usecases.NewDepartmentUsecase(authSvc, repo, logger),
		usecases.NewEduDirectionUsecase(authSvc, repo, logger),
		usecases.NewEduGroupUsecase(authSvc, repo, logger),
		usecases.NewEduPlanUsecase(authSvc, repo, logger),
		usecases.NewFacultyUsecase(authSvc, repo, logger),
		usecases.NewScheduleUsecase(authSvc, repo, exp, logger),
		usecases.NewTeacherUsecase(authSvc, repo, logger),
		usecases.NewCabinetUsecase(authSvc, repo, logger),
		usecases.NewUserUsecase(authSvc, pwdSvc, tokenSvc, repo, logger),
		logger,
	)

	router := h.InitRouter()

	var wg sync.WaitGroup

	wg.Go(func() {
		defer cancel()
		if err := router.Start(":" + os.Getenv("API_PORT")); err != nil {
			logger.Error("Start router error", "error", err)
		}
	})

	// Stop services without context handling support
	wg.Go(func() {
		<-ctx.Done()
		logger.Info("Closing router")
		if err := router.Close(); err != nil {
			logger.Warn("API web server closing error.", "error", err)
		}

		logger.Info("Closing postgres connection")
		if err := db.Close(); err != nil {
			logger.Warn("Close postgres connection error", "error", err)
		}
	})

	go func() {
		defer cancel()
		chSig := make(chan os.Signal, 1)
		signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)
		sig := <-chSig
		logger.Info(fmt.Sprintf("OS signal received: %s", sig))
	}()

	wg.Wait()

	logger.Info("Service finished")
}
