package server

import (
	"log"
	"pupload/internal/models"
	"pupload/internal/worker/node"

	"github.com/hibiken/asynq"
)

func CreateWorkerServer() {

	srv := asynq.NewServer(asynq.RedisClientOpt{
		Addr: "localhost:6379",
	}, asynq.Config{
		Concurrency: 1,
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(models.TypeNodeExecute, node.HandleNodeExecuteTask)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Error starting worker: %s", err.Error())
	}

}
