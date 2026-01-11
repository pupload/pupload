package main

import (
	"context"
	"log/slog"
	"net/http"
	config "pupload/internal/controller/config"
	flows "pupload/internal/controller/flows/service"
	controllerserver "pupload/internal/controller/server"
	"pupload/internal/logging"
	"pupload/internal/syncplane"
	"pupload/internal/telemetry"

	"github.com/redis/go-redis/v9"
)

func main() {

	config := config.LoadControllerConfig("/home/seb/Projects/pupload/config")

	logging.Init(logging.Config{
		AppName: "controller",
		Level:   slog.LevelInfo,
		Format:  "text",
	})

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	test_sync_config := syncplane.SyncPlaneSettings{
		SelectedSyncPlane: "redis",

		Redis: syncplane.RedisSettings{
			Address: "localhost:6379",
		},

		ControllerStepInterval: "@every 10s",
	}

	test_tel_config := telemetry.TelemetrySettings{
		Enabled:    true,
		Exporter:   telemetry.ExporterOTLP,
		Endpoint:   "localhost:4317",
		Insecure:   true,
		SampleRate: 1.0,
	}

	if err := telemetry.Init(test_tel_config, "pupload.controller"); err != nil {

	}
	defer telemetry.Shutdown(context.Background())

	s, err := syncplane.CreateControllerSyncLayer(test_sync_config)
	if err != nil {

	}
	defer s.Close()

	f := flows.CreateFlowService(rdb, s)
	defer f.Close()

	srv := &http.Server{
		Addr:    ":1234",
		Handler: controllerserver.NewServer(config, f),
	}

	srv.ListenAndServe()
}
