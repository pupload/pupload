package service

import (
	"context"
	"log/slog"

	"github.com/pupload/pupload/internal/controller/config"
	"github.com/pupload/pupload/internal/controller/flows/repo"
	"github.com/pupload/pupload/internal/controller/flows/runtime"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
	"github.com/pupload/pupload/internal/validation"

	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/models"

	"go.opentelemetry.io/otel/attribute"
)

type FlowService struct {
	projectRepo repo.ProjectRepo
	runtimeRepo repo.RuntimeRepo

	syncLayer syncplane.SyncLayer

	log *slog.Logger
}

func CreateFlowService(cfg *config.ControllerSettings, s syncplane.SyncLayer) (*FlowService, error) {
	slog := logging.ForService("flow")

	runtimeRepo, err := repo.CreateRuntimeRepo(cfg.RuntimeRepo)
	if err != nil {
		return nil, err
	}

	projectRepo, err := repo.CreateProjectRepo(cfg.ProjectRepo)
	if err != nil {
		return nil, err
	}

	f := FlowService{
		projectRepo: projectRepo,
		runtimeRepo: runtimeRepo,

		syncLayer: s,

		log: slog,
	}

	s.RegisterFlowStepHandler(f.FlowStepHandler)
	s.RegisterNodeFinishedHandler(f.NodeFinishedHandler)
	s.RegisterNodeFailedHandler(f.NodeFailedHandler)

	s.Start()

	return &f, nil
}

func (f *FlowService) Close(ctx context.Context) {
	f.projectRepo.Close(ctx)
	f.runtimeRepo.Close(ctx)
}

func (f *FlowService) RunFlow(flow models.Flow, nodeDefs []models.NodeDef) (models.FlowRun, error) {

	ctx, span := telemetry.Tracer("pupload.controller").Start(context.Background(), "RunFlow")
	defer span.End()

	flow.Normalize()
	for i := range nodeDefs {
		nodeDefs[i].Normalize()
		f.log.Info("node def tier", "node_def", nodeDefs[i].Name, "tier", nodeDefs[i].Tier)
	}

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
