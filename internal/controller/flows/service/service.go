package service

import (
	"context"
	"log/slog"
	"pupload/internal/controller/flows/repo"
	runtime_repo "pupload/internal/controller/flows/repo/runtime"
	"pupload/internal/controller/flows/runtime"
	"pupload/internal/syncplane"
	"pupload/internal/telemetry"
	"pupload/internal/validation"

	"pupload/internal/logging"
	"pupload/internal/models"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

type FlowService struct {
	projectRepo repo.ProjectRepo
	runtimeRepo repo.RuntimeRepo

	syncLayer syncplane.SyncLayer

	log *slog.Logger
}

func CreateFlowService(rdb *redis.Client, s syncplane.SyncLayer) *FlowService {
	slog := logging.ForService("flow")

	runtimeStore := runtime_repo.CreateRedisRuntimeRepo(rdb)

	f := FlowService{
		runtimeRepo: runtimeStore,

		syncLayer: s,

		log: slog,
	}

	s.RegisterFlowStepHandler(f.FlowStepHandler)
	s.RegisterNodeFinishedHandler(f.NodeFinishedHandler)
	s.RegisterNodeErrorHandler(f.NodeErrorHandler)

	s.Start()

	return &f
}

func (f *FlowService) Close() {

}

func (f *FlowService) RunFlow(flow models.Flow, nodeDefs []models.NodeDef) (models.FlowRun, error) {

	ctx, span := telemetry.Tracer("pupload.controller").Start(context.Background(), "RunFlow")
	defer span.End()

	flow.Normalize()

	if err := validation.ValidateFlow(flow, nodeDefs); err != nil {
		f.log.Error("unable to validate flow", "err", err)
		return models.FlowRun{}, err
	}

	runtime, err := runtime.CreateRuntimeFlow(ctx, flow, nodeDefs)
	if err != nil {
		return models.FlowRun{}, err
	}

	span.SetAttributes(attribute.String("run_id", runtime.FlowRun.ID))

	runtime.Start(f.syncLayer)
	f.runtimeRepo.SaveRuntime(runtime)
	f.syncLayer.AddRunToScheduler(runtime.FlowRun.ID)

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
	f.syncLayer.RemoveRunFromScheduler(runID)
	return nil
}
