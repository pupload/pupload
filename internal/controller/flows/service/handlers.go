package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pupload/pupload/internal/syncplane"
)

func (f *FlowService) FlowStepHandler(ctx context.Context, payload syncplane.FlowStepPayload) error {
	m := f.syncLayer.NewMutex(payload.RunID, 10*time.Second)
	err := m.Lock(ctx)

	if err != nil {
		f.log.Error("runtime lock already in use", "run_id", payload.RunID)
		return err
	}
	defer m.Unlock(ctx)

	runtime, err := f.runtimeRepo.LoadRuntime(payload.RunID)
	if err != nil {
		f.log.Error("unable to get runtime flow from runtimeRepo", "runID", payload.RunID)
	}

	runtime.RebuildRuntimeFlow()
	runtime.Step(f.syncLayer)
	if runtime.IsComplete() || runtime.IsError() {
		f.HandleFlowComplete(payload.RunID)
		return nil
	}

	f.runtimeRepo.SaveRuntime(runtime)

	return nil
}

func (f *FlowService) NodeFinishedHandler(ctx context.Context, payload syncplane.NodeFinishedPayload) error {
	f.log.Info("HandleNodeFinishedTask: starting node finished task", "run_id", payload.RunID)

	key := fmt.Sprintf("runtimelock:%s", payload.RunID)
	m := f.syncLayer.NewMutex(key, 10*time.Second)
	err := m.Lock(ctx)
	if err != nil {
		f.log.Error("HandleNodeFinishedTask: runtime lock already in use", "run_id", payload.RunID, "err", err)
		return err
	}
	defer m.Unlock(ctx)

	runtime, err := f.runtimeRepo.LoadRuntime(payload.RunID)
	if err != nil {
		f.log.Error("HandleNodeFinishedTask: error loading runtime", "run_id", payload.RunID, "err", err)
		return err
	}

	runtime.RebuildRuntimeFlow()

	if err := runtime.HandleNodeFinished(payload.NodeID, payload.Logs); err != nil {
		f.log.Error("HandleNodeFinishedTask: error handling node finished", "run_id", payload.RunID, "node_id", payload.NodeID, "err", err)
		return err
	}

	runtime.Step(f.syncLayer)
	if err := f.runtimeRepo.SaveRuntime(runtime); err != nil {
		f.log.Error("HandleNodeFinishedTask: error saving runtime", "run_id", payload.RunID, "node_id", payload.NodeID, "err", err)
		return err
	}

	return nil
}
func (f *FlowService) NodeFailedHandler(ctx context.Context, payload syncplane.NodeFailedPayload) error {

	isFinalFailure := payload.Attempt >= payload.MaxAttempts
	key := fmt.Sprintf("runtimelock:%s", payload.RunID)
	m := f.syncLayer.NewMutex(key, 10*time.Second)
	err := m.Lock(ctx)
	if err != nil {
		f.log.Error("HandleNodeFinishedTask: runtime lock already in use", "run_id", payload.RunID, "err", err)
		return err
	}
	defer m.Unlock(ctx)

	runtime, err := f.runtimeRepo.LoadRuntime(payload.RunID)
	if err != nil {
		f.log.Error("HandleNodeFinishedTask: error loading runtime", "run_id", payload.RunID, "err", err)
		return err
	}

	runtime.RebuildRuntimeFlow()

	if err := runtime.HandleNodeFailed(payload.NodeID, payload.Logs, payload.Error, payload.Attempt, payload.MaxAttempts, isFinalFailure); err != nil {
		f.log.Error("HandleNodeFinishedTask: error handling node failed", "run_id", payload.RunID, "node_id", payload.NodeID, "err", err)
		return err
	}

	runtime.Step(f.syncLayer)
	if err := f.runtimeRepo.SaveRuntime(runtime); err != nil {
		f.log.Error("HandleNodeFinishedTask: error saving runtime", "run_id", payload.RunID, "node_id", payload.NodeID, "err", err)
		return err
	}

	return nil
}
