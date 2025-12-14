package service

import (
	"context"
	"log/slog"
	"pupload/internal/controller/flows/repo"
	runtime_repo "pupload/internal/controller/flows/repo/runtime"
	"pupload/internal/controller/flows/runtime"

	"pupload/internal/controller/scheduler"
	"pupload/internal/logging"
	"pupload/internal/models"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type FlowService struct {
	projectRepo repo.ProjectRepo
	runtimeRepo repo.RuntimeRepo

	redisClient *redis.Client
	asynqClient *asynq.Client
	asynqServer *asynq.Server
	rs          *redsync.Redsync

	scheduler *scheduler.Scheduler
	log       *slog.Logger
}

func CreateFlowService(rdb *redis.Client) *FlowService {

	ctx := context.Background()
	slog := logging.ForService("flow")

	asynqClient := asynq.NewClientFromRedisClient(rdb)
	asynqServer := asynq.NewServerFromRedisClient(rdb, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"controller": 1,
		},
	})

	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)

	runtimeStore := runtime_repo.CreateRedisRuntimeRepo(rdb)

	flowRunScheduleProvider := CreateFlowRunScheduleProvider(runtimeStore, "@every 10s")
	flowRunScheduler, err := scheduler.NewScheduler(&flowRunScheduleProvider, rdb)
	if err != nil {
		slog.Error("could not create flow run scheduler", "err", err)
		panic("")
	}

	f := FlowService{

		redisClient: rdb,
		asynqClient: asynqClient,
		asynqServer: asynqServer,
		rs:          rs,

		runtimeRepo: runtimeStore,

		log:       slog,
		scheduler: flowRunScheduler,
	}

	go func() {
		asynqServer.Start(f.AsynqConfigureHandlers())
	}()

	go func() {
		flowRunScheduler.Start(ctx)
	}()

	f.log.Info("test")

	return &f
}

func (f *FlowService) Close() {
	f.asynqServer.Stop()
	f.asynqClient.Close()
	f.redisClient.Close()
	f.scheduler.Close()
}

func (f *FlowService) RunFlow(flow models.Flow, nodeDefs []models.NodeDef) (models.FlowRun, error) {
	err := f.ValidateFlow(&flow)
	if err != nil {
		f.log.Error("unable to validate flow", "err", err)
		return models.FlowRun{}, err
	}

	runtime, err := runtime.CreateRuntimeFlow(flow, nodeDefs)
	if err != nil {
		return models.FlowRun{}, err
	}

	runtime.Start(f.asynqClient)
	f.runtimeRepo.SaveRuntime(runtime)
	return runtime.FlowRun, nil
}

func (f *FlowService) Status(runID string) (models.FlowRun, error) {
	runtime, err := f.runtimeRepo.LoadRuntime(runID)
	if err != nil {
		return models.FlowRun{}, err
	}

	return runtime.FlowRun, nil
}

func (f *FlowService) HandleFlowComplete(runID string) error {
	f.runtimeRepo.DeleteRuntime(runID)
	return nil
}
