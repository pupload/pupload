package flows

import (
	"context"
	"fmt"
	"log"
	"pupload/internal/models"
	"time"
)

func (f *FlowService) ListFlows() map[string]models.Flow {
	return f.FlowList
}

func (f *FlowService) GetFlow(name string) (models.Flow, error) {

	flow, ok := f.FlowList[name]
	if !ok {
		return flow, fmt.Errorf("pooo %s does not exist", name)
	}

	return flow, nil
}

func (f *FlowService) StartFlow(ctx context.Context, name string) (string, error) {
	flow, exists := f.FlowList[name]
	if !exists {
		return "", fmt.Errorf("Flow %s does not exist", name)
	}

	if len(flow.Nodes) == 0 {
		return "", fmt.Errorf("Flow %s does not contain any nodes", name)
	}

	flowRun, err := f.CreateFlowRun(name)

	if err != nil {
		return "", err
	}

	err = f.HandleStepFlow(ctx, flowRun.ID)

	if err != nil {
		return "", err
	}

	return flowRun.ID, nil
}

func (f *FlowService) HandleStepFlow(ctx context.Context, id string) error {

	run, err := f.GetFlowRun(id)
	if err != nil {
		return fmt.Errorf("Can't retrieve flow status %s. Is Redis running? %s", id, err)
	}

	if len(run.NodesLeft) == 0 {
		run.Status = FLOWRUN_COMPLETE
		f.updateFlowRun(run)
		return nil
	}

	nodesToRun := f.nodesAvailableToRun(run)

	switch run.Status {
	case FLOWRUN_IDLE:

		f.initalizeWaitingURLs(&run)

		if len(nodesToRun) == 0 {
			run.Status = FLOWRUN_WAITING
			f.updateFlowRun(run)
			return f.HandleStepFlow(ctx, id)
		}

		run.Status = FLOWRUN_RUNNING
		f.updateFlowRun(run)
		return f.HandleStepFlow(ctx, id)

	case FLOWRUN_WAITING:

		updated, err := f.checkWaitingUrls(&run)
		if err != nil {
			log.Fatalf("Error in flow step function: %s", err)
		}

		if updated {
			run.Status = FLOWRUN_RUNNING
			f.updateFlowRun(run)
			return f.HandleStepFlow(ctx, id)
		}

		if err := f.EnqueueFlowStepTask(id, time.Second*3); err != nil {
			log.Printf("Error Enqueing Step: %s", err)
		}

		return nil

	case FLOWRUN_RUNNING:

		for _, nodeIndex := range nodesToRun {
			f.HandleExecuteNode(run, nodeIndex)
		}

		run.Status = FLOWRUN_WAITING
		f.updateFlowRun(run)
		return f.HandleStepFlow(ctx, id)
	}

	return nil
}

func (f *FlowService) GetStore(flowName string, storeName string) (store models.Store, ok bool) {

	// prefer local store for the given flow, fall back to global store
	stores, ok := f.LocalStoreMap[LocalStoreKey{flowName, storeName}]
	if !ok {
		stores, ok = f.GlobalStoreMap[storeName]
	}

	return stores, ok
}

func (f *FlowService) NodeLength(flowName string) int {
	return len(f.FlowList[flowName].Nodes)
}

func (f *FlowService) nodesAvailableToRun(flowRun FlowRun) []int {

	nodes := make([]int, 0)

	for i := range flowRun.NodesLeft {

		runnable := true

		node := f.FlowList[flowRun.FlowName].Nodes[i]
		for _, input := range node.Inputs {

			_, ok := flowRun.Artifacts[input.Edge]
			if !ok {
				runnable = false
				break
			}
		}

		if runnable {
			nodes = append(nodes, i)
		}
	}

	return nodes
}
