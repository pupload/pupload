package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"pupload/internal/models"
	"time"

	"github.com/hibiken/asynq"
)

func (rt *RuntimeFlow) handleExecuteNode(nodeID string, asynqClient *asynq.Client) error {
	node := rt.nodes[nodeID]
	inputs := make(map[string]string)

	for _, edge := range node.Inputs {
		artifact := rt.FlowRun.Artifacts[edge.Edge]
		store := rt.stores[artifact.StoreName]

		url, err := store.GetURL(context.TODO(), artifact.ObjectName, 1*time.Hour)
		if err != nil {
			rt.log.Error("unable to generate store get url", "message", err)
			return err
		}

		inputs[edge.Name] = url.String()
	}

	outputs := make(map[string]string)
	for _, edge := range node.Outputs {

		artifact, err := rt.makeOutputArtifact(edge)
		if err != nil {
			return err
		}

		store, ok := rt.stores[artifact.StoreName]
		if !ok {
			rt.log.Error("unable to acquire store", "store_name", artifact.StoreName)
			return fmt.Errorf("unable to acquire store described in artifact")
		}

		url, err := store.PutURL(context.TODO(), artifact.ObjectName, 10*time.Second)
		if err != nil {
			rt.log.Error("could not generate put url", "err", err)
			return err
		}

		outputs[edge.Name] = url.String()
		WaitingURL := models.WaitingURL{
			Artifact: *artifact,
			PutURL:   url.String(),
			TTL:      time.Now().Add(10 * time.Second),
		}

		rt.FlowRun.WaitingURLs = append(rt.FlowRun.WaitingURLs, WaitingURL)
	}

	err := node.executeNode(asynqClient, rt.FlowRun.ID, inputs, outputs)
	if err != nil {
		return nil
	}

	return nil
}

func (rt *RuntimeFlow) makeOutputArtifact(edge models.NodeEdge) (*models.Artifact, error) {
	for _, well := range rt.Flow.DataWells {
		if well.Edge != edge.Edge {
			continue
		}

		artifact := models.Artifact{
			StoreName:  well.Store,
			EdgeName:   well.Edge,
			ObjectName: rt.processDatawellKey(well),
		}

		return &artifact, nil
	}

	if rt.Flow.DefaultStore == nil {
		return nil, fmt.Errorf("default store is nil")
	}

	artifact := models.Artifact{
		StoreName:  *rt.Flow.DefaultStore,
		ObjectName: fmt.Sprintf("%s-%s", edge.Edge, rt.FlowRun.ID),
		EdgeName:   edge.Edge,
	}

	return &artifact, nil

}

func (rt *RuntimeFlow) HandleNodeFinished(nodeID string, logs []models.LogRecord) error {
	_, ok := rt.nodes[nodeID]
	if !ok {
		return fmt.Errorf("node does not exist")
	}

	curr_state := rt.FlowRun.NodeState[nodeID]
	new_logs := append(curr_state.Logs, logs...)
	rt.FlowRun.NodeState[nodeID] = models.NodeState{Status: models.NODERUN_COMPLETE, Logs: new_logs}

	return nil
}

func (rn *RuntimeNode) executeNode(asynqClient *asynq.Client, runID string, input, output map[string]string) error {
	payload := models.NodeExecutePayload{
		RunID:      runID,
		Node:       *rn.Node,
		NodeDef:    rn.NodeDef,
		InputURLs:  input,
		OutputURLs: output,
	}

	task, err := NewNodeExecuteTask(payload)
	if err != nil {
		return err
	}

	_, err = asynqClient.Enqueue(task, asynq.Queue("worker"))
	return err
}

func (rt *RuntimeFlow) shouldNodeReady(nodeID string) {
	node := rt.nodes[nodeID]
	curr_state := rt.FlowRun.NodeState[nodeID].Status

	if curr_state != models.NODERUN_IDLE {
		return
	}

	for _, input := range node.Inputs {
		_, ok := rt.FlowRun.Artifacts[input.Edge]
		if !ok {
			return
		}
	}

	rt.FlowRun.NodeState[nodeID] = models.NodeState{Status: models.NODERUN_READY, Logs: rt.FlowRun.NodeState[nodeID].Logs}
}

func NewNodeExecuteTask(p models.NodeExecutePayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(models.TypeNodeExecute, payload), nil
}
