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
)

var cfgPath string

func init() {
	flag.StringVar(&cfgPath, "config", "config.yaml", "config file path")
	flag.Parse()
}

func setupZap() *zap.Logger { return zap.L() }

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	var cfg config.Config
	if err := config.Load(cfgPath, &cfg); err != nil {
		logger.Fatal("could not load config", zap.Error(err))
	}

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

	go func() {
		err := srv.ListenAndServe()
		zap.L().Error("http server closed", zap.Error(err))
		cancel()
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
