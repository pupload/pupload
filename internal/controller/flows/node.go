package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pupload/internal/models"
	"pupload/internal/util"
	"time"

	"github.com/hibiken/asynq"
)

func (f *FlowService) GetNode(flowName string, nodeIndex int) models.Node {
	return f.FlowList[flowName].Nodes[nodeIndex]
}

func (f *FlowService) HandleExecuteNode(run *FlowRun, nodeIndex int) {

	node := f.GetNode(run.FlowName, nodeIndex)
	flow, _ := f.GetFlow(run.FlowName)

	nodeDef, exists := f.NodeDefs[node.DefName]
	if !exists {
		log.Fatalf("Attempting to run node that does not have definition")
	}

	inputs := make(map[string]string)
	for _, edge := range node.Inputs {
		artifact := run.Artifacts[edge.Edge]

		store, _ := f.GetStore(run.FlowName, artifact.StoreName)
		url, err := store.GetURL(context.TODO(), artifact.ObjectName, 1*time.Hour)
		if err != nil {
			fmt.Println(err.Error())
		}

		inputs[edge.Name] = url.String()
	}

	outputs := make(map[string]string)
	for _, edge := range node.Outputs {

		artifact := Artifact{
			StoreName:  *flow.DefaultStore,
			ObjectName: fmt.Sprintf("%s-%s", edge.Edge, run.ID),
			EdgeName:   edge.Edge,
		}

		if flow.DefaultStore == nil {
			f.log.Error("default flow store is nil")
		}

		store, _ := f.GetStore(run.FlowName, *flow.DefaultStore)

		url, err := store.PutURL(context.TODO(), artifact.ObjectName, 10*time.Second)
		if err != nil {
			f.log.Error("could not generate put url", "err", err)
		}

		outputs[edge.Name] = url.String()
		WaitingURL := WaitingURL{
			Artifact: artifact,
			PutURL:   url.String(),
			TTL:      time.Now().Add(10 * time.Second),
		}

		run.WaitingURLs = append(run.WaitingURLs, WaitingURL)
	}

	p := models.NodeExecutePayload{
		RunID:      run.ID,
		NodeDef:    nodeDef,
		Node:       node,
		InputURLs:  inputs,
		OutputURLs: outputs,
	}

	info, err := f.executeNode(p)
	if err != nil {
		f.log.Error("error enqueing node to execute", "err", err, "node", nodeIndex)
	}

	run.RunningNodes[nodeIndex] = info

}

func (f *FlowService) executeNode(payload models.NodeExecutePayload) (*asynq.TaskInfo, error) {

	task, err := NewNodeExecuteTask(payload)

	if err != nil {
		return nil, err
	}
	taskinfo, _ := f.AsynqClient.Enqueue(task, asynq.Queue("worker"))

	return taskinfo, nil
}

func NewNodeExecuteTask(p models.NodeExecutePayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(models.TypeNodeExecute, payload), nil
}

func (f *FlowService) HandleNodeFinishedTask(ctx context.Context, t *asynq.Task) error {
	var p models.NodeFinishedPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		f.log.Error("error processing node finished")
		return err
	}

	mutex_key := fmt.Sprintf("flowrunlock:%s", p.RunID)
	ok, err := util.AcquireLock(f.RedisClient, mutex_key, 10*time.Second)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("flowrun lock %s currently being held", p.RunID)
	}

	return nil
}

func HandleNodeErrorTask(ctx context.Context, t *asynq.Task) error {
	return nil
}
