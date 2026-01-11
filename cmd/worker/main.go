package main

import (
	"log/slog"
	"os"
	"os/signal"
	"pupload/internal/logging"
	"pupload/internal/syncplane"
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
