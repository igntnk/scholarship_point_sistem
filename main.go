package main

import (
	"context"
	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/igntnk/scholarship_point_system/config"
	"github.com/igntnk/scholarship_point_system/controllers"
	"github.com/igntnk/scholarship_point_system/jwk"
	"github.com/igntnk/scholarship_point_system/middleware"
	"github.com/igntnk/scholarship_point_system/repository"
	"github.com/igntnk/scholarship_point_system/service"
	"github.com/igntnk/scholarship_point_system/web"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"syscall"
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

	privateKey, err := os.ReadFile(cfg.Secure.JWTPrivateKeyPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read jwt private key")
		return
	}

	jwkey := jwk.CreateJWK(privateKey)

	permissionRepo := repository.NewPermissionRepository(conn)
	permissionService := service.NewPermissionService(permissionRepo)
	m := middleware.NewMiddleware(permissionService, jwkey)
	permissionController := controllers.NewPermissionController(permissionService, m)

	categoryRepo := repository.NewCategoryRepository(conn)
	orderService := service.NewCategoryService(categoryRepo)
	orderController := controllers.NewCategoryController(orderService, m)

	passwordManager := service.NewPasswordManager(cfg.Secure.PasswordBcryptCost)

	userRepo := repository.NewUserRepository(conn)
	userService := service.NewUserService(userRepo, passwordManager)
	userController := controllers.NewUserController(userService, m)

	authRepo := repository.NewAuthRepository(conn)
	authService := service.NewAuthService(
		authRepo,
		userRepo,
		passwordManager,
		permissionRepo,
		jwkey,
		cfg.Secure.AccessTokenDuration,
		cfg.Secure.RefreshTokenDuration,
		cfg.Secure.AdminGroupName,
	)
	authController := controllers.NewAuthController(authService, m)

	achievementRepo := repository.NewAchievementRepository(conn)
	achievementService := service.NewAchievementService(achievementRepo, userRepo)
	achievementController := controllers.NewAchievementController(achievementService, m)

	httpServer, err := web.New(
		logger,
		cfg.Server.RESTPort,
		cfg.CORS,
		orderController,
		permissionController,
		userController,
		authController,
		achievementController,
	)
	if err != nil {
		logger.Fatal().Err(err).Send()
		return
	}

	adminUUID, err := authService.ActualizeAdmin(mainCtx, cfg.Secure.AdminEmail, cfg.Secure.AdminPassword)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to reload admin data")
		return
	}

	err = permissionService.ActualizeResources(mainCtx, httpServer.GetRoutes())
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to reload active resources")
		return
	}

	err = permissionService.ActualizeAdminGroupAndRole(
		mainCtx,
		adminUUID,
		cfg.Secure.AdminGroupName,
		cfg.Secure.AdminRoleName,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to reload active admin group")
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
