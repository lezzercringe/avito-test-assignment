package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"

	"github.com/lezzercringe/avito-test-assignment/internal/api/handlers/prs"
	"github.com/lezzercringe/avito-test-assignment/internal/api/handlers/teams"
	"github.com/lezzercringe/avito-test-assignment/internal/api/handlers/users"
	"github.com/lezzercringe/avito-test-assignment/internal/api/router"
	"github.com/lezzercringe/avito-test-assignment/internal/config"
	"github.com/lezzercringe/avito-test-assignment/internal/platform/postgres"
	"github.com/lezzercringe/avito-test-assignment/internal/usecases"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var cfgPath string

func init() {
	flag.StringVar(&cfgPath, "config", "config.yaml", "config file path")
	flag.Parse()

	globalLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(globalLogger)
}

func setupZap(cfg config.Config) *zap.Logger {
	levels := map[string]zapcore.Level{
		"debug": zap.DebugLevel,
		"warn":  zap.WarnLevel,
		"info":  zap.InfoLevel,
		"error": zap.ErrorLevel,
		"fatal": zap.FatalLevel,
	}

	lvl, ok := levels[cfg.LogLevel]
	if !ok {
		lvl = zap.InfoLevel
	}

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level.SetLevel(lvl)

	logger, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

func main() {
	var cfg config.Config
	if err := config.Load(cfgPath, &cfg); err != nil {
		zap.L().Fatal("could not load config", zap.Error(err))
	}

	logger := setupZap(cfg)
	defer logger.Sync()

	pool, err := postgres.SetupPool(cfg.Postgres)
	if err != nil {
		logger.Fatal("could not setup postgres pool", zap.Error(err))
	}

	userRepo := postgres.NewUserRepository(pool)
	teamRepo := postgres.NewTeamRepository(pool)
	prRepo := postgres.NewPRRepository(pool)
	txManager := postgres.NewTxManager(pool)

	rpicker := usecases.NewRandomReviewerPicker(userRepo)

	prService := usecases.NewPullRequestService(rpicker, prRepo, teamRepo)
	teamService := usecases.NewTeamService(txManager, teamRepo, userRepo)
	userService := usecases.NewUserService(txManager, userRepo, teamRepo, prRepo, rpicker)

	r := router.New(
		logger, cfg,
		prs.NewHandler(prService),
		teams.NewHandler(teamService),
		users.NewHandler(userService),
	)

	srv := http.Server{
		Addr:    cfg.ServeAddr,
		Handler: r,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	logger.Info("starting http server", zap.String("addr", cfg.ServeAddr))

	go func() {
		err := srv.ListenAndServe()
		logger.Error("http server closed", zap.Error(err))
		cancel()
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
