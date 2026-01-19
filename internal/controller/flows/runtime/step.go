package runtime

import (
	"context"

	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
)

func (rt *RuntimeFlow) Step(s syncplane.SyncLayer) {
	ctx := telemetry.ExtractContext(context.Background(), rt.TraceParent)
	ctx, span := telemetry.Tracer("pupload.controller").Start(ctx, "Step")

	defer span.End()

	for {
		rt.log.Info("stepFlow state", "runID", rt.FlowRun.ID, "state", rt.FlowRun.Status)

		if rt.IsComplete() {
			rt.FlowRun.Status = models.FLOWRUN_COMPLETE
			return
		}

		switch rt.FlowRun.Status {
		case models.FLOWRUN_STOPPED:

		case models.FLOWRUN_WAITING:

			rt.updateWaiting()
			rt.updateAllNodes()

			if len(rt.nodesReady()) == 0 {
				return
			}

			rt.FlowRun.Status = models.FLOWRUN_RUNNING

		case models.FLOWRUN_RUNNING:
			for _, nodeID := range rt.nodesReady() {
				if err := rt.handleExecuteNode(ctx, nodeID, s); err != nil {
					rt.log.Error("error executing node", "err", err, "node_id", nodeID)
					rt.FlowRun.Status = models.FLOWRUN_ERROR
					return
				}

				rt.FlowRun.NodeState[nodeID] = models.NodeState{Status: models.NODERUN_RUNNING, Logs: rt.FlowRun.NodeState[nodeID].Logs}
			}

			rt.FlowRun.Status = models.FLOWRUN_WAITING

		case models.FLOWRUN_COMPLETE:
			return

		case models.FLOWRUN_ERROR:
			return

		}

	}
}

func (rt *RuntimeFlow) IsError() bool {
	return rt.FlowRun.Status == models.FLOWRUN_ERROR
}

func (rt *RuntimeFlow) IsComplete() bool {
	return rt.FlowRun.Status == models.FLOWRUN_COMPLETE || len(rt.nodesLeft()) == 0
}

func (rt *RuntimeFlow) nodesLeft() []string {
	left := make([]string, 0)
	for nodeID, state := range rt.FlowRun.NodeState {
		if state.Status != models.NODERUN_COMPLETE {
			left = append(left, nodeID)
		}
	}

	return left
}

func (rt *RuntimeFlow) nodesReady() []string {
	ready := make([]string, 0)
	for nodeID, state := range rt.FlowRun.NodeState {
		if state.Status == models.NODERUN_READY {
			ready = append(ready, nodeID)
		}
	}

	return ready
}

func (rt *RuntimeFlow) updateAllNodes() {
	for id := range rt.nodes {
		rt.shouldNodeReady(id)
	}

}

type WaitingURLResult int

const (
	WaitNoChange   WaitingURLResult = iota
	WaitReady                       // Object Exists
	WaitURLExpired                  // URL Expired
	WaitFailed                      // Non retryable error
)

func (rt *RuntimeFlow) updateWaiting() {

	for i, url := range rt.FlowRun.WaitingURLs {
		result := rt.checkWaitingURL(url)
		switch result {
		case WaitNoChange:

		case WaitReady:
			rt.FlowRun.WaitingURLs = append(rt.FlowRun.WaitingURLs[:i], rt.FlowRun.WaitingURLs[i+1:]...)
			rt.FlowRun.Artifacts[url.Artifact.EdgeName] = url.Artifact

		case WaitURLExpired:

		case WaitFailed:
			return
		}
	}
}

func (rt *RuntimeFlow) checkWaitingURL(w models.WaitingURL) WaitingURLResult {

	// if time.Now().After(w.TTL) {
	// 	return WaitURLExpired
	// }

	store, ok := rt.stores[w.Artifact.StoreName]
	if !ok {
		return WaitFailed
	}
	exists := store.Exists(w.Artifact.ObjectName)
	if exists {
		return WaitReady
	}

	return WaitNoChange
}
