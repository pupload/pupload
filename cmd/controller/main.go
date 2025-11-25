package main

import (
	"log/slog"
	"net/http"
	config "pupload/internal/controller/config"
	"pupload/internal/controller/flows"
	controllerserver "pupload/internal/controller/server"
	"pupload/internal/logging"

	"github.com/redis/go-redis/v9"
)

func main() {

	config := config.LoadControllerConfig("/home/seb/Projects/OpenUpload/test_config")

	logging.Init(logging.Config{
		AppName: "controller",
		Level:   slog.LevelInfo,
		Format:  "text",
	})

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	f := flows.CreateFlowService(config.Storage.DataPath, rdb)
	defer f.Close()

	srv := &http.Server{
		Addr:    ":1234",
		Handler: controllerserver.NewServer(config, f),
	}

	srv.ListenAndServe()
}
