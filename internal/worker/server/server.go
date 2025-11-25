package server

import (
	"log"
	"pupload/internal/models"
	"pupload/internal/worker/container"
	"pupload/internal/worker/node"

	"github.com/hibiken/asynq"
)

func NewWorkerServer() {

	srv := asynq.NewServer(asynq.RedisClientOpt{
		Addr: "localhost:6379",
	}, asynq.Config{
		Concurrency: 1,
		Queues: map[string]int{
			"worker": 1,
		},
	})

	ds := container.CreateContainerService()
	ns := node.CreateNodeService(&ds)

	mux := asynq.NewServeMux()
	mux.HandleFunc(models.TypeNodeExecute, ns.HandleNodeExecuteTask)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Error starting worker: %s", err.Error())
	}

}
