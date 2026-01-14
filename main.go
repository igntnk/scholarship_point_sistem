package main

import (
	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"

	"github.com/igntnk/scholarship_point_system/config"
	"github.com/igntnk/scholarship_point_system/controllers"
	"github.com/igntnk/scholarship_point_system/repository"
	"github.com/igntnk/scholarship_point_system/service"
	"github.com/igntnk/scholarship_point_system/web"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"os/signal"
	"syscall"

	"context"
	"github.com/rs/zerolog"
	"os"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	mainCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.Get(logger)

	dbConf, err := pgxpool.ParseConfig(cfg.Database.URI)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse database config")
		return
	}

	pool, err := pgxpool.NewWithConfig(mainCtx, dbConf)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
		return
	}

	db := stdlib.OpenDBFromPool(pool)

	err = goose.SetDialect("postgres")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to set postgres dialect")
		return
	}

	err = goose.Up(db, "cmd/changelog")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to migrate database")
		return
	}

	conn := trmpgx.DefaultCtxGetter.DefaultTrOrDB(mainCtx, pool)

	// Category Logic
	categoryRepo := repository.NewCategoryRepository(conn)
	orderService := service.NewCategoryService(categoryRepo)
	orderController := controllers.NewCategoryController(orderService)

	httpServer, err := web.New(logger, cfg.Server.RESTPort, orderController)
	if err != nil {
		logger.Fatal().Err(err).Send()
		return
	}

	serverErrorChan := make(chan error, 1)
	go func() {
		serverErrorChan <- httpServer.ListenAndServe()
	}()
	logger.Info().Msgf("Server started on port: %d", cfg.Server.RESTPort)

	select {
	case <-mainCtx.Done():
		logger.Info().Msg("shutting down by context down")
		err = httpServer.Shutdown(context.Background())
		if err != nil {
			logger.Err(err).Msg("error while shutting down http server")
		}
	case err = <-serverErrorChan:
		logger.Err(err).Msg("shutting down by error")
	}
}
