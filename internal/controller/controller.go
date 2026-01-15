package controller

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/pupload/pupload/internal/controller/config"
	flow "github.com/pupload/pupload/internal/controller/flows/service"
	controllerserver "github.com/pupload/pupload/internal/controller/server"
	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
)

func RunWithConfig(ctx context.Context, cfg *config.ControllerSettings) error {

	logging.Init(logging.Config{
		AppName: "pupload.controller",
		Level:   slog.LevelInfo,
		Format:  "text",
	})

	log := logging.Root()

	if err := telemetry.Init(cfg.Telemetry, "pupload.controller"); err != nil {
		log.Error("error initalizing telemetry", "err", err)
	}
	defer telemetry.Shutdown(context.Background())

	// SyncPlane
	s, err := syncplane.CreateControllerSyncLayer(cfg.SyncPlane)
	if err != nil {
		log.Error("error intalizing sync plane", "err", err)
		return err
	}
	defer s.Close()

	// Services

	f, err := flow.CreateFlowService(cfg, s)
	if err != nil {
		return err
	}
	defer f.Close(ctx)

	// Handlers

	handler := controllerserver.NewServer(*cfg, f)
	srv := &http.Server{
		Addr:    ":1234",
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("server forced to shutdown", "err", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log.Info("Controller shutting down...")
	return srv.Shutdown(shutdownCtx)
}

func RunWithConfigSilent(ctx context.Context, cfg *config.ControllerSettings) error {

	log.SetOutput(io.Discard)
	logging.Init(logging.Config{
		AppName: "pupload.controller",
		Level:   slog.LevelInfo,
		Format:  "text",
		Out:     io.Discard,
	})
	log := logging.Root()

	if err := telemetry.Init(cfg.Telemetry, "pupload.controller"); err != nil {
		log.Error("error initalizing telemetry", "err", err)
	}
	defer telemetry.Shutdown(context.Background())

	// SyncPlane
	s, err := syncplane.CreateControllerSyncLayer(cfg.SyncPlane)
	if err != nil {
		log.Error("error intalizing sync plane", "err", err)
		return err
	}
	defer s.Close()

	// Services

	f, err := flow.CreateFlowService(cfg, s)
	if err != nil {
		return err
	}
	defer f.Close(ctx)

	// Handlers

	handler := controllerserver.NewServer(*cfg, f)
	srv := &http.Server{
		Addr:    ":1234",
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("server forced to shutdown", "err", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log.Info("Controller shutting down...")
	return srv.Shutdown(shutdownCtx)

}

func Run() error {
	cfg := config.DefaultConfig()
	ctx := context.Background()
	return RunWithConfig(ctx, cfg)
}
