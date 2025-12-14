package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pupload/internal/controller/flows/repo"
	"pupload/internal/models"

	"github.com/go-redsync/redsync/v4"
	"github.com/hibiken/asynq"
)

type FlowRunSchedulerConfigProvider struct {
	runtimeRepo repo.RuntimeRepo
	Cronspec    string
}

func CreateFlowRunScheduleProvider(repo repo.RuntimeRepo, cronspec string) FlowRunSchedulerConfigProvider {
	return FlowRunSchedulerConfigProvider{
		runtimeRepo: repo,
		Cronspec:    cronspec,
	}
}

func (f *FlowService) AsynqConfigureHandlers() *asynq.ServeMux {
	mux := asynq.NewServeMux()

	mux.HandleFunc(models.TypeFlowStep, f.HandleFlowStepTask)
	mux.HandleFunc(models.TypeNodeFinished, f.HandleNodeFinishedTask)

	return mux
}

func (p *FlowRunSchedulerConfigProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	var configs []*asynq.PeriodicTaskConfig

	ids, err := p.runtimeRepo.ListRuntimeIDs()
	if err != nil {
		return nil, err
	}

	for _, id := range ids {

		task, err := NewFlowStepTask(id)
		if err != nil {
			continue
		}

		configs = append(configs, &asynq.PeriodicTaskConfig{Task: task, Cronspec: p.Cronspec, Opts: []asynq.Option{asynq.Queue("controller")}})
	}

	return configs, nil
}

func (f *FlowService) HandleFlowStepTask(ctx context.Context, t *asynq.Task) error {
	var p models.FlowStepPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Printf("Error unmarshaling in Flow ID: %s\n", p.RunID)
		return err
	}

	key := fmt.Sprintf("runtimelock:%s", p.RunID)
	m := f.rs.NewMutex(key, redsync.WithTries(1))
	err := m.LockContext(ctx)
	if err != nil {
		f.log.Error("runtime lock already in use", "run_id", p.RunID)
		return err
	}
	defer m.UnlockContext(ctx)

	runtime, err := f.runtimeRepo.LoadRuntime(p.RunID)
	if err != nil {
		f.log.Error("unable to get runtime flow from runtimeRepo", "runID", p.RunID)
	}

	runtime.RebuildRuntimeFlow()
	runtime.Step(f.asynqClient)
	if runtime.IsComplete() {
		f.HandleFlowComplete(p.RunID)
		return nil
	}

	f.runtimeRepo.SaveRuntime(runtime)

	return nil
}

func NewFlowStepTask(runID string) (*asynq.Task, error) {
	payload, err := json.Marshal(models.FlowStepPayload{
		RunID: runID,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(models.TypeFlowStep, payload, asynq.TaskID(runID)), nil
}

func (f *FlowService) HandleNodeFinishedTask(ctx context.Context, t *asynq.Task) error {

	var p models.NodeFinishedPayload

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		f.log.Error("HandleNodeFinishedTask: error processing node finished", "err", err)
		return err
	}

	f.log.Info("HandleNodeFinishedTask: starting node finished task", "run_id", p.RunID)

	key := fmt.Sprintf("runtimelock:%s", p.RunID)
	m := f.rs.NewMutex(key, redsync.WithTries(1))
	err := m.LockContext(ctx)
	if err != nil {
		f.log.Error("HandleNodeFinishedTask: runtime lock already in use", "run_id", p.RunID, "err", err)
		return err
	}
	defer m.UnlockContext(ctx)

	runtime, err := f.runtimeRepo.LoadRuntime(p.RunID)
	if err != nil {
		f.log.Error("HandleNodeFinishedTask: error loading runtime", "run_id", p.RunID, "err", err)
		return err
	}

	runtime.RebuildRuntimeFlow()

	if err := runtime.HandleNodeFinished(p.NodeID, p.Logs); err != nil {
		f.log.Error("HandleNodeFinishedTask: error handling node finished", "run_id", p.RunID, "node_id", p.NodeID, "err", err)
		return err
	}

	runtime.Step(f.asynqClient)
	if err := f.runtimeRepo.SaveRuntime(runtime); err != nil {
		f.log.Error("HandleNodeFinishedTask: error saving runtime", "run_id", p.RunID, "node_id", p.NodeID, "err", err)

		return err
	}

	return nil
}
