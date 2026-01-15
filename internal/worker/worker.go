package worker

import (
	"context"
	"io"
	"log"
	"log/slog"

	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
	"github.com/pupload/pupload/internal/worker/config"
	"github.com/pupload/pupload/internal/worker/container"
	"github.com/pupload/pupload/internal/worker/server"
)

func Run() error {
	cfg := config.DefaultConfig()
	ctx := context.Background()

	return RunWithConfig(ctx, cfg)
}

func RunWithConfig(ctx context.Context, cfg *config.WorkerConfig) error {

	logging.Init(logging.Config{
		AppName: "worker",
		Level:   slog.LevelInfo,
		Format:  "text",
	})

	log := logging.Root()

	telemetry.Init(cfg.Telemetry, "pupload.worker")

	s, err := syncplane.CreateWorkerSyncLayer(cfg.SyncPlane, cfg.Resources)
	if err != nil {
		return err
	}

	cs := container.CreateContainerService()
	server.NewWorkerServer(s, &cs)

	<-ctx.Done()

	log.Info("Worker shutting down...")

	return s.Close()

}

func RunWithConfigSilent(ctx context.Context, cfg *config.WorkerConfig) error {
	log.SetOutput(io.Discard)
	logging.Init(logging.Config{
		AppName: "worker",
		Level:   slog.LevelInfo,
		Format:  "json",
		Out:     io.Discard,
	})

	log := logging.Root()

	telemetry.Init(cfg.Telemetry, "pupload.worker")

	s, err := syncplane.CreateWorkerSyncLayer(cfg.SyncPlane, cfg.Resources)
	if err != nil {
		return err
	}

	cs := container.CreateContainerService()
	server.NewWorkerServer(s, &cs)

	<-ctx.Done()

	log.Info("Worker shutting down...")

	return s.Close()

}
