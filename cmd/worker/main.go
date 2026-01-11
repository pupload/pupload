package main

import (
	"log/slog"
	"os"
	"os/signal"
	"pupload/internal/logging"
	"pupload/internal/syncplane"
	"pupload/internal/telemetry"
	"pupload/internal/worker/config"
	"pupload/internal/worker/container"
	"pupload/internal/worker/server"
	"syscall"
)

func main() {

	cfg := config.DefaultConfig()

	logging.Init(logging.Config{
		AppName: "worker",
		Level:   slog.LevelInfo,
		Format:  "json",
	})

	test_tel_config := telemetry.TelemetrySettings{
		Enabled:    true,
		Exporter:   telemetry.ExporterOTLP,
		Endpoint:   "localhost:4317",
		Insecure:   true,
		SampleRate: 1.0,
	}

	telemetry.Init(test_tel_config, "pupload.worker")

	s, err := syncplane.CreateWorkerSyncLayer(cfg.SyncPlane, cfg.Resources)
	if err != nil {
		return
	}

	cs := container.CreateContainerService()
	server.NewWorkerServer(s, &cs)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	sig := <-quit
	_ = sig

	if err := s.Close(); err != nil {
		os.Exit(1)
	}

}
